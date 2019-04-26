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

//func (l *Log) Run() () {
//	go func() {
//		for entry := range l.apply {
//			l.sm.Apply(entry.Command)
//			l.LastApplied = entry.Index
//		}
//	}()
//}

func (l *Log) Append(term Term, cmd interface{}) (err error) {
	l.mutex.Lock()
	e := newLogEntry(l, term, cmd)
	// todo: save disk or make entries stable, committed
	l.Entries = append(l.Entries, e)
	// todo: stable storage error
	//if err != nil {
	//	return
	//}
	l.CommitIndex = e.Index
	l.sm.Apply(e.Command)
	l.LastApplied = e.Index
	l.mutex.Unlock()
	return
}
