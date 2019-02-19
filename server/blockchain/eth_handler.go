package blockchain

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
	"math/big"
	"time"
)

//TODO(xrisheng):move config To config file.
//TODO(xrisheng):eth event handler
//var ethRPCAddr = "https://rinkeby.infura.io"
var ethRPCAddr = "http://127.0.0.1:7545"

type ETHHandler struct {
	ToChains   chan *MsgToChain
	FromChains chan *MsgFromChain
	RunDone    chan bool
	PollDone   chan bool

	abi *ABIHandler

	ticker   *time.Ticker
	add      chan TxPending
	pendings []TxPending
}

func NewETHHandler() *ETHHandler {

	h := &ETHHandler{
		ToChains:   make(chan *MsgToChain, 1),
		FromChains: make(chan *MsgFromChain, 1),
		RunDone:    make(chan bool, 1),
		PollDone:   make(chan bool, 1),

		abi: NewABIHandler(),

		ticker: time.NewTicker(time.Second * 1),
		add:    make(chan TxPending, 1),
	}
	go h.run()
	go h.poll()
	return h
}

func (h *ETHHandler) poll() {
	c, err := ethclient.Dial(ethRPCAddr)

	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-h.ticker.C:
			for i := len(h.pendings) - 1; i >= 0; i-- {

				hash := common.HexToHash(h.pendings[i].TxHash)
				ret, _ := c.TransactionReceipt(context.Background(), hash)

				if ret != nil {
					m := &MsgFromChain{
						To:      h.pendings[i].From,
						User:    h.pendings[i].User,
						Version: h.pendings[i].Version,
						ChainID: h.pendings[i].ChainID,
						Typ:     "tx_receipt",

						TxReceipt: &MsgTxReceipt{
							Confirmed: true,
							TxHash:    h.pendings[i].TxHash,
							GasUsed:   ret.GasUsed,
						},}
					s := ret.ContractAddress.String()
					if len(s) > 0 {
						m.TxReceipt.ContractAddr = &s
					}
					h.FromChains <- m
					h.pendings = append(h.pendings[:i], h.pendings[i+1:]...)
				}
			}
		case s := <-h.add:
			h.pendings = append(h.pendings, s)
		case <-h.PollDone:
			log.Printf("Stop ETH handler polling")
			return
		}
	}
}

func (h *ETHHandler) run() {
	for {
		select {
		case msg := <-h.ToChains:
			switch msg.Typ {
			case "signed_tx":
				go h.sendSignedTx(msg)
			case "request_tx":
				go h.generateTxInfo(msg)
			case "contract_call":
				go h.callContract(msg)
			default:
				log.Printf("Unrecognized message in eth run loop")
			}
		case <-h.RunDone:
			log.Printf("Stop ETH hanlder running")
			return
		}
	}
}

func (h *ETHHandler) sendSignedTx(r *MsgToChain) {

	if r.SignedTx == nil {
		log.Fatal("signed is empty")
	}

	c, err := ethclient.Dial(ethRPCAddr)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	rawTxBytes, err := hex.DecodeString(*r.SignedTx)
	if err != nil {
		log.Fatal(err)
	}

	tx := new(types.Transaction)
	_ = rlp.DecodeBytes(rawTxBytes, &tx)

	from := common.HexToAddress(r.From)
	balance, err := c.BalanceAt(context.Background(),
		from,
		nil)
	if err != nil {
		log.Fatal(err)
	}

	if balance.Cmp(tx.Value()) < 1 {
		str := fmt.Sprintf("Not enough balance in %v", r.From)
		log.Fatal(str)
	}

	msg := ethereum.CallMsg{
		From:     from,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Data:     tx.Data(),
		Value:    tx.Value(),
	}
	est, err := c.EstimateGas(context.Background(), msg)
	//estimate := uint64(21000)

	err = c.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatal(err)
	}

	h.add <- TxPending{
		From:    r.From,
		User:    r.User,
		Version: r.Version,
		ChainID: r.ChainID,
		TxHash:  tx.Hash().String(),
	}

	h.FromChains <- &MsgFromChain{
		To:      r.From,
		User:    r.User,
		Version: r.Version,
		ChainID: r.ChainID,
		Typ:     "tx_sent",
		TxSent: &MsgTxSent{
			TxHash:       tx.Hash().String(),
			GasPrice:     tx.GasPrice().Int64(),
			Nonce:        tx.Nonce(),
			gasEstimated: est,
		},}
}

func (h *ETHHandler) generateTxInfo(r *MsgToChain) {
	if r.RequestTx == nil {
		log.Fatal("request is empty")
	}

	c, err := ethclient.Dial(ethRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	from := common.HexToAddress(r.From)

	nonce, _ := c.PendingNonceAt(context.Background(), from)
	price, _ := c.SuggestGasPrice(context.Background())

	var txData []byte = nil

	if r.RequestTx != nil {
		txData = h.abi.packContractFunc(r.RequestTx.Function, r.RequestTx.Inputs)
	}

	h.FromChains <- &MsgFromChain{
		To:      r.From,
		User:    r.User,
		Version: r.Version,
		ChainID: r.ChainID,
		Typ:     "tx_info",

		TxInfo: &MsgTxInfo{
			Function: r.RequestTx.Function,
			GasPrice: price.Int64(),
			Nonce:    uint64(nonce),
			Data:     txData,
			GasLimit: uint64(1000000),
		},}
}

func (h *ETHHandler) callContract(r *MsgToChain) {
	if r.Call == nil {
		log.Fatal("Call is empty")
	}

	c, err := ethclient.Dial(ethRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	price, _ := c.SuggestGasPrice(context.Background())
	data := h.abi.packContractFunc(r.Call.ContractFunc.Function, r.Call.ContractFunc.Inputs)
	to := common.HexToAddress(r.Call.ContractAddr)

	msg := ethereum.CallMsg{
		From:     common.HexToAddress(r.From),
		To:       &to,
		Gas:      uint64(1000000),
		GasPrice: price,
		Data:     data,
	}

	if r.Call.Value != nil {
		msg.Value = big.NewInt(*r.Call.Value)
	}

	returnByte, err := c.CallContract(context.Background(), msg, nil)

	if err != nil {
		log.Fatal(err)
	}
	retStr, err := h.abi.unpackContractFunc(returnByte, r.Call.ContractFunc.Function)

	h.FromChains <- &MsgFromChain{
		To:      r.From,
		User:    r.User,
		Version: r.Version,
		ChainID: r.ChainID,
		Typ:     "call_return",

		CallReturn: &MsgCallReturn{
			Function:     r.Call.ContractFunc.Function,
			ContractAddr: r.Call.ContractAddr,
			Output:       retStr,},
	}
}

/*func (h *ETHHandler) requestReceipt(r *MsgToChain){
	if r.TxHash ==nil{
		log.Fatal("tx hash is empty")
	}

	c, err := ethclient.Dial(ethRPCAddr)
	if err != nil{
		log.Fatal(err)
	}

	toByte := []byte(*r.TxHash)
	hash := common.BytesToHash(toByte)
	ret,err:= c.TransactionReceipt(context.Background(),hash)
	if err != nil{
		log.Fatal(err)
	}
	if ret == nil{
		h.FromChains <- &MsgFromChain{
			To:      r.From,
			User:    r.User,
			Version: r.Version,
			ChainID: r.ChainID,
			Typ:	 "tx_receipt",

			receipt:&MsgTxReceipt{
				Confirmed:	false,
			},

		}
		return
	}

	h.FromChains <- &MsgFromChain{
		To:      r.From,
		User:    r.User,
		Version: r.Version,
		ChainID: r.ChainID,
		Typ:	 "tx_receipt",

		receipt:&MsgTxReceipt{
			Confirmed:	true,
			TxHash:		string(toByte),
			GasUsed: 	ret.GasUsed,
		},

	}
	return
}*/
