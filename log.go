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

type LogEntry struct {
	Index   LogEntryIndex
	Term    Term
	Command interface{}
}

func newLogEntry(l *Log, term Term, cmd interface{}) *LogEntry {
	return &LogEntry{
		Index:   LogEntryIndex(len(l.Entries) + 1),
		Term:    term,
		Command: cmd,
	}
}

type Log struct {
	Entries     []*LogEntry
	CommitIndex LogEntryIndex
	LastApplied LogEntryIndex

	mutex sync.RWMutex
	sm    StateMachine
}

func NewLog(stateMachine StateMachine) *Log {
	return &Log{
		sm: stateMachine,
	}
}

type logTx struct {
	Apply chan struct{}
	Done  chan struct{}
}

func newLogTx() *logTx {
	return &logTx{Apply: make(chan struct{}), Done: make(chan struct{})}
}

func (l *Log) append(term Term, cmd interface{}) (tx *logTx, err error) {
	tx = newLogTx()
	l.mutex.Lock()
	e := newLogEntry(l, term, cmd)
	// todo: save disk or make entries stable, committed
	l.Entries = append(l.Entries, e)
	// todo: stable storage error
	//if err != nil {
	//	return
	//}
	l.CommitIndex = e.Index
	// todo: eventually be applied to the state machine
	<-tx.Apply
	l.sm.Apply(e.Command)
	l.LastApplied = e.Index
	l.mutex.Unlock()
	close(tx.Done)
	return
}
