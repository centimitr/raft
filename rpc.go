package raft

type AppendEntriesArg struct {
	Term         Term
	LeaderId     NodeId
	PrevLogIndex LogEntryId
	PrevLogTerm  Term
	Entries      []LogEntry
	LeaderCommit LogEntryId
}

type AppendEntriesReply struct {
	Term    Term
	Success bool
}

func (s *State) AppendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) error {
	return nil
}

type RequestVotesArg struct {
	Term         Term
	CandidateId  NodeId
	LastLogIndex LogEntryId
	LastLogTerm  Term
}

type RequestVotesReply struct {
	Term        Term
	VoteGranted bool
}

func (s *State) RequestVotes(arg RequestVotesArg, reply *AppendEntriesReply) error {
	return nil
}
