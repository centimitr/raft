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

func (r *Raft) AppendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) error {
	reply.Term = r.CurrentTerm
	if r.Election.Processing {
		if arg.Term.NotEarlierThan(r.CurrentTerm) {
			r.Election.Abandon()
		} else {
			reply.Success = false
		}
	}
	// validity
	r.Election.ResetTimer()
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

func (r *Raft) RequestVotes(arg RequestVotesArg, reply *AppendEntriesReply) error {
	// validity
	r.Election.ResetTimer()
	return nil
}
