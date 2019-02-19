package vote

type MsgToVote struct {
	Owner string
	Topic string
	Typ   string

	NewVote *MsgNewVote
	Ballot  *MsgBallot
}

type MsgFromVote struct {
	Owner string
	Topic string
	Typ   string

	Status *MsgVoteStatus
	Param  *MsgVoteGetParam
	Set    *MsgVoteSetParam
}

type MsgNewVote struct {
	Proposal  MsgVoteProposal
	Duration  uint
	PassRate  uint
	VoterList []string
}

type MsgBallot struct {
	Owner string
	Topic string
	Value uint
}

type MsgVoteStatus struct {

	ForList       []string
	AgainstList   []string
	AbstainedList []string
	CurVoterList  []string //current voter list, remove voter after his vote.

	Start   string
	Expires string
}

type MsgVoteGetParam struct {

	Duration  uint
	PassRate  uint // 0 <= PassRate <= 100
	VoterSize uint // number of all voters
}

type MsgVoteProposal struct {
	Typ  string
	Data interface{}
}

type MsgVoteSetParam struct {
	Duration uint
	PassRate uint
}