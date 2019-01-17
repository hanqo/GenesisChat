package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"testing"
)
const addr = "0xb04b61254B42d64f17938E5DCe2eb728cAfF8937"

func generateRawTxNaive() string{

	client, err := ethclient.Dial("http://127.0.0.1:7545") //Ganache local address
	if err != nil {
		log.Fatal(err)
	}

	private_key := "4b62386099abd28f2b63d3a08918cbffc72f4752e3a029747f2a4681b28021c7"
	receiver := "0x218778aA387BCCD5167B6881B4Fc210f0ebFe5Ae"

	chainID := big.NewInt(5777) // ganache

	privateKey, _ := crypto.HexToECDSA(private_key)
	recipientAddr := common.HexToAddress(receiver)
	amount := big.NewInt(1000000000000000000) // 1 ether
	gasLimit := uint64(100000)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, _ := client.PendingNonceAt(context.Background(), fromAddress)

	gasPrice, _ := client.SuggestGasPrice(context.Background())

	tx := types.NewTransaction(nonce, recipientAddr, amount, gasLimit, gasPrice, nil)

	signer := types.NewEIP155Signer(chainID)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)

	fmt.Printf("raw tx hex: %s\n",rawTxHex)
	return rawTxHex
}

func generateRawTxDeployContract(gas uint64, gasPrice int64, nonce uint64, data []byte) string{

	privateStr := "4b62386099abd28f2b63d3a08918cbffc72f4752e3a029747f2a4681b28021c7"

	chainID := big.NewInt(5777) // ganache

	privateKey, _ := crypto.HexToECDSA(privateStr)
	amount := big.NewInt(0) // 1 ether

	tx := types.NewContractCreation(nonce, amount, gas, big.NewInt(gasPrice), data)

	signer := types.NewEIP155Signer(chainID)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)

	fmt.Printf("raw tx 2 hex: %s\n",rawTxHex)
	return rawTxHex
}

func generateRawTxSetContract(gas uint64, gasPrice int64, nonce uint64, data []byte) string{
	contractAddr := "0x443d9c124f875cb3ae6f419532e7dbb2a948ca8b"
	recipientAddr := common.HexToAddress(contractAddr)

	privateStr := "4b62386099abd28f2b63d3a08918cbffc72f4752e3a029747f2a4681b28021c7"

	chainID := big.NewInt(5777) // ganache

	privateKey, _ := crypto.HexToECDSA(privateStr)
	amount := big.NewInt(0) // 1 ether

	tx := types.NewTransaction(nonce, recipientAddr, amount, gas, big.NewInt(gasPrice), data)

	signer := types.NewEIP155Signer(chainID)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)

	fmt.Printf("raw tx 2 hex: %s\n",rawTxHex)
	return rawTxHex
}

//Cannot use name TestforwardRawTx, although the function name is forwardRawTx.
func TestSendSignedTx(t *testing.T) {
	//----------------------Test naive transaction and polling--------------------------------------------------------//
	h := NewETHHandler()
	rawTxData1 :=generateRawTxNaive()

	h.toChains <- &MsgToChain{
		from :     addr,
		user :     "test1",
		version :  "1",
		chainID :  123,
		typ:"signed_tx",
		signedTx:  &rawTxData1,}
	for {
		select {
		case msg := <-h.fromChains:

			if msg.txSent != nil {
				fmt.Printf("naive tx Sent with hash: %s\n", msg.txSent.txHash)
				fmt.Printf("naive tx Sent with gas price: %d\n", msg.txSent.gasPrice)
				fmt.Printf("naive tx Sent with nonce: %d\n", msg.txSent.nonce)
				fmt.Printf("naive tx Sent with estimated gas: %d\n", msg.txSent.gasEstimated)
			} else if msg.txReceipt != nil {
				fmt.Printf("naive tx Confirmed with hash: %s\n", msg.txReceipt.txHash)
				fmt.Printf("naive tx Confirmed with gas price: %d\n", msg.txReceipt.gasUsed)
				return
			}

		}
	}
}

func TestDeployContract(t *testing.T) {
	h := NewETHHandler()
	m:= &MsgToChain{
		from :     		addr,
		user :     		"test1",
		version :  		"1",
		chainID :  		123,
		typ:			"request_tx",
		requestTx:  	&MsgContractFunc{
			function:		"",
			inputs:			[]string{"kingdom","xrisheng","100","1200"},},}

	h.toChains <- m
	for{
		select {
		case msg := <-h.fromChains:
			if msg.txSent != nil {
				fmt.Printf("deploy tx Sent with hash: %s\n", msg.txSent.txHash)
				fmt.Printf("deploy tx Sent with gas price: %d\n", msg.txSent.gasPrice)
				fmt.Printf("deploy tx Sent with nonce: %d\n", msg.txSent.nonce)
				fmt.Printf("deploy tx Sent with estimated gas: %d\n", msg.txSent.gasEstimated)
			}else if msg.txInfo != nil{
				if msg.txInfo.function!= m.requestTx.function{
					t.Error("function name not equal")
				}
				if msg.txInfo.gasPrice <= 0{
					t.Error("gas price not greater than 0")
				}
				if msg.txInfo.nonce <= 0{
					t.Error("nonce not greater than 0")
				}
				if msg.txInfo.data == nil{
					t.Error("data field is nil")
				}
				rawTxData2 := generateRawTxDeployContract(5000000,msg.txInfo.gasPrice,msg.txInfo.nonce,msg.txInfo.data)

				h.toChains <- &MsgToChain{
					from :     addr,
					user :     "test1",
					version :  "1",
					chainID :  123,
					typ:"signed_tx",
					signedTx:  &rawTxData2,}

			} else if msg.txReceipt != nil{
				if msg.txReceipt.contractAddr == nil{
					t.Error("Contract Address is nil, deploy failed")
				}
				fmt.Printf("deploy tx Confirmed with hash: %s\n", msg.txReceipt.txHash)
				fmt.Printf("deploy tx Confirmed with gas used: %d\n", msg.txReceipt.gasUsed)
				fmt.Printf("contract deployed at address %s\n", *msg.txReceipt.contractAddr)
				return
			}
		}
	}
}

func TestSetContract(t *testing.T) {
	h := NewETHHandler()
	m := &MsgToChain{
		from:    addr,
		user:    "test1",
		version: "1",
		chainID: 123,
		typ:     "request_tx",
		requestTx: &MsgContractFunc{
			function: "setEntryCost",
			inputs:   []string{"500"},},}

	h.toChains <- m
	for {
		select {
		case msg := <-h.fromChains:
			if msg.txSent != nil {
				fmt.Printf("set tx Sent with hash: %s\n", msg.txSent.txHash)
				fmt.Printf("set tx Sent with gas price: %d\n", msg.txSent.gasPrice)
				fmt.Printf("set tx Sent with nonce: %d\n", msg.txSent.nonce)
				fmt.Printf("set tx Sent with estimated gas: %d\n", msg.txSent.gasEstimated)
			} else if msg.txInfo != nil {
				rawTxData3 := generateRawTxSetContract(500000, msg.txInfo.gasPrice, msg.txInfo.nonce, msg.txInfo.data)

				h.toChains <- &MsgToChain{
					from:     addr,
					user:     "test1",
					version:  "1",
					chainID:  123,
					typ:      "signed_tx",
					signedTx: &rawTxData3,}
			}else if msg.txReceipt != nil{
				fmt.Printf("set tx Confirmed with hash: %s\n", msg.txReceipt.txHash)
				fmt.Printf("set tx Confirmed with gas used: %d\n", msg.txReceipt.gasUsed)
				return
			}

		}
	}
}

func TestCallContract(t *testing.T) {
	h := NewETHHandler()

	m := &MsgToChain{
		from:    addr,
		user:    "test1",
		version: "1",
		chainID: 123,
		typ:     "contract_call",
		call: &MsgCall{
			contractAddr: "0x443d9c124f875cb3ae6f419532e7dbb2a948ca8b",
			contractFunc: MsgContractFunc{
				function: "getEntryCost",},
		},}

	h.toChains <- m
	for {
		select {
		case msg := <-h.fromChains:
			if msg.callReturn != nil {
				fmt.Printf("call return from address: %s\n", msg.callReturn.contractAddr)
				fmt.Printf("call return of function: %s\n", msg.callReturn.function)
				fmt.Printf("call return with output: %s\n", msg.callReturn.output)
				return
			}
		}

	}
}