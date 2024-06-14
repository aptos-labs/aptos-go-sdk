package aptos

import (
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"sync"
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
	SequenceNumber uint64
	Mutex          sync.Mutex
}

func (snt *SequenceNumberTracker) Increment() uint64 {
	snt.Mutex.Lock()
	seqNumber := snt.SequenceNumber
	snt.SequenceNumber = snt.SequenceNumber + 1
	snt.Mutex.Unlock()
	return seqNumber
}

func (snt *SequenceNumberTracker) Update(new uint64) uint64 {
	snt.Mutex.Lock()
	seqNumber := snt.SequenceNumber
	snt.SequenceNumber = new
	snt.Mutex.Unlock()
	return seqNumber
}

// BuildTransactions start a goroutine to process [TransactionPayload] and spit out [RawTransactionImpl].
//
// TODO: add optional arguments for configuring transactions as a whole?
func (rc *NodeClient) BuildTransactions(sender AccountAddress, payloads chan TransactionSubmissionPayload, responses chan TransactionBuildResponse, setSequenceNumber chan uint64) {
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
	snt := &SequenceNumberTracker{SequenceNumber: sequenceNumber}

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
				go func() {
					curSequenceNumber := snt.Increment()
					txnResponse, err := rc.BuildTransaction(sender, payload.Inner, SequenceNumber(curSequenceNumber))
					if err != nil {
						responses <- TransactionBuildResponse{Err: err}
						return
					}
					responses <- TransactionBuildResponse{Response: txnResponse}
				}()
			case TransactionSubmissionTypeMultiAgent:
				go func() {
					curSequenceNumber := snt.Increment()
					txnResponse, err := rc.BuildTransactionMultiAgent(sender, payload.Inner, SequenceNumber(curSequenceNumber))
					if err != nil {
						responses <- TransactionBuildResponse{Err: err}
						return
					}
					responses <- TransactionBuildResponse{Response: txnResponse}
				}()
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

// SubmitTransactions starts up a worker for sending signed transactions to on-chain
func (rc *NodeClient) SubmitTransactions(signedTxns chan *SignedTransaction, responses chan TransactionSubmissionResponse) {
	for {
		select {
		case signedTxn, ok := <-signedTxns:
			if !ok {
				close(responses)
				return
			}

			response, err := rc.SubmitTransaction(signedTxn)
			if err != nil {
				responses <- TransactionSubmissionResponse{Err: err}
			} else {
				responses <- TransactionSubmissionResponse{Response: response}
			}
		}
	}
}

// BuildSignAndSubmitTransactions starts up a goroutine to process transactions for a single [TransactionSender]
func (rc *NodeClient) BuildSignAndSubmitTransactions(
	sender TransactionSigner,
	payloads chan TransactionSubmissionPayload,
	responses chan TransactionSubmissionResponse,
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
	)
}

// BuildSignAndSubmitTransactionsWithSignFunction allows for signing with a custom function
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
) {
	// TODO: Make internal buffer size configurable with an optional parameter

	// Set up the channel handling building transactions
	buildResponses := make(chan TransactionBuildResponse, 20)
	setSequenceNumber := make(chan uint64)
	go rc.BuildTransactions(sender, payloads, buildResponses, setSequenceNumber)

	signedTxns := make(chan *SignedTransaction, 20)
	go rc.SubmitTransactions(signedTxns, responses)

	for {
		select {
		case buildResponse, ok := <-buildResponses:
			// Input closed, close output
			if !ok {
				close(responses)
			} else if buildResponse.Err != nil {
				responses <- TransactionSubmissionResponse{Id: buildResponse.Id, Err: buildResponse.Err}
			} else {
				go func() {
					signedTxn, err := sign(buildResponse.Response)
					if err != nil {
						responses <- TransactionSubmissionResponse{Id: buildResponse.Id, Err: err}
					} else {
						signedTxns <- signedTxn
					}
				}()
			}
		}
	}
}
