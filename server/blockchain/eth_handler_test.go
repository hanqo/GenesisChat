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

func generateRawTxNaive() string {

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
		log.Fatal("error casting public key To ECDSA")
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

	fmt.Printf("raw tx hex: %s\n", rawTxHex)
	return rawTxHex
}

func generateRawTxDeployContract(gas uint64, gasPrice int64, nonce uint64, data []byte) string {

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

	fmt.Printf("raw tx 2 hex: %s\n", rawTxHex)
	return rawTxHex
}

func generateRawTxSetContract(gas uint64, gasPrice int64, nonce uint64, data []byte) string {
	contractAddr := "0xD083B809Ee1531aA40bb94cbb1F134913aC9d5eE"
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

	fmt.Printf("raw tx 2 hex: %s\n", rawTxHex)
	return rawTxHex
}

//Cannot use name TestforwardRawTx, although the Function name is forwardRawTx.
func TestSendSignedTx(t *testing.T) {
	//----------------------Test naive transaction and polling--------------------------------------------------------//
	h := NewETHHandler()
	rawTxData1 := generateRawTxNaive()

	h.ToChains <- &MsgToChain{
		From:     addr,
		User:     "test1",
		Version:  "1",
		ChainID:  123,
		Typ:      "signed_tx",
		SignedTx: &rawTxData1}
	for {
		select {
		case msg := <-h.FromChains:

			if msg.TxSent != nil {
				fmt.Printf("naive tx Sent with hash: %s\n", msg.TxSent.TxHash)
				fmt.Printf("naive tx Sent with gas price: %d\n", msg.TxSent.GasPrice)
				fmt.Printf("naive tx Sent with Nonce: %d\n", msg.TxSent.Nonce)
				fmt.Printf("naive tx Sent with estimated gas: %d\n", msg.TxSent.GasEstimated)
			} else if msg.TxReceipt != nil {
				fmt.Printf("naive tx Confirmed with hash: %s\n", msg.TxReceipt.TxHash)
				fmt.Printf("naive tx Confirmed with gas cost: %d\n", msg.TxReceipt.GasUsed)
				return
			}

		}
	}
}

func TestDeployContract(t *testing.T) {
	h := NewETHHandler()
	m := &MsgToChain{
		From:    addr,
		User:    "test1",
		Version: "1",
		ChainID: 123,
		Typ:     "request_tx",
		RequestTx: &MsgContractFunc{
			Function: "",
			Inputs:   []string{"kingdom", "xrisheng", "100", "1200"}}}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.TxSent != nil {
				fmt.Printf("deploy tx Sent with hash: %s\n", msg.TxSent.TxHash)
				fmt.Printf("deploy tx Sent with gas price: %d\n", msg.TxSent.GasPrice)
				fmt.Printf("deploy tx Sent with Nonce: %d\n", msg.TxSent.Nonce)
				fmt.Printf("deploy tx Sent with estimated gas: %d\n", msg.TxSent.GasEstimated)
			} else if msg.TxInfo != nil {
				if msg.TxInfo.Function != m.RequestTx.Function {
					t.Error("Function name not equal")
				}
				if msg.TxInfo.GasPrice <= 0 {
					t.Error("gas price not greater than 0")
				}
				if msg.TxInfo.Nonce <= 0 {
					t.Error("Nonce not greater than 0")
				}
				if msg.TxInfo.Data == nil {
					t.Error("Data field is nil")
				}
				rawTxData2 := generateRawTxDeployContract(5000000, msg.TxInfo.GasPrice, msg.TxInfo.Nonce, msg.TxInfo.Data)

				h.ToChains <- &MsgToChain{
					From:     addr,
					User:     "test1",
					Version:  "1",
					ChainID:  123,
					Typ:      "signed_tx",
					SignedTx: &rawTxData2}

			} else if msg.TxReceipt != nil {
				if msg.TxReceipt.ContractAddr == nil {
					t.Error("Contract Address is nil, deploy failed")
				}
				fmt.Printf("deploy tx Confirmed with hash: %s\n", msg.TxReceipt.TxHash)
				fmt.Printf("deploy tx Confirmed with gas used: %d\n", msg.TxReceipt.GasUsed)
				fmt.Printf("contract deployed at address %s\n", *msg.TxReceipt.ContractAddr)
				return
			}
		}
	}
}

func TestSetContract(t *testing.T) {
	h := NewETHHandler()
	m := &MsgToChain{
		From:    addr,
		User:    "test1",
		Version: "1",
		ChainID: 123,
		Typ:     "request_tx",
		RequestTx: &MsgContractFunc{
			Function: "setEntryCost",
			Inputs:   []string{"500"}}}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.TxSent != nil {
				fmt.Printf("set tx Sent with hash: %s\n", msg.TxSent.TxHash)
				fmt.Printf("set tx Sent with gas price: %d\n", msg.TxSent.GasPrice)
				fmt.Printf("set tx Sent with nonce: %d\n", msg.TxSent.Nonce)
				fmt.Printf("set tx Sent with estimated gas: %d\n", msg.TxSent.GasEstimated)
			} else if msg.TxInfo != nil {
				rawTxData3 := generateRawTxSetContract(500000, msg.TxInfo.GasPrice, msg.TxInfo.Nonce, msg.TxInfo.Data)

				h.ToChains <- &MsgToChain{
					From:     addr,
					User:     "test1",
					Version:  "1",
					ChainID:  123,
					Typ:      "signed_tx",
					SignedTx: &rawTxData3}
			} else if msg.TxReceipt != nil {
				fmt.Printf("set tx Confirmed with hash: %s\n", msg.TxReceipt.TxHash)
				fmt.Printf("set tx Confirmed with gas used: %d\n", msg.TxReceipt.GasUsed)
				return
			}

		}
	}
}

func TestCallContract(t *testing.T) {
	h := NewETHHandler()

	m := &MsgToChain{
		From:    addr,
		User:    "test1",
		Version: "1",
		ChainID: 123,
		Typ:     "contract_call",
		Call: &MsgCall{
			ContractAddr: "0xD083B809Ee1531aA40bb94cbb1F134913aC9d5eE",
			ContractFunc: MsgContractFunc{
				Function: "getEntryCost"},
		}}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.CallReturn != nil {
				fmt.Printf("Call return From address: %s\n", msg.CallReturn.ContractAddr)
				fmt.Printf("Call return of Function: %s\n", msg.CallReturn.Function)
				fmt.Printf("Call return with Output: %s\n", msg.CallReturn.Output)
				return
			}
		}
	}
}
