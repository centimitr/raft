package raft

import "sync"

type State struct {
	Role Role
	Id   NodeId

	CurrentTerm Term
	VotedFor    NodeId

	*Log
	Leader *LeaderState

	mutex sync.RWMutex
}

func NewState() *State {
	s := new(State)
	s.Id = NewNodeId()
	return s
}

type LeaderState struct {
	log        *Log
	nextIndex  sync.Map
	matchIndex sync.Map
}

func newLeaderState(l *Log) *LeaderState {
	return &LeaderState{log: l}
}

func (ls *LeaderState) Reset() {
	ls.nextIndex = sync.Map{}
	ls.matchIndex = sync.Map{}
}

func (ls *LeaderState) NextIndex(id PeerId) LogEntryIndex {
	v, _ := ls.nextIndex.LoadOrStore(id, ls.log.NextIndex())
	return v.(LogEntryIndex)
}

func (ls *LeaderState) MatchIndex(id PeerId) LogEntryIndex {
	v, _ := ls.matchIndex.LoadOrStore(id, LogEntryIndex(0))
	return v.(LogEntryIndex)
}
