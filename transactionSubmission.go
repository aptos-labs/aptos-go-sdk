package aptos

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aptos-labs/aptos-go-sdk/api"
)

// TransactionSubmissionType is the counter for an enum
type TransactionSubmissionType uint8

const (
	// TransactionSubmissionTypeSingle represents a single signer transaction, no multi-agent and no-fee payer
	TransactionSubmissionTypeSingle TransactionSubmissionType = iota
	// TransactionSubmissionTypeMultiAgent represents a multi-agent or fee payer transaction
	TransactionSubmissionTypeMultiAgent TransactionSubmissionType = iota
	defaultWorkerCount                  uint32                    = 20
)

type TransactionBuildPayload struct {
	Id      uint64
	Type    TransactionSubmissionType
	Inner   TransactionPayload // The actual transaction payload
	Options []any              // This is a placeholder to allow future optional arguments
}

type TransactionBuildResponse struct {
	Id       uint64
	Response RawTransactionImpl
	Err      error
}
type TransactionSubmissionRequest struct {
	Id        uint64
	SignedTxn *SignedTransaction
}

type TransactionSubmissionResponse struct {
	Id       uint64
	Response *api.SubmitTransactionResponse
	Err      error
}

// WorkerPoolConfig contains configuration for the transaction processing worker pool
type WorkerPoolConfig struct {
	NumWorkers uint32
	// Channel buffer sizes. If 0, defaults to NumWorkers
	BuildResponseBuffer uint32
	SubmissionBuffer    uint32
}

// BuildTransactions start a goroutine to process [TransactionPayload] and spit out [RawTransactionImpl].
func (client *Client) BuildTransactions(sender AccountAddress, payloads chan TransactionBuildPayload, responses chan TransactionBuildResponse, setSequenceNumber chan uint64, options ...any) {
	client.nodeClient.BuildTransactions(sender, payloads, responses, setSequenceNumber, options...)
}

// BuildTransactions start a goroutine to process [TransactionPayload] and spit out [RawTransactionImpl].
func (rc *NodeClient) BuildTransactions(sender AccountAddress, payloads chan TransactionBuildPayload, responses chan TransactionBuildResponse, setSequenceNumber chan uint64, options ...any) {
	// Initialize state
	defer close(responses)
	account, err := rc.Account(sender)
	if err != nil {
		responses <- TransactionBuildResponse{Err: err}
		return
	}
	sequenceNumber, err := account.SequenceNumber()
	if err != nil {
		responses <- TransactionBuildResponse{Err: err}
		return
	}
	snt := sequenceNumber
	optionsLast := len(options)
	options = append(options, SequenceNumber(0))

	for {
		select {
		case payload, ok := <-payloads:
			// End if it's not closed
			if !ok {
				return
			}
			switch payload.Type {
			case TransactionSubmissionTypeSingle:
				options[optionsLast] = SequenceNumber(snt)
				snt++
				txnResponse, err := rc.BuildTransaction(sender, payload.Inner, options...)
				if err != nil {
					responses <- TransactionBuildResponse{Err: err}
				} else {
					responses <- TransactionBuildResponse{Response: txnResponse}
				}
			case TransactionSubmissionTypeMultiAgent:
				options[optionsLast] = SequenceNumber(snt)
				snt++
				txnResponse, err := rc.BuildTransactionMultiAgent(sender, payload.Inner, options...)
				if err != nil {
					responses <- TransactionBuildResponse{Err: err}
				} else {
					responses <- TransactionBuildResponse{Response: txnResponse}
				}
			default:
				// Skip the payload
			}
		case newSequenceNumber := <-setSequenceNumber:
			// This can be used to update the sequence number at anytime
			snt = newSequenceNumber
			// TODO: We should periodically handle reconciliation of the sequence numbers, but this needs to know submission as well
		}
	}
}

// SubmitTransactions consumes signed transactions, submits to aptos-node, yields responses.
// closes output chan `responses` when input chan `signedTxns` is closed.
func (client *Client) SubmitTransactions(requests chan TransactionSubmissionRequest, responses chan TransactionSubmissionResponse) {
	client.nodeClient.SubmitTransactions(requests, responses)
}

// SubmitTransactions consumes signed transactions, submits to aptos-node, yields responses.
// closes output chan `responses` when input chan `signedTxns` is closed.
func (rc *NodeClient) SubmitTransactions(requests chan TransactionSubmissionRequest, responses chan TransactionSubmissionResponse) {
	defer close(responses)
	for request := range requests {
		response, err := rc.SubmitTransaction(request.SignedTxn)
		if err != nil {
			responses <- TransactionSubmissionResponse{Id: request.Id, Err: err}
		} else {
			responses <- TransactionSubmissionResponse{Id: request.Id, Response: response}
		}
	}
}

// BatchSubmitTransactions consumes signed transactions, submits to aptos-node, yields responses.
// closes output chan `responses` when input chan `signedTxns` is closed.
func (rc *NodeClient) BatchSubmitTransactions(requests chan TransactionSubmissionRequest, responses chan TransactionSubmissionResponse) {
	defer close(responses)

	inputs := make([]*SignedTransaction, 20)
	ids := make([]uint64, 20)
	i := uint32(0)

	for request := range requests {
		// Collect 20 inputs before submitting
		// TODO: Handle a timeout or something associated for it
		inputs[i] = request.SignedTxn
		ids[i] = request.Id

		if i >= 19 {
			i = 0
			response, err := rc.BatchSubmitTransaction(inputs)

			// Process the responses
			if err != nil {
				// Error, means all failed
				for j := range i {
					responses <- TransactionSubmissionResponse{Id: ids[j], Err: err}
				}
			} else {
				// Partial failure, means we need to send errors for those that failed
				// and responses for those that succeeded

				for j := range i {
					failed := -1
					for k := range len(response.TransactionFailures) {
						if response.TransactionFailures[k].TransactionIndex == j {
							failed = k
							break
						}
					}
					if failed >= 0 {
						responses <- TransactionSubmissionResponse{Id: ids[j], Response: nil}
					} else {
						responses <- TransactionSubmissionResponse{Id: ids[j], Err: fmt.Errorf("transaction failed: %s", response.TransactionFailures[failed].Error.Message)}
					}
				}
			}
		}
		i++
	}
}

// BuildSignAndSubmitTransactions starts up a goroutine to process transactions for a single [TransactionSender]
// Closes output chan `responses` on completion of input chan `payloads`.
func (client *Client) BuildSignAndSubmitTransactions(
	sender TransactionSigner,
	payloads chan TransactionBuildPayload,
	responses chan TransactionSubmissionResponse,
	buildOptions ...any,
) {
	client.nodeClient.BuildSignAndSubmitTransactions(sender, payloads, responses, buildOptions...)
}

// BuildSignAndSubmitTransactions starts up a goroutine to process transactions for a single [TransactionSender]
// Closes output chan `responses` on completion of input chan `payloads`.
func (rc *NodeClient) BuildSignAndSubmitTransactions(
	sender TransactionSigner,
	payloads chan TransactionBuildPayload,
	responses chan TransactionSubmissionResponse,
	buildOptions ...any,
) {
	singleSigner := func(rawTxn RawTransactionImpl) (*SignedTransaction, error) {
		switch rawTxn := rawTxn.(type) {
		case *RawTransaction:
			return rawTxn.SignedTransaction(sender)
		case *RawTransactionWithData:
			switch rawTxn.Variant {
			case MultiAgentRawTransactionWithDataVariant:
				return nil, errors.New("multi agent not supported, please provide a signer function")
			case MultiAgentWithFeePayerRawTransactionWithDataVariant:
				return nil, errors.New("fee payer not supported, please provide a signer function")
			default:
				return nil, errors.New("unsupported rawTransactionWithData type")
			}
		default:
			return nil, errors.New("unsupported rawTransactionImpl type")
		}
	}

	rc.BuildSignAndSubmitTransactionsWithSignFunction(
		sender.AccountAddress(),
		payloads,
		responses,
		singleSigner,
		buildOptions...,
	)
}

// BuildSignAndSubmitTransactionsWithSignFunction allows for signing with a custom function
//
// Closes output chan `responses` on completion of input chan `payloads`.
//
// This enables the ability to do fee payer, and other approaches while staying concurrent
//
//	func Example() {
//		client := NewNodeClient()
//
//		sender := NewEd25519Account()
//		feePayer := NewEd25519Account()
//
//		payloads := make(chan TransactionBuildPayload)
//		responses := make(chan TransactionSubmissionResponse)
//
//		signingFunc := func(rawTxn RawTransactionImpl) (*SignedTransaction, error) {
//			switch rawTxn.(type) {
//			case *RawTransaction:
//				return nil, fmt.Errorf("only fee payer supported")
//			case *RawTransactionWithData:
//				rawTxnWithData := rawTxn.(*RawTransactionWithData)
//				switch rawTxnWithData.Variant {
//				case MultiAgentRawTransactionWithDataVariant:
//					return nil, fmt.Errorf("multi agent not supported, please provide a fee payer function")
//				case MultiAgentWithFeePayerRawTransactionWithDataVariant:
//					rawTxnWithData.Sign(sender)
//					txn, ok := rawTxnWithData.ToFeePayerTransaction()
//				default:
//					return nil, fmt.Errorf("unsupported rawTransactionWithData type")
//				}
//			default:
//				return nil, fmt.Errorf("unsupported rawTransactionImpl type")
//			}
//		}
//
//		// startup worker
//		go client.BuildSignAndSubmitTransactionsWithSignFunction(
//			sender,
//			payloads,
//			responses,
//			signingFunc
//		)
//
//		// Here add payloads, and wiating on resposnes
//
//	}
func (client *Client) BuildSignAndSubmitTransactionsWithSignFunction(
	sender AccountAddress,
	payloads chan TransactionBuildPayload,
	responses chan TransactionSubmissionResponse,
	sign func(rawTxn RawTransactionImpl) (*SignedTransaction, error),
	buildOptions ...any,
) {
	client.nodeClient.BuildSignAndSubmitTransactionsWithSignFunction(
		sender,
		payloads,
		responses,
		sign,
		buildOptions...,
	)
}

// BuildSignAndSubmitTransactionsWithSignFunction allows for signing with a custom function
//
// Closes output chan `responses` on completion of input chan `payloads`.
//
// This enables the ability to do fee payer, and other approaches while staying concurrent
//
//	func Example() {
//		client := NewNodeClient()
//
//		sender := NewEd25519Account()
//		feePayer := NewEd25519Account()
//
//		payloads := make(chan TransactionBuildPayload)
//		responses := make(chan TransactionSubmissionResponse)
//
//		signingFunc := func(rawTxn RawTransactionImpl) (*SignedTransaction, error) {
//			switch rawTxn.(type) {
//			case *RawTransaction:
//				return nil, fmt.Errorf("only fee payer supported")
//			case *RawTransactionWithData:
//				rawTxnWithData := rawTxn.(*RawTransactionWithData)
//				switch rawTxnWithData.Variant {
//				case MultiAgentRawTransactionWithDataVariant:
//					return nil, fmt.Errorf("multi agent not supported, please provide a fee payer function")
//				case MultiAgentWithFeePayerRawTransactionWithDataVariant:
//					rawTxnWithData.Sign(sender)
//					txn, ok := rawTxnWithData.ToFeePayerTransaction()
//				default:
//					return nil, fmt.Errorf("unsupported rawTransactionWithData type")
//				}
//			default:
//				return nil, fmt.Errorf("unsupported rawTransactionImpl type")
//			}
//		}
//
//		// startup worker
//		go client.BuildSignAndSubmitTransactionsWithSignFunction(
//			sender,
//			payloads,
//			responses,
//			signingFunc
//		)
//
//		// Here add payloads, and wiating on resposnes
//
//	}
func (rc *NodeClient) BuildSignAndSubmitTransactionsWithSignFunction(
	sender AccountAddress,
	payloads chan TransactionBuildPayload,
	responses chan TransactionSubmissionResponse,
	sign func(rawTxn RawTransactionImpl) (*SignedTransaction, error),
	buildOptions ...any,
) {
	// TODO: Make internal buffer size configurable with an optional parameter

	// Set up the channel handling building transactions
	buildResponses := make(chan TransactionBuildResponse, 20)
	setSequenceNumber := make(chan uint64)
	go rc.BuildTransactions(sender, payloads, buildResponses, setSequenceNumber, buildOptions...)

	submissionRequests := make(chan TransactionSubmissionRequest, 20)
	// Note that, I change this to BatchSubmitTransactions, and it caused no change in performance.  The non-batched
	// version is more flexible and gives actual responses.  It is may be that with large payloads that batch more performant.
	go rc.SubmitTransactions(submissionRequests, responses)

	var wg sync.WaitGroup

	for buildResponse := range buildResponses {
		if buildResponse.Err != nil {
			responses <- TransactionSubmissionResponse{Id: buildResponse.Id, Err: buildResponse.Err}
		} else {
			// TODO: replace this with a fixed number (configurable) of sign() workers
			wg.Add(1)
			go func() {
				defer wg.Done()
				signedTxn, err := sign(buildResponse.Response)
				if err != nil {
					responses <- TransactionSubmissionResponse{Id: buildResponse.Id, Err: err}
				} else {
					submissionRequests <- TransactionSubmissionRequest{
						Id:        buildResponse.Id,
						SignedTxn: signedTxn,
					}
				}
			}()
		}
	}

	wg.Wait()
	close(submissionRequests)
}

func (cfg WorkerPoolConfig) getBufferSizes() (uint32, uint32) {
	workers := defaultWorkerCount
	if cfg.NumWorkers > 0 {
		workers = cfg.NumWorkers
	}

	build := workers
	if cfg.BuildResponseBuffer > 0 {
		build = cfg.BuildResponseBuffer
	}

	submission := workers
	if cfg.SubmissionBuffer > 0 {
		submission = cfg.SubmissionBuffer
	}

	return build, submission
}

// startSigningWorkers initializes and starts a pool of worker goroutines for signing transactions.
func startSigningWorkers(
	numWorkers uint32,
	sign func(rawTxn RawTransactionImpl) (*SignedTransaction, error),
	submissionRequests chan<- TransactionSubmissionRequest,
	responses chan<- TransactionSubmissionResponse,
	signingWg *sync.WaitGroup,
	transactionWg *sync.WaitGroup,
) chan TransactionBuildResponse {
	transactionsToSign := make(chan TransactionBuildResponse, numWorkers)

	signingWg.Add(int(numWorkers))
	for range numWorkers {
		go func() {
			defer signingWg.Done()
			for buildResponse := range transactionsToSign {
				signedTxn, err := sign(buildResponse.Response)
				if err != nil {
					responses <- TransactionSubmissionResponse{Id: buildResponse.Id, Err: err}
				} else {
					submissionRequests <- TransactionSubmissionRequest{
						Id:        buildResponse.Id,
						SignedTxn: signedTxn,
					}
				}
				transactionWg.Done()
			}
		}()
	}

	return transactionsToSign
}

// BuildSignAndSubmitTransactionsWithSignFnAndWorkerPool processes transactions using a fixed-size worker pool.
// It coordinates three stages of the pipeline:
// 1. Building transactions (BuildTransactions)
// 2. Signing transactions (worker pool)
// 3. Submitting transactions (SubmitTransactions)
func (rc *NodeClient) BuildSignAndSubmitTransactionsWithSignFnAndWorkerPool(
	sender AccountAddress,
	payloads chan TransactionBuildPayload,
	responses chan TransactionSubmissionResponse,
	sign func(rawTxn RawTransactionImpl) (*SignedTransaction, error),
	workerPoolConfig WorkerPoolConfig,
	buildOptions ...any,
) {
	buildBuffer, submissionBuffer := workerPoolConfig.getBufferSizes()
	numWorkers := workerPoolConfig.NumWorkers
	if numWorkers == 0 {
		numWorkers = defaultWorkerCount
	}

	buildResponses := make(chan TransactionBuildResponse, buildBuffer)
	setSequenceNumber := make(chan uint64)
	go rc.BuildTransactions(sender, payloads, buildResponses, setSequenceNumber, buildOptions...)

	submissionRequests := make(chan TransactionSubmissionRequest, submissionBuffer)
	go rc.SubmitTransactions(submissionRequests, responses)

	var signingWg sync.WaitGroup
	var transactionWg sync.WaitGroup

	transactionsToSign := startSigningWorkers(numWorkers, sign, submissionRequests, responses, &signingWg, &transactionWg)

	for buildResponse := range buildResponses {
		if buildResponse.Err != nil {
			responses <- TransactionSubmissionResponse{Id: buildResponse.Id, Err: buildResponse.Err}
			continue
		}
		transactionWg.Add(1)
		transactionsToSign <- buildResponse
	}

	// 1. Wait for all transactions to complete processing
	transactionWg.Wait()
	// 2. Close signing channel to signal workers to shut down
	close(transactionsToSign)
	// 3. Wait for all workers to finish and clean up
	signingWg.Wait()
	// 4. Close submission channel after all signing is done
	close(submissionRequests)
}
