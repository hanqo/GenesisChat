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
	Param  *MsgVoteCurrentParam
	Result bool
}

type MsgNewVote struct {
	Proposal  *MsgVoteProposal
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

	Start   string
	Expires string
}

type MsgVoteCurrentParam struct {
	Duration  uint
	PassRate  uint // 0 <= PassRate <= 100
	VoterSize uint // number of all voters
}

type MsgVoteProposal struct {
	Typ  string
	Data interface{}
}

type MsgVoteResult struct {
	Topic string
	Value bool
}
