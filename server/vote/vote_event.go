package vote

import (
	"time"
)

type VoteEvent struct {
	Owner string
	Topic string

	Proposal *MsgVoteProposal
	Param    *MsgVoteGetParam
	Status   *MsgVoteStatus

	ticker 		*time.Ticker
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
		Owner:    owner,
		Topic:    topic,
		Proposal: proposal,
		ticker:   time.NewTicker(durationTime),

		Param: &MsgVoteGetParam{
			Duration: duration,
			PassRate: passRate,

			VoterSize: uint(len(voterList)),
		},

		Status:&MsgVoteStatus{
			CurVoterList: voterList,
			Start:        startStr,
			Expires:      expiresStr,
		},
	}
	go e.exec()
	return e
}

func (e *VoteEvent) Vote (voter string, ballot uint) bool{

	  expires, _ := time.Parse(time.RFC850, e.Status.Expires)
	  now := time.Now().UTC()

	  if  now.After(expires) {
	  	return false
	  }

	for	i := len(e.Status.CurVoterList)-1; i>=0; i--{
		v:= e.Status.CurVoterList[i]
		if v == voter {
			switch ballot {
			case 0:
				e.Status.AgainstList = append(e.Status.AgainstList, voter)
			case 1:
				e.Status.ForList = append(e.Status.ForList, voter)
			default:
				e.Status.AbstainedList = append(e.Status.AbstainedList, voter)
			}
			e.Status.CurVoterList = append(e.Status.CurVoterList[:i], e.Status.CurVoterList[i+1:]...)
			return true
		}
	}
	return false
}

func (e *VoteEvent) GetStatus() *MsgVoteStatus {
	return e.Status
}

func (e *VoteEvent) GetParam() *MsgVoteGetParam {
	return  e.Param
}

func (e *VoteEvent) exec() {
		select {
		case <-e.ticker.C:
			forSize := len(e.Status.ForList)
			againstSize := len(e.Status.AgainstList)
			abstainedSize := len(e.Status.AbstainedList)

			currentRate := float64(forSize / (forSize + againstSize + abstainedSize)) * 100

			if currentRate - float64(e.Param.PassRate) > 0 {

				//TODO:
			}
			return
		}
}