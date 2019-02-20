package vote

import (
	"fmt"
	"testing"
	"time"
)

/*func TestVoteHandler1(t *testing.T){
	v:= NewVoteHandler()
	v.ToVote <- &MsgToVote{
		Owner:"test_owner1",
		Topic:"test_topic1",
		Typ:"new_vote",
		NewVote:&MsgNewVote{
			Proposal: &MsgVoteProposal{"test", nil},
			Duration:10,
			PassRate:33,
			VoterList:[]string{"voter1","voter2","voter3","voter4","voter5"},},}

	time.Sleep(time.Second * 1)

	//Vote starts
	v.ToVote <- &MsgToVote{
		Owner:"voter1",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{0},}
	v.ToVote <- &MsgToVote{
		Owner:"voter2",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{2},}
	v.ToVote <- &MsgToVote{
		Owner:"voter3",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{1},}
	v.ToVote <- &MsgToVote{
		Owner:"voter4",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{1},}
	time.Sleep(time.Second * 1)

	//Request status of vote
	v.ToVote <- &MsgToVote{
		Owner:"voter4",
		Topic:"test_topic1",
		Typ:"status",}

	//Request paramter of vote
	v.ToVote <- &MsgToVote{
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
				fmt.Printf("Result for topic %s is %v\n", msg.Topic, msg.Result)
				if msg.Topic != "test_topic1" ||
					msg.Result != true {
					t.Error("Vote Result not correct")
				}
				return

			}else if msg.Typ== "param"{
				param:= msg.Param

				if param.PassRate!=33||param.Duration!= 10||param.VoterSize!= 5{
					t.Error("Vote Paramter not correct")
				}
			}
		}
	}

}*/

func TestVoteHandler2(t *testing.T){
	v:= NewVoteHandler()
	resultCnt := 0
	v.ToVote <- &MsgToVote{
		Owner:"test_owner1",
		Topic:"test_topic1",
		Typ:"new_vote",
		NewVote:&MsgNewVote{
			Proposal: &MsgVoteProposal{"test", nil},
			Duration:10,
			PassRate:33,
			VoterList:[]string{"voter1","voter2","voter3","voter4","voter5"},},}

	v.ToVote <- &MsgToVote{
		Owner:"test_owner2",
		Topic:"test_topic2",
		Typ:"new_vote",
		NewVote:&MsgNewVote{
			Proposal: &MsgVoteProposal{"test", nil},
			Duration:10,
			PassRate:33,
			VoterList:[]string{"voter5","voter6","voter7","voter8","voter9"},},}

	time.Sleep(time.Second * 1)

	//Vote starts
	v.ToVote <- &MsgToVote{
		Owner:"voter1",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{0},}
	v.ToVote <- &MsgToVote{
		Owner:"voter2",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{2},}
	v.ToVote <- &MsgToVote{
		Owner:"voter3",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{1},}
	v.ToVote <- &MsgToVote{
		Owner:"voter4",
		Topic:"test_topic1",
		Typ:"vote",
		Ballot: &MsgBallot{1},}

	//Vote starts
	v.ToVote <- &MsgToVote{
		Owner:"voter5",
		Topic:"test_topic2",
		Typ:"vote",
		Ballot: &MsgBallot{0},}
	v.ToVote <- &MsgToVote{
		Owner:"voter6",
		Topic:"test_topic2",
		Typ:"vote",
		Ballot: &MsgBallot{2},}
	v.ToVote <- &MsgToVote{
		Owner:"voter7",
		Topic:"test_topic2",
		Typ:"vote",
		Ballot: &MsgBallot{2},}
	v.ToVote <- &MsgToVote{
		Owner:"voter8",
		Topic:"test_topic2",
		Typ:"vote",
		Ballot: &MsgBallot{1},}
	time.Sleep(time.Second * 1)

	for{
		select {
		case msg := <-v.FromVote:
			 if msg.Typ== "result"{
			 	resultCnt ++
				fmt.Printf("Result for topic %s is %v\n", msg.Topic, msg.Result)

				if msg.Topic == "test_topic1" &&
					msg.Result != true {
					t.Error("Vote Result not correct")
				}

				 if msg.Topic == "test_topic2" &&
					 msg.Result != false {
					 t.Error("Vote Result not correct")
				 }

				if resultCnt >=2{
					return
				}
			}
		}
	}

}