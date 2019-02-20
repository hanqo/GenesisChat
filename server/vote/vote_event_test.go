package vote

import (
	"fmt"
	"testing"
	"time"
)

func TestVoteEvent(t *testing.T) {
	resultVote := make(chan *MsgVoteResult, 1)
	voterList := []string{"voter1", "voter2", "voter3", "voter4", "voter5", "voter6", "voter7", "voter8", "voter9", "voter10"}

	fmt.Printf("Create voting event\n")
	event := NewVoteEvent(
		"test_owner",
		"test_topic",
		nil,
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

	if len(status.ForList) != 5 ||
		len(status.AgainstList) != 2 ||
		len(status.AbstainedList) != 1 {
		t.Error("The result of voting is not correct")
	}

	fmt.Printf("Abstained Name List\n")
	for _, item := range status.AbstainedList {
		fmt.Printf("%s\t", item)
	}
	fmt.Printf("\n")

	fmt.Printf("Against Name List\n")
	for _, item := range status.AgainstList {
		fmt.Printf("%s\t", item)
	}
	fmt.Printf("\n")

	fmt.Printf("For Name List\n")
	for _, item := range status.ForList {
		fmt.Printf("%s\t", item)
	}
	fmt.Printf("\n")
	fmt.Printf("Start time in %s\n", status.Start)
	fmt.Printf("Expire time in %s\n", status.Expires)

	select {
	case msg := <-resultVote:
		fmt.Printf("Result for topic %s is %v", msg.Topic, msg.Value)
		if msg.Topic != "test_topic" ||
			msg.Value != true {
			t.Error("Vote Result not correct")
		}
	}
}
