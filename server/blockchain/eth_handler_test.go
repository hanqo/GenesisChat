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

func generateRawTx(client *ethclient.Client, priv string, to string) string{
	chainID := big.NewInt(5777) // ganache

	privateKey, _ := crypto.HexToECDSA(priv)
	recipientAddr := common.HexToAddress(to)
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

	fmt.Printf(rawTxHex)
	return rawTxHex
}

func TestForwardRawTx(t *testing.T) {
	client, err := ethclient.Dial("http://127.0.0.1:7545") //Ganache local address
	priv := "4b62386099abd28f2b63d3a08918cbffc72f4752e3a029747f2a4681b28021c7"
	addr := "0xb04b61254B42d64f17938E5DCe2eb728cAfF8937"
	to := "0x218778aA387BCCD5167B6881B4Fc210f0ebFe5Ae"
	h := NewHandler()
	if err != nil {
		log.Fatal(err)
	}
	rawTxData :=generateRawTx(client,priv,to)
	h.ForwardRawTx(&MsgToChain{
		from :     addr,
		user :     "test",
		version :  "test",
		chainID :  123,
		rawTx:     &MsgRawTx{
			       rawTxData:rawTxData,
			       },
	             })
	select {
	case msg:= <-h.fromChains:
		if msg.txReceipt.confirmed != false {
			t.Error("Transaction not confirmed")
		} else if msg.txReceipt.confirmed == true{
		t.Log("Transaction confirmed")
		}
	}

	return
}