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

//TODO(xrisheng):move config to config file.
//var ethRPCAddr = "https://rinkeby.infura.io"
var ethRPCAddr = "http://127.0.0.1:7545"

type ETHHandler struct {

	toChains      chan *MsgToChain
	fromChains    chan *MsgFromChain
	runDone       chan bool
	pollDone      chan bool

	abi           *ABIHandler

	ticker       *time.Ticker
	add		     chan TxPending
	pendings     []TxPending
}

func NewETHHandler() *ETHHandler {

	h := &ETHHandler{
		toChains:		make(chan *MsgToChain,1),
		fromChains:		make(chan *MsgFromChain,1),
		runDone:		make(chan bool, 1),
		pollDone:		make(chan bool, 1),

		abi:            NewABIHandler(),

		ticker:			time.NewTicker(time.Second * 1),
		add:    		make(chan TxPending,1),
	}
	go h.run()
	go h.poll()
	return h
}

func (h *ETHHandler) poll(){
	c, err := ethclient.Dial(ethRPCAddr)

	if err != nil{
		log.Fatal(err)
	}

	for{
		select {
		case <-h.ticker.C:
			for i:= len(h.pendings)-1; i>=0; i--{

				hash := common.HexToHash(h.pendings[i].txHash)
				ret,_:= c.TransactionReceipt(context.Background(),hash)

				if ret != nil {
					m:= &MsgFromChain{
						to:      h.pendings[i].from,
						user:    h.pendings[i].user,
						version: h.pendings[i].version,
						chainID: h.pendings[i].chainID,
						typ:     "tx_receipt",

						txReceipt: &MsgTxReceipt{
							confirmed: 		true,
							txHash:    		h.pendings[i].txHash,
							gasUsed:   		ret.GasUsed,
						},}
					s:= ret.ContractAddress.String()
					if len(s) > 0{
						m.txReceipt.contractAddr = &s
					}
					h.fromChains <- m
					h.pendings = append(h.pendings[:i], h.pendings[i+1:]...)
				}
			}
		case s:= <-h.add:
			h.pendings = append(h.pendings,s)
		case <-h.pollDone:
			log.Printf("Stop ETH handler polling")
			return
		}
	}
}

func (h *ETHHandler) run(){
	for{
		select {
		case msg := <- h.toChains:
			switch msg.typ {
			case "signed_tx":
				go h.sendSignedTx(msg)
			case "request_tx":
				go h.generateTxInfo(msg)
			case "contract_call":
				go h.callContract(msg)
			default:
				log.Printf("Unrecognized message in run loop")
			}
		case <-h.runDone:
			log.Printf("Stop ETH hanlder running")
			return
		}
	}
}

func (h *ETHHandler) sendSignedTx(r *MsgToChain){

	if r.signedTx ==nil{
		log.Fatal("signed is empty")
	}

	c, err := ethclient.Dial(ethRPCAddr)
	if err != nil{
		log.Fatal(err)
	}

	defer c.Close()

	rawTxBytes, err := hex.DecodeString(*r.signedTx)
	if err != nil{
		log.Fatal(err)
	}

	tx := new(types.Transaction)
	_ = rlp.DecodeBytes(rawTxBytes, &tx)

	from := common.HexToAddress(r.from)
	balance, err := c.BalanceAt(context.Background(),
		from,
		nil)
	if err != nil{
		log.Fatal(err)
	}


	if balance.Cmp(tx.Value()) < 1 {
		str:= fmt.Sprintf("Not enough balance in %v", r.from)
		log.Fatal(str)
	}

	msg:= ethereum.CallMsg{
		From:from,
		To:tx.To(),
		Gas:tx.Gas(),
		GasPrice:tx.GasPrice(),
		Data:tx.Data(),
		Value:tx.Value(),
	}
	est,err := c.EstimateGas(context.Background(),msg)
	//estimate := uint64(21000)

	err = c.SendTransaction(context.Background(),tx)
	if err != nil{
		log.Fatal(err)
	}

	h.add <- TxPending{
		from:      	r.from,
		user:    	r.user,
		version: 	r.version,
		chainID: 	r.chainID,
		txHash:		tx.Hash().String(),
	}

	h.fromChains <- &MsgFromChain{
		to:      r.from,
		user:    r.user,
		version: r.version,
		chainID: r.chainID,
		typ:     "tx_sent",
		txSent: &MsgTxSent{
			txHash:          tx.Hash().String(),
			gasPrice:        tx.GasPrice().Int64(),
			nonce:           tx.Nonce(),
			gasEstimated:    est,
		},}
}

func (h *ETHHandler) generateTxInfo(r *MsgToChain){
	if r.requestTx ==nil{
		log.Fatal("request is empty")}

	c, err := ethclient.Dial(ethRPCAddr)
	if err != nil{
		log.Fatal(err)
	}
	defer c.Close()


	from := common.HexToAddress(r.from)

	nonce, _ := c.PendingNonceAt(context.Background(), from)
	price, _ := c.SuggestGasPrice(context.Background())

	var txData []byte = nil

	if r.requestTx != nil{
		txData = h.abi.packContractFunc(r.requestTx.function, r.requestTx.inputs)
	}

	h.fromChains <- &MsgFromChain{
		to:      r.from,
		user:    r.user,
		version: r.version,
		chainID: r.chainID,
		typ:     "tx_info",

		txInfo: &MsgTxInfo{
			function:  r.requestTx.function,
			gasPrice:  price.Int64(),
			nonce:     uint64(nonce),
			data:      txData,
			gasLimit:  uint64(1000000),
		},}
}

func (h *ETHHandler) callContract(r *MsgToChain) {
	if r.call == nil {
		log.Fatal("call is empty")
	}

	c, err := ethclient.Dial(ethRPCAddr)
	if err != nil{
		log.Fatal(err)
	}
	defer c.Close()

	price, _ := c.SuggestGasPrice(context.Background())
	data:= h.abi.packContractFunc(r.call.contractFunc.function, r.call.contractFunc.inputs)
	to:= common.HexToAddress(r.call.contractAddr)

	msg:= ethereum.CallMsg{
		From:common.HexToAddress(r.from),
		To: &to,
		Gas:uint64(1000000),
		GasPrice:price,
		Data:data,
	}

	if r.call.value != nil{
		msg.Value = big.NewInt(*r.call.value)
	}

	returnByte, err:= c.CallContract(context.Background(),msg,nil)

	if err != nil{
		log.Fatal(err)
	}
	retStr,err := h.abi.unpackContractFunc(returnByte,r.call.contractFunc.function)

	h.fromChains <- &MsgFromChain{
		to:      r.from,
		user:    r.user,
		version: r.version,
		chainID: r.chainID,
		typ:     "call_return",

		callReturn: &MsgCallReturn{
			function: r.call.contractFunc.function,
			contractAddr:r.call.contractAddr,
			output:   retStr,},
	}
	c.Close()
}

	/*func (h *ETHHandler) requestReceipt(r *MsgToChain){
		if r.txHash ==nil{
			log.Fatal("tx hash is empty")
		}

		c, err := ethclient.Dial(ethRPCAddr)
		if err != nil{
			log.Fatal(err)
		}

		toByte := []byte(*r.txHash)
		hash := common.BytesToHash(toByte)
		ret,err:= c.TransactionReceipt(context.Background(),hash)
		if err != nil{
			log.Fatal(err)
		}
		if ret == nil{
			h.fromChains <- &MsgFromChain{
				to:      r.from,
				user:    r.user,
				version: r.version,
				chainID: r.chainID,
				typ:	 "tx_receipt",

				receipt:&MsgTxReceipt{
					confirmed:	false,
				},

			}
			return
		}

		h.fromChains <- &MsgFromChain{
			to:      r.from,
			user:    r.user,
			version: r.version,
			chainID: r.chainID,
			typ:	 "tx_receipt",

			receipt:&MsgTxReceipt{
				confirmed:	true,
				txHash:		string(toByte),
				gasUsed: 	ret.GasUsed,
			},

		}
		return
	}*/