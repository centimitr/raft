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

func checkReqTerm(r *Raft, term Term) bool {
	if term.EarlierThan(r.CurrentTerm) {
		return false
	}
	if term.LaterThan(r.CurrentTerm) {
		if r.Role.Is(Candidate) {
			if r.Election.Processing {
				r.Election.Abandon()
			}
		}
		r.Role.set(Follower)
		r.CurrentTerm = term
	}
	r.Election.ResetTimer()
	return true
}

func (r *Raft) appendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) (err error) {
	reply.Term = r.CurrentTerm
	if !checkReqTerm(r, arg.Term) {
		return
	}
	if !r.Log.match(arg.PrevLogIndex, arg.PrevLogTerm) {
		return
	}
	// todo: two leader
	if !r.Role.Is(Follower) {
		panic("debug: heartbeats should keep node followers")
		return
	}
	// todo: conflicts
	// todo: append new entries
	if len(arg.Entries) == 0 {
		log("heartbeats")
	}
	if arg.LeaderCommit > r.CommitIndex {
		if r.Log.LastIndex < arg.LeaderCommit {
			r.CommitIndex = r.Log.LastIndex
		} else {
			r.CommitIndex = arg.LeaderCommit
		}
		go r.Log.apply()
	}
	//
	reply.Success = true
	return
}

func (r *Raft) requestVotes(arg RequestVotesArg, reply *AppendEntriesReply) (err error) {
	reply.Term = r.CurrentTerm
	if !checkReqTerm(r, arg.Term) {
		return
	}
	// todo: log validity
	//var logValid bool
	//if r.VotedFor.IsEmptyOrEqualTo(arg.CandidateId) {
	r.VotedFor = arg.CandidateId
	reply.Success = true
	//}
	return
}
