package raft

import (
	"sync"
)

type LogEntry struct {
	Term    int
	Command interface{}
}

type Log struct {
	Logs        []LogEntry
	commitCond  *sync.Cond
	CommitIndex LogEntryIndex
	LastApplied LogEntryIndex
	Snapshot    LogSnapshot
}

type LogSnapshot struct {
	LastIncludedIndex LogEntryIndex
	LastIncludedTerm  Term
}

func (l *Log) init(locker sync.Locker) {
	l.Logs = []LogEntry{{}}
	l.commitCond = sync.NewCond(locker)
}

func (l *Log) retrieve(index LogEntryIndex) LogEntry {
	return l.Logs[l.realIndex(index)]
}

func (l *Log) last() LogEntry {
	return l.Logs[l.realIndex(l.lastIndex())]
}

func (l *Log) lastIndex() LogEntryIndex {
	return l.len() - 1
}

func (l *Log) firstIndex() int {
	return l.Snapshot.LastIncludedIndex + 1
}

func (l *Log) append(term Term, command interface{}) {
	l.Logs = append(l.Logs, LogEntry{term, command})
}

func (l *Log) len() int {
	return len(l.Logs) + l.Snapshot.LastIncludedIndex
}

func (l *Log) realIndex(i LogEntryIndex) LogEntryIndex {
	return i - l.Snapshot.LastIncludedIndex
}
