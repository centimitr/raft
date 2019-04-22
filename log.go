package raft

type LogEntryId int

type LogEntry struct {
	Id   LogEntryId
	Term Term
}

type Log struct {
	Entries []*LogEntry
}
