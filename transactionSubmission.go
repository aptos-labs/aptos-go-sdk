package aptos

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/aptos-labs/aptos-go-sdk/api"
)

// TransactionSubmissionType is the counter for an enum
type TransactionSubmissionType uint8

const (
	// TransactionSubmissionTypeSingle represents a single signer transaction, no multi-agent and no-fee payer
	TransactionSubmissionTypeSingle TransactionSubmissionType = iota
	// TransactionSubmissionTypeMultiAgent represents a multi-agent or fee payer transaction
	TransactionSubmissionTypeMultiAgent TransactionSubmissionType = iota
)

type TransactionSubmissionPayload struct {
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

type TransactionSubmissionResponse struct {
	Id       uint64
	Response *api.SubmitTransactionResponse
	Err      error
}

type SequenceNumberTracker struct {
	SequenceNumber atomic.Uint64
}

func (snt *SequenceNumberTracker) Increment() uint64 {
	for {
		seqNumber := snt.SequenceNumber.Load()
		next := seqNumber + 1
		ok := snt.SequenceNumber.CompareAndSwap(seqNumber, next)
		if ok {
			return seqNumber
		}
	}
}

func (snt *SequenceNumberTracker) Update(next uint64) uint64 {
	return snt.SequenceNumber.Swap(next)
}

// BuildTransactions start a goroutine to process [TransactionPayload] and spit out [RawTransactionImpl].
func (rc *NodeClient) BuildTransactions(sender AccountAddress, payloads chan TransactionSubmissionPayload, responses chan TransactionBuildResponse, setSequenceNumber chan uint64, options ...any) {
	// Initialize state
	account, err := rc.Account(sender)
	if err != nil {
		responses <- TransactionBuildResponse{Err: err}
		close(responses)
		return
	}
	sequenceNumber, err := account.SequenceNumber()
	if err != nil {
		responses <- TransactionBuildResponse{Err: err}
		close(responses)
		return
	}
	snt := &SequenceNumberTracker{}
	snt.SequenceNumber.Store(sequenceNumber)
	optionsLast := len(options)
	options = append(options, SequenceNumber(0))

	for {
		select {
		case payload, ok := <-payloads:
			// End if it's not closed
			if !ok {
				close(responses)
				return
			}
			switch payload.Type {
			case TransactionSubmissionTypeSingle:
				curSequenceNumber := snt.Increment()
				options[optionsLast] = SequenceNumber(curSequenceNumber)
				txnResponse, err := rc.BuildTransaction(sender, payload.Inner, options...)
				if err != nil {
					responses <- TransactionBuildResponse{Err: err}
				} else {
					responses <- TransactionBuildResponse{Response: txnResponse}
				}
			case TransactionSubmissionTypeMultiAgent:
				curSequenceNumber := snt.Increment()
				options[optionsLast] = SequenceNumber(curSequenceNumber)
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
			snt.Update(newSequenceNumber)
			// TODO: We should periodically handle reconciliation of the sequence numbers, but this needs to know submission as well
		}
	}
}

// SubmitTransactions consumes signed transactions, submits to aptos-node, yields responses.
// closes output chan `responses` when input chan `signedTxns` is closed.
func (rc *NodeClient) SubmitTransactions(signedTxns chan *SignedTransaction, responses chan TransactionSubmissionResponse) {
	defer close(responses)
	for signedTxn := range signedTxns {
		response, err := rc.SubmitTransaction(signedTxn)
		if err != nil {
			responses <- TransactionSubmissionResponse{Err: err}
		} else {
			responses <- TransactionSubmissionResponse{Response: response}
		}
	}
}

// BuildSignAndSubmitTransactions starts up a goroutine to process transactions for a single [TransactionSender]
// Closes output chan `responses` on completion of input chan `payloads`.
func (rc *NodeClient) BuildSignAndSubmitTransactions(
	sender TransactionSigner,
	payloads chan TransactionSubmissionPayload,
	responses chan TransactionSubmissionResponse,
	buildOptions ...any,
) {
	singleSigner := func(rawTxn RawTransactionImpl) (*SignedTransaction, error) {
		switch rawTxn.(type) {
		case *RawTransaction:
			return rawTxn.(*RawTransaction).SignedTransaction(sender)
		case *RawTransactionWithData:
			rawTxnWithData := rawTxn.(*RawTransactionWithData)
			switch rawTxnWithData.Variant {
			case MultiAgentRawTransactionWithDataVariant:
				return nil, fmt.Errorf("multi agent not supported, please provide a signer function")
			case MultiAgentWithFeePayerRawTransactionWithDataVariant:
				return nil, fmt.Errorf("fee payer not supported, please provide a signer function")
			default:
				return nil, fmt.Errorf("unsupported rawTransactionWithData type")
			}
		default:
			return nil, fmt.Errorf("unsupported rawTransactionImpl type")
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
//		payloads := make(chan TransactionSubmissionPayload)
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
	payloads chan TransactionSubmissionPayload,
	responses chan TransactionSubmissionResponse,
	sign func(rawTxn RawTransactionImpl) (*SignedTransaction, error),
	buildOptions ...any,
) {
	// TODO: Make internal buffer size configurable with an optional parameter

	// Set up the channel handling building transactions
	buildResponses := make(chan TransactionBuildResponse, 20)
	setSequenceNumber := make(chan uint64)
	go rc.BuildTransactions(sender, payloads, buildResponses, setSequenceNumber, buildOptions...)

	signedTxns := make(chan *SignedTransaction, 20)
	go rc.SubmitTransactions(signedTxns, responses)

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
					signedTxns <- signedTxn
				}
			}()
		}
	}

	wg.Wait()
	close(signedTxns)
}
