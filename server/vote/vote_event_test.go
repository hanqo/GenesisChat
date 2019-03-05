package vote

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
	"time"
)

func TestVoteEvent(t *testing.T) {
	resultVote := make(chan *MsgVoteResult, 1)
	voterList := []string{"voter1", "voter2", "voter3", "voter4", "voter5", "voter6", "voter7", "voter8", "voter9", "voter10"}

	nonce := int64(1)
	funcName := string("setCost")
	contracAddr := "0xf4B256427DEd7eaE62040b13684ba7487c0a1825"

	fmt.Printf("Create voting event\n")
	event := NewVoteEvent(
		"test_owner",
		"test_topic",
		&MsgVoteProposal{"contract",nil,&contracAddr,&funcName,&nonce,},
		10,
		33,
		voterList,
		resultVote)

	time.Sleep(time.Second * 1)
	fmt.Printf("Start test voting\n")
	event.Vote("voter1", 0)
	time.Sleep(time.Second * 1)
	event.Vote("voter2", 0)
	time.Sleep(time.Second * 1)
	event.Vote("voter3", 1)
	time.Sleep(time.Second * 1)
	event.Vote("voter4", 1)
	time.Sleep(time.Second * 1)
	event.Vote("voter5", 1)
	time.Sleep(time.Second * 1)
	event.Vote("voter6", 1)
	time.Sleep(time.Second * 1)
	event.Vote("voter7", 1)
	time.Sleep(time.Second * 1)
	event.Vote("voter8", 2)
	time.Sleep(time.Second * 1)

	param,_ := event.GetParam("voter1")

	if param.PassRate != 33 {
		t.Error("Passrate is not as expected")
	}

	if param.Duration != 10 {
		t.Error("Duration is not as expected")
	}

	status,_ := event.GetStatus("voter1")

	if len(status.ForList) != 5  ||
		len(status.AgainstList) != 2 ||
		len(status.AbstainedList) != 1 {
		t.Error("The result of voting is not correct")
	}

	t.Log("Abstained Name List")
	for _, item := range status.AbstainedList {
		t.Log(item)
	}

	t.Log("Against Name List")
	for _, item := range status.AgainstList {
		t.Log(item)
	}

	t.Log("For Name List")
	for _, item := range status.ForList {
		t.Log(item)
	}
	fmt.Printf("Start time in %s\n", status.Start)
	fmt.Printf("Expire time in %s\n", status.Expires)

	select {
	case msg := <-resultVote:
		fmt.Printf("Result for topic %s is %v\n", msg.Topic, msg.Value)
		if msg.Topic != "test_topic" ||
			msg.Value != true{
			t.Error("Vote Result not correct")
		}

		//uint256 nonce = 1;
		//string funcName = 'setCost';
		//address contractAddr = 0xf4B256427DEd7eaE62040b13684ba7487c0a1825;
		//hash = keccak256(abi.encodePacked(contractAddr, funcName, nonce));
		//hash = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", hash));
		hashStr := "1c04ad9fc1764607002e5bfe4b71538883f2fa7ad488d9d48d71ced019e6224a"

		hash,err:= hex.DecodeString(hashStr)
		if err != nil{
			t.Error ("Decode hash from string triggers error")
		}

		sig,err:= hex.DecodeString(*msg.Signature)
		if err != nil{
			t.Error ("Decode sig from string triggers error")
		}

		sigECDSA, err := crypto.SigToPub(hash, sig)

		pub := crypto.PubkeyToAddress(*sigECDSA)

		addrStr := pub.String()

		if addrStr != addr{
			t.Error ("Vote Signature not correct")
		}
		fmt.Printf("address signed is %s, as same as admin addr\n", addrStr)
	}
}
