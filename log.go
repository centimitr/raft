package raft

import (
	"sync"
)

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

	mu sync.RWMutex
	sm StateMachine
}

func NewLog(stateMachine StateMachine) *Log {
	l := new(Log)
	// entries start from index 1
	l.Entries = []*LogEntry{new(LogEntry)}
	l.sm = stateMachine
	return l
}

type LogTx struct {
	Cancel chan error
	Apply  chan struct{}
	Done   chan struct{}
}

func newLogTx() *LogTx {
	return &LogTx{
		Cancel: make(chan error),
		Apply:  make(chan struct{}),
		Done:   make(chan struct{}),
	}
}

func retrieve(entries []*LogEntry, index LogEntryIndex) (entry *LogEntry) {
	if int(index) < 0 || int(index) >= len(entries) {
		return nil
	}
	entry = entries[index]
	return
}

// retrieve returns a LogEntry at given index.
func (l *Log) retrieve(index LogEntryIndex) (entry *LogEntry) {
	l.mu.RLock()
	entry = retrieve(l.Entries, index)
	l.mu.RUnlock()
	return
}

func (l *Log) last() (entry *LogEntry) {
	return l.retrieve(l.LastIndex)
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
func (l *Log) slice(startFrom LogEntryIndex) (entries []*LogEntry) {
	l.mu.RLock()
	entries = l.Entries[startFrom:]
	l.mu.RUnlock()
	return
}

// apply applies committed logs into the state machine.
func (l *Log) apply() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for l.CommitIndex > l.LastApplied {
		l.LastApplied++
		entry := retrieve(l.Entries, l.LastApplied)
		if entry != nil {
			l.sm.Apply(entry.Command)
		}
	}
}

// patch is used by followers to remove conflicts and append entries from the leader.
func (l *Log) patch(prevLogIndex LogEntryIndex, entries []*LogEntry) {
	l.mu.Lock()
	var pos int
	for i, entry := range l.Entries {
		if entry.Index == prevLogIndex {
			pos = i
			break
		}
	}
	l.Entries = append(l.Entries[:pos+1], entries...)
	l.mu.Unlock()
}

// append executes a log transaction to append a new log entry, and then apply it.
func (l *Log) append(term Term, cmd interface{}) (tx *LogTx, err error) {
	tx = newLogTx()
	l.mu.Lock()
	e := newLogEntry(l.LastIndex+1, term, cmd)
	// todo: save disk or make entries stable, committed
	l.Entries = append(l.Entries, e)
	l.LastIndex = e.Index
	l.CommitIndex = e.Index
	l.mu.Unlock()
	return
}

// NextIndex returns the upcoming log entry's index
func (l *Log) NextIndex() LogEntryIndex {
	return l.LastIndex + 1
}

func (l *Log) UpdateCommitIndex(currentTerm Term, peerCount int, ls *LeaderState) (ok bool) {
	l.mu.Lock()
	// when there is no peer, update commit index to latest directly
	// this is for development compatibility, not a use case
	if peerCount == 0 {
		l.CommitIndex = l.LastIndex
		ok = true
	}
	// from the log entry next to current commit index, find an index N that most followers can match
	for i := l.CommitIndex + 1; i <= l.LastIndex; i++ {
		entry := retrieve(l.Entries, i)
		if entry.Term != currentTerm {
			continue
		}
		cnt := 0
		ls.matchIndex.Range(func(key, value interface{}) bool {
			idx := value.(LogEntryIndex)
			if idx > i {
				cnt++
			}
			return true
		})
		if cnt*2 > peerCount {
			l.CommitIndex = i
			ok = true
		}
	}
	l.mu.Unlock()
	go l.apply()
	return
}

func (l *Log) UpdateCommitIndexFromLeader(leaderCommit LogEntryIndex) {
	if leaderCommit > l.CommitIndex {
		if l.LastIndex < leaderCommit {
			l.CommitIndex = l.LastIndex
		} else {
			l.CommitIndex = leaderCommit
		}
		go l.apply()
	}
}
