package raft

import "sync"

type StateMachineDelegate interface {
	Apply(cmd interface{}) error
}

type StateMachine interface {
	Apply(cmd interface{})
	Delegate(delegate StateMachineDelegate)
}

type LogEntryIndex = int

//func (i LogEntryIndex) succ() LogEntryIndex {
//	return i + 1
//}

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

// retrieve returns a LogEntry at given index.
func (l *Log) retrieve(index LogEntryIndex) *LogEntry {
	if int(index) < 0 || int(index) >= len(l.Entries) {
		return nil
	}
	return l.Entries[index]
}

// match checks if the entry at given index has the specified term value.
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

// slice returns log entries after given index.
func (l *Log) slice(startFrom LogEntryIndex) []*LogEntry {
	return l.Entries[startFrom:]
}

// apply applies committed logs into the state machine.
func (l *Log) apply() {
	for l.CommitIndex > l.LastApplied {
		l.LastApplied++
		entry := l.retrieve(l.LastApplied)
		if entry != nil {
			l.sm.Apply(entry.Command)
		}
	}
}

// patch is used by followers to remove conflicts and append entries from the leader.
func (l *Log) patch([]LogEntry) {

}

// append executes a log transaction to append a new log entry, and then apply it.
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

// NextIndex returns the upcoming log entry's index
func (l *Log) NextIndex() LogEntryIndex {
	return l.LastIndex + 1
}
