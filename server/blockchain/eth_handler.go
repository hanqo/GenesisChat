package blockchain

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/onrik/ethrpc"
	"log"
	"time"

)

//TODO(xrisheng):move config to config file.
var ethRPCAddr = "http://127.0.0.1:7545"
var waitPeriod time.Duration = 20 //second

// messages and the wrapper send to block chain.
type MsgToChain struct {

	from 		string
	user  		string
	version 	string
	chainID 	int32

	rawTx 		*MsgRawTx
}

type MsgRawTx struct{
	rawTxData 	string
}

// messages and the wrapper send from block chain
type MsgFromChain struct{
	to 			string
	user  	    string
	version 	string
	chainID 	int32

	txReceipt   *MsgTxReceipt
}

type MsgTxReceipt struct{

	confirmed	bool
	txBlockNr	int
	txBlockHash string
	txHash 		string
	txIndex     int
	gasUsed     int
}

type Handler struct {

	toChains chan *MsgToChain
	fromChains chan *MsgFromChain
	done chan bool

}

func NewHandler() * Handler{

	h := &Handler{
		toChains:make(chan *MsgToChain,1),
		fromChains:make(chan *MsgFromChain,1),
		done:make(chan bool, 1),
	}
	go h.run()
	return h
}

func (h *Handler) run(){
	for{
		select {
			case msg := <- h.toChains:

				switch {
				case  msg.rawTx !=  nil:
					go h.forwardRawTx(msg)
				}
		    case <-h.done:
			 	log.Printf("Stop ETH hanlder")
			 	return
		}
	}
}

func (h *Handler) forwardRawTx(r *MsgToChain){

	client := ethrpc.New(ethRPCAddr)

	balance,err := client.EthGetBalance(r.from, "latest")
	if err != nil{
		log.Fatal(err)
	}

	rawTxBytes, _ := hex.DecodeString(r.rawTx.rawTxData)
	tx := new(types.Transaction)
	_ = rlp.DecodeBytes(rawTxBytes, &tx)

	if balance.Cmp(tx.Value()) < 1{
		log.Fatal(err)
	}

	txHash, err := client.EthSendRawTransaction(r.rawTx.rawTxData)

	if err != nil{
		log.Fatal(err)
	}

	//TODO(xrisheng): Parameterization?
	for i := 1; i <= 10000; i++ {
		time.Sleep(waitPeriod * time.Second)
		receipt, err := client.EthGetTransactionReceipt(txHash)

		if err != nil {
			log.Fatal(err)
		}

		if receipt != nil {
			h.fromChains <- &MsgFromChain{
				to:      r.from,
				user:    r.user,
				version: r.version,
				chainID: r.chainID,

				txReceipt: &MsgTxReceipt{
					confirmed: 	 true,
					txHash:      receipt.TransactionHash,
					txIndex:     receipt.TransactionIndex,
					txBlockNr:   receipt.BlockNumber,
					txBlockHash: receipt.BlockHash,
					gasUsed:     receipt.GasUsed,
				},
			}
			return
		}
	}
		h.fromChains <- &MsgFromChain{
			to:      r.from,
			user:    r.user,
			version: r.version,
			chainID: r.chainID,

			txReceipt: &MsgTxReceipt{
				confirmed: false,
				},
			}
}
