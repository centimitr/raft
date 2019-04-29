package raft

type AppendEntriesArg struct {
	Term         Term
	LeaderId     NodeId
	PrevLogIndex LogEntryIndex
	PrevLogTerm  Term
	Entries      []LogEntry
	LeaderCommit LogEntryIndex
}

type AppendEntriesReply struct {
	Term    Term
	Success bool
}

type RequestVotesArg struct {
	Term         Term
	CandidateId  NodeId
	LastLogIndex LogEntryIndex
	LastLogTerm  Term
}

type RequestVotesReply struct {
	Term        Term
	VoteGranted bool
}

func (r *Raft) appendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) (err error) {
	reply.Term = r.CurrentTerm
	if arg.Term.LaterThan(r.CurrentTerm) {
		if r.Role.Is(Candidate) {
			if r.Election.Processing {
				r.Election.Abandon()
			}
		}
		r.Role.set(Follower)
		r.CurrentTerm = arg.Term
	}

	switch r.Role.typ {
	case Follower:
		r.Election.ResetTimer()
		// todo: appendEntries
		reply.Success = true
	case Leader, Candidate:
	}
	return nil
}

func (r *Raft) requestVotes(arg RequestVotesArg, reply *AppendEntriesReply) (err error) {
	reply.Term = r.CurrentTerm
	if arg.Term.LaterThan(r.CurrentTerm) {
		if r.Role.Is(Candidate) {
			if r.Election.Processing {
				r.Election.Abandon()
			}
		}
		r.Role.set(Follower)
		r.CurrentTerm = arg.Term
	}

	switch r.Role.typ {
	case Follower:
		r.Election.ResetTimer()
		// todo: log validity
		var logValid bool
		if r.VotedFor.IsEmptyOrEqualTo(arg.CandidateId) && logValid {
			r.VotedFor = arg.CandidateId
			reply.Success = true
		}
	case Leader, Candidate:

	}
	return
}
