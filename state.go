package raft

import "sync"

type LeaderState struct {
	NextIndex  []LogEntryIndex
	MatchIndex []LogEntryIndex
}

type State struct {
	Role Role
	Id   NodeId

	CurrentTerm Term
	VotedFor    NodeId

	*Log
	*LeaderState

	mutex sync.RWMutex
}

func NewState() *State {
	s := new(State)
	s.Id = NewNodeId()
	return s
}
