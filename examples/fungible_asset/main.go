// fungible_asset is an example of how to create and transfer fungible assets
package main

import (
	"encoding/json"
	"fmt"
	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
)

const testEd25519PrivateKey = "0xc5338cd251c22daa8c9c9cc94f498cc8a5c7e1d2e75287a5dda91096fe64efa5"
const rupeePublisherAddress = "0x978c213990c4833df71548df7ce49d54c759d6b6d932de22b24d56060b7af2aa"

// These come from fungible_asset.json
const metadata = "0x0572757065650100000000000000004038393534453933384132434137314536433445313139434230333341363036453341333537424245353843354430304235453132354236383238423745424331e7011f8b08000000000002ff3d8ecd6ec4200c84ef3cc58a7b13427ea9d4432f7d8928aa0c7636d136100149fbf885ed764ff68cbeb167dcc1dce04a13b3b0d1e5edc2fdb1137176920fabb3d9a90a5108cee0888bf32139e3c4d808889e42a030b17be4331b19173faa1f8cac6aa5846986bac6b9afda6648c350a3b06d4cdf2aec7487aa9648526ad960db894ee81e6609c0d379a49d2c92352b85e27d8f2e7cf8d4f0dbf9dbc4ae6bcc9f9618f7f05a96492e872e8cdb4ac8e4cb17e8f0588df3542480334f670e6db05a4b498743e37a6ffc476eeea472fe7ff2883f3567bf9aa419822b010000010572757065650000000300000000000000000000000000000000000000000000000000000000000000010e4170746f734672616d65776f726b00000000000000000000000000000000000000000000000000000000000000010b4170746f735374646c696200000000000000000000000000000000000000000000000000000000000000010a4d6f76655374646c696200"
const bytecode = "0xa11ceb0b060000000c01000e020e30033e850104c3010605c901980107e102c30408a4074006e407950110f9088a010a830a120c950ae1020df60c060000010101020103010401050106000708000110060001120600011406000212060002170600011a0800021b07010001021e02000323070100000625070000080001000009010200000a030100000b030100000c000100000d040100000e05010005180302000219070200021c02090108021d0a0b010804080c0100021f0e0f000220101100022110120002221301000324011501000626161700042718010001281019000121101a000129101b00022a101c00040c1d0100042b1e0100042c1f010009080a08101403060c050300010501060c03060c050104060c050503030505050206050a02010806010b07010900020b070109000501010306080305030708030808080108050c0804080202060c0a0201080801060808010805010804010608040104010b09010900010a0201080a070608080b090104080a080a02080a080a010801010802010803010c030608010503030608020501040608020505030572757065650e66756e6769626c655f6173736574066f626a656374066f7074696f6e167072696d6172795f66756e6769626c655f73746f7265067369676e657206737472696e6709527570656552656673046275726e0a66615f616464726573730b696e69745f6d6f64756c650a696e697469616c697a65046d696e740a7365745f667265657a65087472616e73666572086d696e745f726566074d696e745265660c7472616e736665725f7265660b5472616e73666572526566086275726e5f726566074275726e526566106f626a5f7472616e736665725f7265660e6f626a5f657874656e645f72656609457874656e645265660a616464726573735f6f66156372656174655f6f626a6563745f61646472657373084d65746164617461064f626a65637411616464726573735f746f5f6f626a6563740869735f6f776e65720e436f6e7374727563746f72526566136372656174655f6e616d65645f6f626a6563741367656e65726174655f657874656e645f7265661567656e65726174655f7472616e736665725f7265661864697361626c655f756e67617465645f7472616e73666572064f7074696f6e046e6f6e6506537472696e6704757466382b6372656174655f7072696d6172795f73746f72655f656e61626c65645f66756e6769626c655f61737365741167656e65726174655f6d696e745f7265661167656e65726174655f6275726e5f7265660f67656e65726174655f7369676e65720f7365745f66726f7a656e5f666c6167117472616e736665725f776974685f726566978c213990c4833df71548df7ce49d54c759d6b6d932de22b24d56060b7af2aa0000000000000000000000000000000000000000000000000000000000000001030801000000000000000201020a02060552757065650a021413527570656573417265466f7254657374696e670a02060552555045450a023635697066733a2f2f516d585342534c6f337744426e6133314d4c56784b4a356f424a486b676f444d7845676a527a39745044665a506e0520978c213990c4833df71548df7ce49d54c759d6b6d932de22b24d56060b7af2aa0a020100126170746f733a3a6d657461646174615f7631760101000000000000000b455f4e4f545f4f574e45522a43616c6c6572206973206e6f74206f776e6572206f6620746865206d65746164617461206f626a6563740109527570656552656673010301183078313a3a6f626a6563743a3a4f626a65637447726f7570010a66615f616464726573730101000002050f08011108021308031508041608050000040100061a0b0011070c0507060c030e030703110838000b053801040d050f07002707060c040e04070311082b0010000b010b02110b0201010000020607060c000e0007031108020200000001030b00110302030000000d2d0b000703110c0c020e02110d0c040e02110e0c060e06110f0e02380207021111070411110701070511110707111111120e0211130c030e0211140c070e0211150c010e0211160c050e050b030b070b010b060b0412002d00020400040100061a0b0011070c0507060c030e030703110838000b053801040d050f07002707060c040e04070311082b0010010b010b021117020500040100061a0b0011070c0507060c030e030703110838000b053801040d050f07002707060c040e04070311082b0010020b010b021118020600040100061b0b0011070c0607060c040e040703110838000b063801040d050f07002707060c050e05070311082b0010020b010b020b0311190200020000000100"
const TransferAmount = uint64(1_000)
const FundAmount = uint64(100_000_000)

func example(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	// Create a sender locally
	key := crypto.Ed25519PrivateKey{}
	err = key.FromHex(testEd25519PrivateKey)
	if err != nil {
		panic("Failed to decode Ed25519 private key:" + err.Error())
	}
	sender, err := aptos.NewAccountFromSigner(&key)
	if err != nil {
		panic("Failed to create sender:" + err.Error())
	}

	// Fund the sender with the faucet to create it on-chain
	println("SENDER: ", sender.Address.String())
	err = client.Fund(sender.Address, FundAmount)
	if err != nil {
		panic("Failed to fund sender:" + err.Error())
	}

	// Publish the package for FA
	metadataBytes, err := aptos.ParseHex(metadata)
	bytecodeBytes, err := aptos.ParseHex(bytecode)
	payload, err := aptos.PublishPackagePayloadFromJsonFile(metadataBytes, [][]byte{bytecodeBytes})
	if err != nil {
		panic("Failed to create publish payload:" + err.Error())
	}
	response, err := client.BuildSignAndSubmitTransaction(sender, *payload)
	if err != nil {
		panic("Failed to build sign and submit publish transaction:" + err.Error())
	}
	waitResponse, err := client.WaitForTransaction(response.Hash)
	if err != nil {
		panic("Failed to wait for publish transaction:" + err.Error())
	}
	if waitResponse.Success == false {
		responseStr, _ := json.Marshal(response)
		panic(fmt.Sprintf("Failed to publish transaction %s", responseStr))
	}

	// Get the fungible asset address by view function
	rupeeModule := aptos.ModuleId{Address: sender.Address, Name: "rupee"}
	var noTypeTags []aptos.TypeTag
	viewResponse, err := client.View(&aptos.ViewPayload{
		Module:   rupeeModule,
		Function: "fa_address",
		ArgTypes: noTypeTags,
		Args:     [][]byte{},
	})
	if err != nil {
		panic("Failed to view fa address:" + err.Error())
	}
	faMetadataAddress := &aptos.AccountAddress{}
	err = faMetadataAddress.ParseStringRelaxed(viewResponse[0].(string))
	if err != nil {
		panic("Failed to parse fa address:" + err.Error())
	}
	faClient, err := aptos.NewFungibleAssetClient(client, faMetadataAddress)
	if err != nil {
		panic("Failed to create fungible asset client:" + err.Error())
	}

	beforeBalance, err := faClient.PrimaryBalance(&sender.Address)
	if err != nil {
		panic("Failed to get balance:" + err.Error())
	}

	// Let's mint and transfer some coins
	amount, err := bcs.SerializeU64(FundAmount)
	if err != nil {
		panic("Failed to serialize amount:" + err.Error())
	}
	serializedSenderAddress, _ := bcs.Serialize(&sender.Address) // This can't fail
	response, err = client.BuildSignAndSubmitTransaction(sender, aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module:   rupeeModule,
			Function: "mint",
			ArgTypes: noTypeTags,
			Args:     [][]byte{serializedSenderAddress, amount},
		},
	})
	if err != nil {
		panic("Failed to build sign and submit mint transaction:" + err.Error())
	}
	fmt.Printf("Submitted mint as: %s\n", response.Hash)
	_, err = client.WaitForTransaction(response.Hash)
	if err != nil {
		panic("Failed to wait for publish transaction:" + err.Error())
	}

	afterBalance, err := faClient.PrimaryBalance(&sender.Address)
	if err != nil {
		panic("Failed to get balance:" + err.Error())
	}

	fmt.Printf("Before mint: %d, after mint: %d\n", beforeBalance, afterBalance)

	receiver := &aptos.AccountAddress{}
	err = receiver.ParseStringRelaxed("0xCAFE")
	if err != nil {
		panic("Failed to parse receiver address:" + err.Error())
	}

	// Transfer some to 0xCAFE
	receiverBeforeBalance, err := faClient.PrimaryBalance(receiver)
	if err != nil {
		panic("Failed to get balance:" + err.Error())
	}
	transferTxn, err := faClient.TransferPrimaryStore(sender, *receiver, TransferAmount)
	if err != nil {
		panic("Failed to create primary store transfer transaction:" + err.Error())
	}
	response, err = client.SubmitTransaction(transferTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	fmt.Printf("Submitted transfer as: %s\n", response.Hash)
	err = client.PollForTransactions([]string{response.Hash})
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}
	receiverAfterBalance, err := faClient.PrimaryBalance(receiver)
	if err != nil {
		panic("Failed to get store balance:" + err.Error())
	}

	fmt.Printf("Receiver Before transfer: %d, after transfer: %d\n", receiverBeforeBalance, receiverAfterBalance)

	// Now run a script version
	fmt.Printf("\n== Now running script version ==\n")
	runScript(client, sender, receiver, faMetadataAddress)

	receiverAfterAfterBalance, err := faClient.PrimaryBalance(receiver)
	if err != nil {
		panic("Failed to get store balance:" + err.Error())
	}
	fmt.Printf("After Script: Receiver Before transfer: %d, after transfer: %d\n", receiverAfterBalance, receiverAfterAfterBalance)
}

const TransferScript = "0xa11ceb0b0700000a0701000a020a0a03141a042e040532240756870108dd01200000000100020003000403050701000102060b0003070304010801000803060001010903050001040a07050108010002030204060c050503010b000108010108010105010b0001090000010104060c0b000109000503076163636f756e740d6170746f735f6163636f756e740e66756e6769626c655f6173736574066f626a656374167072696d6172795f66756e6769626c655f73746f7265064f626a656374084d6574616461746111616464726573735f746f5f6f626a656374096578697374735f61740e6372656174655f6163636f756e74087472616e7366657200000000000000000000000000000000000000000000000000000000000000010000010f0b0138000c040a0211012004090a0211020b000b040b020b03380102"

func runScript(client *aptos.Client, alice *aptos.Account, bob *aptos.AccountAddress, faMetadataAddress *aptos.AccountAddress) {
	scriptBytes, err := aptos.ParseHex(TransferScript)
	if err != nil {
		panic("Failed to parse script:" + err.Error())
	}

	// 1. Build transaction
	rawTxn, err := client.BuildTransaction(alice.AccountAddress(), aptos.TransactionPayload{
		Payload: &aptos.Script{
			Code:     scriptBytes,
			ArgTypes: []aptos.TypeTag{},
			Args: []aptos.ScriptArgument{
				{Variant: aptos.ScriptArgumentAddress, Value: *faMetadataAddress},
				{Variant: aptos.ScriptArgumentAddress, Value: *bob},
				{Variant: aptos.ScriptArgumentU64, Value: TransferAmount},
			},
		}},
	)

	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// 2. Simulate transaction (optional)
	// This is useful for understanding how much the transaction will cost
	// and to ensure that the transaction is valid before sending it to the network
	// This is optional, but recommended
	simulationResult, err := client.SimulateTransaction(rawTxn, alice)
	if err != nil {
		panic("Failed to simulate transaction:" + err.Error())
	}
	fmt.Printf("\n=== Simulation ===\n")
	fmt.Printf("Gas unit price: %d\n", simulationResult[0].GasUnitPrice)
	fmt.Printf("Gas used: %d\n", simulationResult[0].GasUsed)
	fmt.Printf("Total gas fee: %d\n", simulationResult[0].GasUsed*simulationResult[0].GasUnitPrice)
	fmt.Printf("Status: %s\n", simulationResult[0].VmStatus)

	// 3. Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(alice)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	// 4. Submit transaction
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// 5. Wait for the transaction to complete
	_, err = client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}
}

// main This example shows how to create and transfer fungible assets
func main() {
	example(aptos.DevnetConfig)
}
