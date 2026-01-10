// Package sponsored provides utilities for gas-sponsored (fee payer) transactions.
//
// Sponsored transactions allow a third party (the sponsor) to pay gas fees on behalf
// of the transaction sender. This enables gasless experiences for end users.
//
// # Basic Usage
//
// For a simple sponsored transaction:
//
//	// Sender creates and signs the transaction
//	rawTxn, _ := client.BuildTransaction(ctx, sender.Address(), payload,
//	    aptos.WithFeePayer(aptos.AccountZero), // Placeholder for sponsor
//	)
//
//	// Sender signs
//	signingMsg, _ := sponsored.SigningMessage(rawTxn, nil, sponsorAddr)
//	senderAuth, _ := sender.Sign(signingMsg)
//
//	// Sponsor signs (could be on different server)
//	sponsorAuth, _ := sponsor.Sign(signingMsg)
//
//	// Combine signatures and submit
//	signedTxn := sponsored.NewFeePayerTransaction(rawTxn, senderAuth, sponsorAuth, sponsorAddr)
//	result, _ := client.SubmitTransaction(ctx, signedTxn)
//
// # Gas Station Integration
//
// For integrating with a gas station service:
//
//	gasStation := sponsored.NewGasStation(gasStationURL, apiKey)
//
//	// Get sponsor signature for a transaction
//	sponsorSig, _ := gasStation.SponsorTransaction(ctx, rawTxn, senderAuth)
//
// See the examples directory for complete working examples.
package sponsored
