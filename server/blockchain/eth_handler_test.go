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
	"time"
	"github.com/AfterworkBlockchain/GenesisChat/server/vote"
)
const addr = "0xb66785f087B0A100c39c39B801104D22086FF1bE"
const priv = "E9A6D816389523F51B7CB44EB16CD661050F6F85B4268D452E4745B74619F1D2"
const receiver = "0xb66785f087B0A100c39c39B801104D22086FF1bE"

var chainID = big.NewInt(3) //ropsten

func generateRawTxNaive(t *testing.T) string{

	client, err := ethclient.Dial(ethRPCAddr) //Ganache local address
	if err != nil {
		log.Fatal(err)
	}

	privateKey, _ := crypto.HexToECDSA(priv)
	recipientAddr := common.HexToAddress(receiver)
	amount := big.NewInt(1000000000000 ) // 0.01ether
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

	chainID := big.NewInt(3) // ropsten

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
		ChainID: 3,
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
					ChainID:  3,
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

func voteProcess(t *testing.T, contractAddr *string, funcName *string, voteNonce *string) *vote.MsgVoteResult{
	v:= vote.NewVoteHandler()
	v.ToVote <- &vote.MsgToVote{
		Owner:"test_owner1",
		Topic:"test_topic1",
		Typ:"new_vote",
		NewVote:&vote.MsgNewVote{
			Proposal: &vote.MsgVoteProposal{"contract", nil,contractAddr,funcName, voteNonce},
			Duration:10,
			PassRate:33,
			VoterList:[]string{"voter1","voter2","voter3","voter4","voter5"},},}

	time.Sleep(time.Second * 1)

	//Vote starts
	v.ToVote <- &vote.MsgToVote{
		Owner:"voter1",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &vote.MsgBallot{0},}
	v.ToVote <- &vote.MsgToVote{
		Owner:"voter2",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &vote.MsgBallot{2},}
	v.ToVote <- &vote.MsgToVote{
		Owner:"voter3",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &vote.MsgBallot{1},}
	v.ToVote <- &vote.MsgToVote{
		Owner:"voter4",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &vote.MsgBallot{1},}
	time.Sleep(time.Second * 1)

	//Request status of vote
	v.ToVote <- &vote.MsgToVote{
		Owner:"voter4",
		Topic:"test_topic1",
		Typ:"status",}

	//Request paramter of vote
	v.ToVote <- &vote.MsgToVote{
		Owner:"voter5",
		Topic:"test_topic1",
		Typ:"parameter",}

	time.Sleep(time.Second * 1)

	for{
		select {
		case msg := <-v.FromVote:
			if msg.Typ == "status"{
				status := msg.Status
				if len(status.ForList) != 2||
					len(status.AgainstList) != 1 ||
					len(status.AbstainedList) != 1 {
					t.Error("The status of voting is not correct")
				}

			}else if msg.Typ== "result"{
				fmt.Printf("result for topic %s is %v\n", msg.Topic, msg.Result.Value)
				if msg.Topic != "test_topic1" ||
					msg.Result.Value != true {
					t.Error("Vote Result not correct")
				}
				return  msg.Result

			}else if msg.Typ== "parameter"{
				param:= msg.Param

				if param.PassRate!=33||param.Duration!= 10||param.VoterSize!= 5{
					t.Error("Vote Paramter not correct")
				}
			}
		}
	}
}

func setContract(t *testing.T, contractAddr *string) {
	h := NewETHHandler()
	funcName := "setCost"
	voteNonce := "1"
	res := voteProcess(t,contractAddr, &funcName, &voteNonce)

	if res.Value != true{
		t.Error("Proposal is not passed in vote")
	}

	input := []string{"500","600"}

	input = append(input, *res.Proposal.Nonce, *res.Signature)
	m := &MsgToChain{
		From:    addr,
		User:    "test1",
		Version: "1",
		ChainID: 3,
		Typ:     "request_tx",
		RequestTx: &MsgContractFunc{
			Function: funcName,
			Inputs:   input,},}

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
					ChainID:  3,
					Typ:      "signed_tx",
					SignedTx: &rawTxData3,}
			}else if msg.TxReceipt != nil{
				fmt.Printf("set tx Confirmed with hash: %s\n", msg.TxReceipt.TxHash)
				t.Log("set tx confirmed with gas used:", msg.TxReceipt.GasUsed)
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
		ChainID: 3,
		Typ:     "contract_call",
		Call: &MsgCall{
			ContractAddr: *contractAddr,
			ContractFunc: MsgContractFunc{
				Function: "getCost",},
		},}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.CallReturn != nil {
				t.Log("call return From address:", msg.CallReturn.ContractAddr)
				t.Log("call return of function: ", msg.CallReturn.Function)
				fmt.Printf("call return with output: %v\n", msg.CallReturn.Output)
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
		ChainID:  3,
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