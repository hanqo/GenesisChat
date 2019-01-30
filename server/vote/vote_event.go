package vote

import (
	"sync"
	"time"
)

type VoteEvent struct {
	owner		string
	topic       string

	proposal MsgVoteProposal
	param    MsgVoteGetParam
	status   MsgVoteStatus

	ticker 		*time.Ticker
	mutex 		sync.Mutex

}

func NewVoteEvent(owner string,
	              topic string,
	              proposal *MsgVoteProposal,
	              duration uint,
	              passRate uint,
	              voterList []string) *VoteEvent{

	nowTime:= time.Now().UTC()
	startStr := nowTime.Format(time.RFC850)

	durationTime := time.Duration(duration) * time.Second
	expiresTime := nowTime.Add(durationTime)
	expiresStr := expiresTime.Format(time.RFC850)

	e := &VoteEvent{
		owner:		owner,
		topic:		topic,
		proposal: 	*proposal,
		ticker: 	time.NewTicker(durationTime),

		param: MsgVoteGetParam{
			duration:duration,
			passRate:passRate,

			voterSize:uint(len(voterList)),
		},

		status:MsgVoteStatus{
			curVoterList: voterList,
			start:startStr,
			expires:expiresStr,
		},
	}
	go e.exec()
	return e
}

func (e *VoteEvent) Vote (voter string, ballot uint) bool{

	  expires, _ := time.Parse(time.RFC850, e.status.expires)
	  now := time.Now().UTC()

	  if  now.After(expires) {
	  	return false
	  }

	for	i := len(e.status.curVoterList)-1; i>=0; i--{
		v:= e.status.curVoterList[i]
		if v == voter {
			e.mutex.Lock()
			switch ballot {
			case 0:
				e.status.againstList = append(e.status.againstList, voter)
			case 1:
				e.status.forList = append(e.status.forList, voter)
			default:
				e.status.abstainedList = append(e.status.abstainedList, voter)
			}
			e.mutex.Unlock()
			e.status.curVoterList = append(e.status.curVoterList[:i], e.status.curVoterList[i+1:]...)
			return true
		}
	}
	return false
}

func (e *VoteEvent) GetStatus() *MsgVoteStatus {
	return &e.status
}

func (e *VoteEvent) GetParam() *MsgVoteGetParam {
	return  &e.param
}

func (e *VoteEvent) exec() {
	for{
		select {
		case <-e.ticker.C:
			forSize := len(e.status.forList)
			againstSize := len(e.status.againstList)
			abstainedSize := len(e.status.abstainedList)

			currentRate := float64(forSize / (forSize + againstSize + abstainedSize)) * 100

			if currentRate - float64(e.param.passRate) > 0 {

				//TODO:
			}
			return
		}
	}
}