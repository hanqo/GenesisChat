package vote

import (
	"log"
)

type VoteHandler struct {

   pendings 	map[string]	*VoteEvent

   ToVote     chan *MsgToVote
   FromVote   chan *MsgFromVote
   resultVote chan *MsgVoteResult

   RunDone   chan bool
   EpollDone chan bool
}

func NewVoteHandler() *VoteHandler{
	v := &VoteHandler{
		ToVote:     make(chan *MsgToVote,1),
		FromVote:   make(chan *MsgFromVote,1),
		resultVote: make(chan *MsgVoteResult,1000),

		RunDone:   make(chan bool, 1),
		EpollDone: make(chan bool, 1),

	}
	go v.run()
	go v.epoll()
	return v

}

func (v *VoteHandler) epoll(){

	for{
		select {
		case r :=<-v.resultVote:
			e, ok := v.pendings[r.Topic]

			if ok != true {
				log.Fatal("Epoll error, vote event not exist")
				return
			}

			v.FromVote <- &MsgFromVote{
				Owner:			e.Owner,
				Topic:			e.Topic,
				Typ:            "result",
				Result: 	r.Value,
			}

			delete(v.pendings, r.Topic)

		case <-v.EpollDone:
			log.Printf("Stop Vote handler polling")
			return
			}
		}
}

func (v *VoteHandler)run(){
	for{
		select {
		case msg := <- v.ToVote:
			switch msg.Typ {
			case "new_vote":
				go v.startNewEvent(msg)
			case "vote":
				go v.vote(msg)
			case "status":
				go v.getEventStatus(msg)
			default:
				log.Printf("Unrecognized message in voting run loop")
			}
		case <-v.RunDone:
			log.Printf("Stop Vote hanlder running")
			return
		}
	}
}

func (v *VoteHandler) startNewEvent(msg *MsgToVote) bool{

	if msg.NewVote == nil {
		log.Fatal("New vote info missing")
		return false
	}

	_, ok := v.pendings[msg.Topic]

	if ok == true {
		log.Fatal("Already exist a vote for this topic")
		return false
	}

	event:= NewVoteEvent(msg.Owner,
		msg.Topic,
		msg.NewVote.Proposal,
		msg.NewVote.Duration,
		msg.NewVote.PassRate,
		msg.NewVote.VoterList,
		v.resultVote)

	v.pendings[event.Topic] = event
	return true
}

func (v *VoteHandler) getEventStatus(msg *MsgToVote) bool{

	e, ok := v.pendings[msg.Topic]

	if ok != true {
		log.Fatal("Event not exists")
		return false
	}

	v.FromVote <- &MsgFromVote{
		Owner:		msg.Owner,
		Topic:		msg.Topic,
		Typ:        "status",
		Status:	e.GetStatus(),
	}
	return true
}

func (v *VoteHandler) getEventParam(msg *MsgToVote) bool{

	e, ok := v.pendings[msg.Topic]

	if ok != true {
		log.Fatal("Event not exists")
		return false
	}

	v.FromVote <- &MsgFromVote{
		Owner:		msg.Owner,
		Topic:		msg.Topic,
		Typ:        "param",
		Param:	e.GetParam(),
	}
	return true
}

func (v *VoteHandler) vote(msg *MsgToVote)bool{


	e, ok := v.pendings[msg.Topic]

	if ok != true {
		log.Fatal("Event not exists")
		return false
	}

	e.Vote(msg.Owner, msg.Ballot.Value)
	return true
}

