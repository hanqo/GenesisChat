package vote

type MsgToVote struct {
	owner 		string
	topic 		string
	typ 		string

	newVote 	*MsgNewVote
	ballot  	*MsgBallot
}

type MsgFromVote struct {
	owner 		string
	topic 		string
	typ 		string

	status 		*MsgVoteStatus
	param  		*MsgVoteGetParam
	set     	*MsgVoteSetParam
}

type MsgNewVote struct {
	proposal 	MsgVoteProposal
	duration    uint
	passRate    uint
	voterList   []string
}

type MsgBallot struct {
	owner 		string
	topic 		string
	value       uint
}

type MsgVoteStatus struct {

	forList				[]string
	againstList			[]string
	abstainedList       []string
	curVoterList		[]string //current voter list, remove voter after his vote.

	start 				string
	expires      		string
}

type MsgVoteGetParam struct {

	duration 			uint
	passRate 			uint // 0 <= passRate <= 100
	voterSize        	uint // number of all voters
}

type MsgVoteProposal struct {
	typ 				string
	setValues 			*[]string
	kickoutCitizens 	*[]string
}

type MsgVoteSetParam struct {
	duration     uint
	passRate     uint
}