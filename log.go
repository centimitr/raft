package raft

import "sync"

type StateMachineDelegate interface {
	Apply(cmd interface{}) error
}

type StateMachine interface {
	Apply(cmd interface{})
	Delegate(delegate StateMachineDelegate)
}

type LogEntryIndex int

func (i LogEntryIndex) succ() LogEntryIndex {
	return i + 1
}

type LogEntry struct {
	Index   LogEntryIndex
	Term    Term
	Command interface{}
}

func newLogEntry(index LogEntryIndex, term Term, cmd interface{}) *LogEntry {
	return &LogEntry{
		Index:   index,
		Term:    term,
		Command: cmd,
	}
}

type Log struct {
	Entries     []*LogEntry
	CommitIndex LogEntryIndex
	LastApplied LogEntryIndex
	LastIndex   LogEntryIndex

	mutex sync.RWMutex
	sm    StateMachine
}

func NewLog(stateMachine StateMachine) *Log {
	l := new(Log)
	// entries start from index 1
	l.Entries = []*LogEntry{nil}
	l.sm = stateMachine
	return l
}

type logTx struct {
	Cancel chan error
	Apply  chan struct{}
	Done   chan struct{}
}

func newLogTx() *logTx {
	return &logTx{
		Cancel: make(chan error),
		Apply:  make(chan struct{}),
		Done:   make(chan struct{}),
	}
}

func (l *Log) retrieve(index LogEntryIndex) *LogEntry {
	if int(index) < 0 || int(index) >= len(l.Entries) {
		return nil
	}
	return l.Entries[index]
}

func (l *Log) match(index LogEntryIndex, wantedTerm Term) bool {
	e := l.retrieve(index)
	if e == nil {
		return false
	}
	if e.Term == wantedTerm {
		return true
	}
	return false
}

func (l *Log) apply() {
	for l.CommitIndex > l.LastApplied {
		l.LastApplied++
		entry := l.retrieve(l.LastApplied)
		if entry != nil {
			l.sm.Apply(entry.Command)
		}
	}
}

func (l *Log) append(term Term, cmd interface{}) (tx *logTx, err error) {
	tx = newLogTx()
	l.mutex.Lock()
	e := newLogEntry(l.LastIndex+1, term, cmd)
	// todo: save disk or make entries stable, committed
	l.Entries = append(l.Entries, e)
	l.LastIndex = e.Index
	// todo: stable storage error
	//if err != nil {
	//	return
	//}
	l.CommitIndex = e.Index
	// todo: eventually be applied to the state machine
	<-tx.Apply
	l.apply()
	l.mutex.Unlock()
	close(tx.Done)
	return
}
