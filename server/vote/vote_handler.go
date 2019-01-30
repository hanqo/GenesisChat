package vote

import (
	"log"
	"time"
)

type VoteHandler struct {

   pendings 	[]*VoteEvent

   toVote   	chan *MsgToVote
   fromVote 	chan *MsgFromVote

   runDone 		chan bool
   pollDone 	chan bool

   ticker       *time.Ticker
   add		    chan *VoteEvent

}

func NewVoteHandler() *VoteHandler{
	v := &VoteHandler{
		toVote:			make(chan *MsgToVote,1),
		fromVote:		make(chan *MsgFromVote,1),
		runDone:		make(chan bool, 1),
		pollDone:		make(chan bool, 1),

		ticker:			time.NewTicker(time.Second * 1),
		add:    		make(chan *VoteEvent,1),
	}
	go v.run()
	go v.poll()
	return v

}

func (v *VoteHandler) poll(){

	for{
		select {
		case <-v.ticker.C:
			for i:= len(v.pendings)-1; i>=0; i--{
				now := time.Now()
				event:= v.pendings[i]
				expires, _ := time.Parse(time.RFC850, event.status.expires)
				if now.After(expires){
					v.pendings = append(v.pendings[:i], v.pendings[i+1:]...)
				}
			}
		case s:= <-v.add:
			v.pendings = append(v.pendings,s)
		case <-v.pollDone:
			log.Printf("Stop Vote handler polling")
			return

			}
		}

}

func (v *VoteHandler)run(){
	for{
		select {
		case msg := <- v.toVote:
			switch msg.typ {
			case "new_vote":
				go v.startNewEvent(msg)
			case "vote":
				go v.vote(msg)
			case "status":
				go v.getEventStatus(msg)
			default:
				log.Printf("Unrecognized message in voting run loop")
			}
		case <-v.runDone:
			log.Printf("Stop Vote hanlder running")
			return
		}
	}

}

func (v *VoteHandler) startNewEvent(msg *MsgToVote){

	if msg.newVote == nil {
		log.Printf("New vote info missing")
	}


	//TODO if possible to use map instead array
	for _,e:= range v.pendings{
		if e.topic == msg.topic{
		log.Printf("Already exist a vote for this topic")
			return
		}
	}

	event:= NewVoteEvent(msg.owner,
		msg.topic,
		&msg.newVote.proposal,
		msg.newVote.duration,
		msg.newVote.passRate,
		msg.newVote.voterList)

	v.add<-event
}

func (v *VoteHandler) getEventStatus(msg *MsgToVote){

	//TODO: very ugly search
	for _,e:= range v.pendings{

		if e.topic == msg.topic{

			v.fromVote <- &MsgFromVote{
					owner:msg.owner,
					topic:msg.topic,
					status:e.GetStatus(),
					}
			return
		}
	}

	log.Fatal("Event not exist")
}

func (v *VoteHandler) vote(msg *MsgToVote){

	//TODO: very ugly search
	for _,e:= range v.pendings{

		if e.topic == msg.topic{
			e.Vote(msg.owner, msg.ballot.value)
			return
		}
	}

	log.Fatal("Event not exist")
}

