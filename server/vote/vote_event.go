package vote

import (
	"log"
	"sync"
	"time"
)

type VoteEvent struct {
	Owner string
	Topic string

	Proposal *MsgVoteProposal
	Param    *MsgVoteCurrentParam
	Status   *MsgVoteStatus

	ticker   *time.Ticker
	votedMap map[string]bool

	chanResult chan *MsgVoteResult
	mutex    sync.Mutex
}

func NewVoteEvent(
	owner string,
	topic string,
	proposal *MsgVoteProposal,
	duration uint,
	passRate uint,
	voterList []string,
	chanResult chan *MsgVoteResult) *VoteEvent {

	nowTime := time.Now().UTC()
	startStr := nowTime.Format(time.RFC850)

	durationTime := time.Duration(duration) * time.Second
	expiresTime := nowTime.Add(durationTime)
	expiresStr := expiresTime.Format(time.RFC850)


	votedMap := make(map[string]bool)

	for _, v := range voterList {

		votedMap[v] = false
	}

	e := &VoteEvent{
		Owner:      owner,
		Topic:      topic,
		ticker:     time.NewTicker(durationTime),
		votedMap:	votedMap,
		chanResult: chanResult,

		Proposal: proposal,
		Param: &MsgVoteCurrentParam{
			Duration: duration,
			PassRate: passRate,

			VoterSize: uint(len(voterList)),
		},

		Status: &MsgVoteStatus{
			Start:   startStr,
			Expires: expiresStr,
		},
	}
	go e.timeOut()
	return e
}

func (e *VoteEvent) Vote(voter string, value uint) bool {

	expires, _ := time.Parse(time.RFC850, e.Status.Expires)
	now := time.Now().UTC()

	if now.After(expires) {
		return false
	}

	v, ok := e.votedMap[voter]

	if ok != true || v != false {
		log.Fatal("Voter doesn't exist or voted")
		return false
	}

	e.mutex.Lock()
	e.votedMap[voter] = false
	switch value {
	case 0:
		e.Status.AgainstList = append(e.Status.AgainstList, voter)
	case 1:
		e.Status.ForList = append(e.Status.ForList, voter)
	default:
		e.Status.AbstainedList = append(e.Status.AbstainedList, voter)
	}
	e.mutex.Unlock()
	return true
}

func (e *VoteEvent) GetStatus(voter string) (*MsgVoteStatus, bool) {
	_, ok := e.votedMap[voter]

	if ok != true  {
		log.Fatal("Only candidates can access status")
		return nil, false
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.Status,true
}

func (e *VoteEvent) GetParam(voter string) (*MsgVoteCurrentParam,bool) {
	_, ok := e.votedMap[voter]

	if ok != true  {
		log.Fatal("Only candidates can access parameter")
		return nil, false
	}
	return e.Param, true
}

func (e *VoteEvent) timeOut() {
	select {
	case <-e.ticker.C:
		e.mutex.Lock()
		forSize := len(e.Status.ForList)
		e.mutex.Unlock()

		currentRate := float64(forSize)/float64(len(e.votedMap)) * 100

		if currentRate-float64(e.Param.PassRate) > 0 {
			e.chanResult <- &MsgVoteResult{
				Topic: e.Topic,
				Value: true,
			}

		} else {
			e.chanResult <- &MsgVoteResult{
				Topic: e.Topic,
				Value: false,
			}
		}

		return
	}
}
