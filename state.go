package raft

type LeaderState struct {
	NextIndex  []LogEntryId
	MatchIndex []LogEntryId
}

type State struct {
	Role NodeRole

	CurrentTerm Term
	VotedFor    NodeId
	Log         Log

	CommitIndex LogEntryId
	LastApplied LogEntryId

	*LeaderState
}
