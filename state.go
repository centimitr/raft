package raft

type LeaderState struct {
	NextIndex  []LogEntryId
	MatchIndex []LogEntryId
}

type State struct {
	Role Role

	CurrentTerm Term
	VotedFor    NodeId
	Log         Log

	CommitIndex LogEntryId
	LastApplied LogEntryId

	*LeaderState
}

func NewState() *State {
	return new(State)
}
