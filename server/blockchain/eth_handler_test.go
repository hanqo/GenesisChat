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
const priv = "4b62386099abd28f2b63d3a08918cbffc72f4752e3a029747f2a4681b28021c7"
const receiver = "0x218778aA387BCCD5167B6881B4Fc210f0ebFe5Ae"
var chainID = big.NewInt(5777)

func generateRawTxNaive(t *testing.T) string{

	client, err := ethclient.Dial("http://127.0.0.1:7545") //Ganache local address
	if err != nil {
		log.Fatal(err)
	}

	privateKey, _ := crypto.HexToECDSA(priv)
	recipientAddr := common.HexToAddress(receiver)
	amount := big.NewInt(1000000000000000000) // 1 ether
	gasLimit := uint64(100000)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Error("error casting public key To ECDSA")
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

	t.Log("naive raw tx hex:",rawTxHex)
	return rawTxHex
}

func generateRawTxDeployContract(t *testing.T, gas uint64, gasPrice int64, nonce uint64, data []byte) string{

	privateKey, _ := crypto.HexToECDSA(priv)
	amount := big.NewInt(0) // 1 ether

	tx := types.NewContractCreation(nonce, amount, gas, big.NewInt(gasPrice), data)

	signer := types.NewEIP155Signer(chainID)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)

	t.Log("raw tx 2 hex:",rawTxHex)
	return rawTxHex
}

func generateRawTxSetContract(t *testing.T, contractAddr string, gas uint64, gasPrice int64, nonce uint64, data []byte) string{
	recipientAddr := common.HexToAddress(contractAddr)

	chainID := big.NewInt(5777) // ganache

	privateKey, _ := crypto.HexToECDSA(priv)
	amount := big.NewInt(0) // 1 ether

	tx := types.NewTransaction(nonce, recipientAddr, amount, gas, big.NewInt(gasPrice), data)

	signer := types.NewEIP155Signer(chainID)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)

	t.Log("raw tx 2 hex:",rawTxHex)
	return rawTxHex
}

func deployContract(t *testing.T) *string{
	h := NewETHHandler()
	m:= &MsgToChain{
		From:    addr,
		User:    "test1",
		Version: "1",
		ChainID: 123,
		Typ:     "request_tx",
		RequestTx:  	&MsgContractFunc{
			Function: "",
			Inputs:   []string{"kingdom","xrisheng","100","1200"},},}

	h.ToChains <- m
	for{
		select {
		case msg := <-h.FromChains:
			if msg.TxSent != nil {
				t.Log("deploy tx Sent with hash:", msg.TxSent.TxHash)
				t.Log("deploy tx Sent with gas price:", msg.TxSent.GasPrice)
				t.Log("deploy tx Sent with Nonce:", msg.TxSent.Nonce)
				t.Log("deploy tx Sent with estimated gas:", msg.TxSent.GasEstimated)
			}else if msg.TxInfo != nil{
				if msg.TxInfo.Function != m.RequestTx.Function {
					t.Error("Function name not equal")
				}
				if msg.TxInfo.GasPrice <= 0{
					t.Error("gas price not greater than 0")
				}
				if msg.TxInfo.Nonce <= 0{
					t.Error("Nonce not greater than 0")
				}
				if msg.TxInfo.Data == nil{
					t.Error("Data field is nil")
				}
				rawTxData2 := generateRawTxDeployContract(t,5000000,msg.TxInfo.GasPrice,msg.TxInfo.Nonce,msg.TxInfo.Data)

				h.ToChains <- &MsgToChain{
					From:     addr,
					User:     "test1",
					Version:  "1",
					ChainID:  123,
					Typ:      "signed_tx",
					SignedTx: &rawTxData2,}

			} else if msg.TxReceipt != nil{
				if msg.TxReceipt.ContractAddr == nil{
					t.Error("Contract Address is nil, deploy failed")
				}
				t.Log("deploy tx Confirmed with hash:", msg.TxReceipt.TxHash)
				t.Log("deploy tx Confirmed with gas used:", msg.TxReceipt.GasUsed)
				fmt.Printf("contract deployed at address: %s\n", *msg.TxReceipt.ContractAddr)
				return msg.TxReceipt.ContractAddr
			}
		}
	}
}

func setContract(t *testing.T, contractAddr *string) {
	h := NewETHHandler()
	m := &MsgToChain{
		From:    addr,
		User:    "test1",
		Version: "1",
		ChainID: 123,
		Typ:     "request_tx",
		RequestTx: &MsgContractFunc{
			Function: "setEntryCost",
			Inputs:   []string{"500"},},}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.TxSent != nil {
				t.Log("set tx Sent with hash:", msg.TxSent.TxHash)
				t.Log("set tx Sent with gas price:", msg.TxSent.GasPrice)
				t.Log("set tx Sent with nonce:", msg.TxSent.Nonce)
				t.Log("set tx Sent with estimated gas:", msg.TxSent.GasEstimated)
			} else if msg.TxInfo != nil {
				rawTxData3 := generateRawTxSetContract(t,*contractAddr, 500000, msg.TxInfo.GasPrice, msg.TxInfo.Nonce, msg.TxInfo.Data)

				h.ToChains <- &MsgToChain{
					From:     addr,
					User:     "test1",
					Version:  "1",
					ChainID:  123,
					Typ:      "signed_tx",
					SignedTx: &rawTxData3,}
			}else if msg.TxReceipt != nil{
				fmt.Printf("set tx Confirmed with hash: %s\n", msg.TxReceipt.TxHash)
				t.Log("set tx Confirmed with gas used:", msg.TxReceipt.GasUsed)
				return
			}

		}
	}
}

func callContract(t *testing.T, contractAddr *string) {
	h := NewETHHandler()

	m := &MsgToChain{
		From:    addr,
		User:    "test1",
		Version: "1",
		ChainID: 123,
		Typ:     "contract_call",
		Call: &MsgCall{
			ContractAddr: *contractAddr,
			ContractFunc: MsgContractFunc{
				Function: "getEntryCost",},
		},}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.CallReturn != nil {
				t.Log("call return From address:", msg.CallReturn.ContractAddr)
				t.Log("call return of function: ", msg.CallReturn.Function)
				fmt.Printf("call return with output: %s\n", msg.CallReturn.Output)
				return
			}
		}
	}
}

func TestSendNaiveTx(t *testing.T) {
	//----------------------Test naive transaction and polling--------------------------------------------------------//
	h := NewETHHandler()
	rawTxData1 :=generateRawTxNaive(t)

	h.ToChains <- &MsgToChain{
		From:     addr,
		User:     "test1",
		Version:  "1",
		ChainID:  123,
		Typ:      "signed_tx",
		SignedTx: &rawTxData1,}
	for {
		select {
		case msg := <-h.FromChains:

			if msg.TxSent != nil {
				t.Log("naive tx Sent with hash:", msg.TxSent.TxHash)
				t.Log("naive tx Sent with gas price:", msg.TxSent.GasPrice)
				t.Log("naive tx Sent with Nonce:", msg.TxSent.Nonce)
				fmt.Printf("naive tx Sent with estimated gas: %d\n", msg.TxSent.GasEstimated)
			} else if msg.TxReceipt != nil {
				fmt.Printf("naive tx Confirmed with hash: %s\n", msg.TxReceipt.TxHash)
				t.Log("naive tx Confirmed with gas cost:", msg.TxReceipt.GasUsed)
				return
			}

		}
	}
}

func TestSmartContract(t *testing.T){
	contractAddr :=deployContract(t)
	setContract(t,contractAddr)
	callContract(t,contractAddr)
	return
}