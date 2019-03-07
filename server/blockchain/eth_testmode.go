package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/tinode/chat/server/vote"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"time"
)

const addrTestMode = "0xb66785f087B0A100c39c39B801104D22086FF1bE"
const privTestMode = "E9A6D816389523F51B7CB44EB16CD661050F6F85B4268D452E4745B74619F1D2"

var chainID = big.NewInt(3) //ropsten

func generateRawTxDeployContractTestMode(gas uint64, gasPrice int64, nonce uint64, data []byte) string {

	privateKey, _ := crypto.HexToECDSA(privTestMode)
	amount := big.NewInt(0) // 1 ether

	tx := types.NewContractCreation(nonce, amount, gas, big.NewInt(gasPrice), data)

	signer := types.NewEIP155Signer(chainID)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return rawTxHex
}

func generateRawTxSetContractTestMode(contractAddr string, gas uint64, gasPrice int64, nonce uint64, data []byte) string {
	recipientAddr := common.HexToAddress(contractAddr)

	chainID := big.NewInt(3) // ropsten

	privateKey, _ := crypto.HexToECDSA(privTestMode)
	amount := big.NewInt(0) // 1 ether

	tx := types.NewTransaction(nonce, recipientAddr, amount, gas, big.NewInt(gasPrice), data)

	signer := types.NewEIP155Signer(chainID)
	signedTx, _ := types.SignTx(tx, signer, privateKey)

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	rawTxHex := hex.EncodeToString(rawTxBytes)
	return rawTxHex
}

func DeployContractTestMode() (*string, *string) {
	h := NewETHHandler()
	m := &MsgToChain{
		From:    addrTestMode,
		User:    "test1",
		Version: "1",
		ChainID: 3,
		Typ:     "request_tx",
		RequestTx: &MsgContractFunc{
			Function: "",
			Inputs:   []string{"genesis001", "admin_ggg", "100", "20"}}}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.TxSent != nil {

			} else if msg.TxInfo != nil {

				rawTxData2 := generateRawTxDeployContractTestMode(5000000, msg.TxInfo.GasPrice, msg.TxInfo.Nonce, msg.TxInfo.Data)

				h.ToChains <- &MsgToChain{
					From:     addrTestMode,
					User:     "test1",
					Version:  "1",
					ChainID:  3,
					Typ:      "signed_tx",
					SignedTx: &rawTxData2}

			} else if msg.TxReceipt != nil {
				fmt.Printf("contract deployed at addrTestModeess: %s\n", *msg.TxReceipt.ContractAddr)
				return &msg.TxReceipt.TxHash, msg.TxReceipt.ContractAddr
			}
		}
	}
}

func VoteProcessTestMode(contractAddr *string, funcName *string, voteNonce *string) *vote.MsgVoteResult {
	v := vote.NewVoteHandler()
	v.ToVote <- &vote.MsgToVote{
		Owner: "test_owner1",
		Topic: "test_topic1",
		Typ:   "new_vote",
		NewVote: &vote.MsgNewVote{
			Proposal:  &vote.MsgVoteProposal{"contract", nil, contractAddr, funcName, voteNonce},
			Duration:  20,
			PassRate:  33,
			VoterList: []string{"voter1", "voter2", "voter3", "voter4", "voter5"}}}

	time.Sleep(time.Second * 1)

	//Vote starts
	v.ToVote <- &vote.MsgToVote{
		Owner:  "voter1",
		Topic:  "test_topic1",
		Typ:    "vote",
		Ballot: &vote.MsgBallot{0}}
	v.ToVote <- &vote.MsgToVote{
		Owner:  "voter2",
		Topic:  "test_topic1",
		Typ:    "vote",
		Ballot: &vote.MsgBallot{2}}
	v.ToVote <- &vote.MsgToVote{
		Owner:  "voter3",
		Topic:  "test_topic1",
		Typ:    "vote",
		Ballot: &vote.MsgBallot{1}}
	v.ToVote <- &vote.MsgToVote{
		Owner:  "voter4",
		Topic:  "test_topic1",
		Typ:    "vote",
		Ballot: &vote.MsgBallot{1}}
	time.Sleep(time.Second * 1)

	//Request status of vote
	v.ToVote <- &vote.MsgToVote{
		Owner: "voter4",
		Topic: "test_topic1",
		Typ:   "status"}

	//Request paramter of vote
	v.ToVote <- &vote.MsgToVote{
		Owner: "voter5",
		Topic: "test_topic1",
		Typ:   "parameter"}

	time.Sleep(time.Second * 1)

	for {
		select {
		case msg := <-v.FromVote:
			if msg.Typ == "result" {
				fmt.Printf("result for topic %s is %v\n", msg.Topic, msg.Result.Value)
				if msg.Topic != "test_topic1" ||
					msg.Result.Value != true {
				}
				return msg.Result

			}
		}
	}
}

func SetContractTestMode(contractAddr *string) *string{
	h := NewETHHandler()
	funcName := "setCost"
	voteNonce := "1"
	res := VoteProcessTestMode(contractAddr, &funcName, &voteNonce)

	input := []string{"500", "20"}

	input = append(input, *res.Proposal.Nonce, *res.Signature)
	m := &MsgToChain{
		From:    addrTestMode,
		User:    "test1",
		Version: "1",
		ChainID: 3,
		Typ:     "request_tx",
		RequestTx: &MsgContractFunc{
			Function: funcName,
			Inputs:   input}}

	h.ToChains <- m

	for {
		select {
		case msg := <-h.FromChains:
			if msg.TxInfo != nil {
				rawTxData3 := generateRawTxSetContractTestMode(*contractAddr, 500000, msg.TxInfo.GasPrice, msg.TxInfo.Nonce, msg.TxInfo.Data)

				h.ToChains <- &MsgToChain{
					From:     addrTestMode,
					User:     "test1",
					Version:  "1",
					ChainID:  3,
					Typ:      "signed_tx",
					SignedTx: &rawTxData3}
			} else if msg.TxReceipt != nil {
				fmt.Printf("set tx Confirmed with hash: %s\n", msg.TxReceipt.TxHash)
				return &msg.TxReceipt.TxHash
			}
		}
	}
}

func CallContractTestMode(contractAddr *string) *MsgCallReturn {
	h := NewETHHandler()

	m := &MsgToChain{
		From:    addrTestMode,
		User:    "test1",
		Version: "1",
		ChainID: 3,
		Typ:     "contract_call",
		Call: &MsgCall{
			ContractAddr: *contractAddr,
			ContractFunc: MsgContractFunc{
				Function: "getCost"},
		}}

	h.ToChains <- m
	for {
		select {
		case msg := <-h.FromChains:
			if msg.CallReturn != nil {
				fmt.Printf("call return with output: %v\n", msg.CallReturn.Output)
				return msg.CallReturn
			}
		}
	}
}

func TestSmartContracttTestMode() *MsgCallReturn {
	_, contractAddr := DeployContractTestMode()

	SetContractTestMode(contractAddr)
	res:= CallContractTestMode(contractAddr)
	return res
}
