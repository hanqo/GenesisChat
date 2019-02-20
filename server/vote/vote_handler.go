package vote

import (
	"log"
	"sync"
)

type VoteHandler struct {
	pendings map[string]*VoteEvent

	ToVote     chan *MsgToVote
	FromVote   chan *MsgFromVote
	resultVote chan *MsgVoteResult

	RunDone   chan bool
	EpollDone chan bool

	mutex sync.Mutex
}

func NewVoteHandler() *VoteHandler {
	v := &VoteHandler{
		ToVote:     make(chan *MsgToVote, 100),
		FromVote:   make(chan *MsgFromVote, 100),
		resultVote: make(chan *MsgVoteResult, 100),
		pendings:   make(map[string]*VoteEvent),

		RunDone:   make(chan bool, 1),
		EpollDone: make(chan bool, 1),
	}
	go v.run()
	go v.epoll()
	return v

}

func (v *VoteHandler) epoll() {

	for {
		select {
		case r := <-v.resultVote:
			v.mutex.Lock()
			e, ok := v.pendings[r.Topic]

			if ok != true {
				log.Fatal("Epoll error, vote event not exist")
				break
			}

			v.FromVote <- &MsgFromVote{
				Owner:  e.Owner,
				Topic:  e.Topic,
				Typ:    "result",
				Result: r.Value,
			}

			delete(v.pendings, r.Topic)
			v.mutex.Unlock()
		case <-v.EpollDone:
			log.Printf("Stop Vote handler polling")
			return
		}
	}
}

func (v *VoteHandler) run() {
	for {
		select {
		case msg := <-v.ToVote:
			switch msg.Typ {
			case "new_vote":
				go v.startNewEvent(msg)
			case "vote":
				go v.vote(msg)
			case "status":
				go v.getEventStatus(msg)
			case "parameter":
				go v.getEventParam(msg)
			default:
				log.Printf("Unrecognized message in voting run loop")
			}
		case <-v.RunDone:
			log.Printf("Stop Vote hanlder running")
			return
		}
	}
}

func (v *VoteHandler) startNewEvent(msg *MsgToVote) bool {

	if msg.NewVote == nil {
		log.Fatal("New vote info missing")
		return false
	}
	v.mutex.Lock()
	defer v.mutex.Unlock()
	_, ok := v.pendings[msg.Topic]

	if ok == true {
		log.Fatal("Already exist a vote for this topic")
		return false
	}

	event := NewVoteEvent(msg.Owner,
		msg.Topic,
		msg.NewVote.Proposal,
		msg.NewVote.Duration,
		msg.NewVote.PassRate,
		msg.NewVote.VoterList,
		v.resultVote)

	v.pendings[event.Topic] = event
	return true
}

func (v *VoteHandler) getEventStatus(msg *MsgToVote) bool {
	v.mutex.Lock()
	e, ok := v.pendings[msg.Topic]
	v.mutex.Unlock()

	if ok != true {
		log.Fatal("Event not exists")
		return false
	}
	status, _ := e.GetStatus(msg.Owner)

	v.FromVote <- &MsgFromVote{
		Owner:  msg.Owner,
		Topic:  msg.Topic,
		Typ:    "status",
		Status: status,
	}
	return true
}

func (v *VoteHandler) getEventParam(msg *MsgToVote) bool {

	v.mutex.Lock()
	e, ok := v.pendings[msg.Topic]
	v.mutex.Unlock()

	if ok != true {
		log.Fatal("Event not exists")
		return false
	}

	param, _ := e.GetParam(msg.Owner)

	v.FromVote <- &MsgFromVote{
		Owner: msg.Owner,
		Topic: msg.Topic,
		Typ:   "parameter",
		Param: param,
	}
	return true
}

func (v *VoteHandler) vote(msg *MsgToVote) bool {


	v.mutex.Lock()
	e, ok := v.pendings[msg.Topic]
	v.mutex.Unlock()

	if ok != true {
		log.Fatal("Event not exists")
		return false
	}

	e.Vote(msg.Owner, msg.Ballot.Value)
	return true
}
