package raft

import "fmt"

type AppendEntriesArg struct {
	Term         Term
	LeaderId     NodeId
	PrevLogIndex LogEntryIndex
	PrevLogTerm  Term
	Entries      []*LogEntry
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
	r.mu.Lock()
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
	r.mu.Unlock()
	r.Election.ResetTimer()
	return true
}

func (r *Raft) appendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) (err error) {
	fmt.Println("CALL APPEND", len(arg.Entries))
	fmt.Printf("%+v\n", arg)

	r.mu.RLock()
	reply.Term = r.CurrentTerm
	r.mu.RUnlock()
	if !checkReqTerm(r, arg.Term) {
		return
	}
	if !r.Log.match(arg.PrevLogIndex, arg.PrevLogTerm) {
		return
	}
	// todo: check two leader
	r.mu.RLock()
	role := r.Role
	r.mu.RUnlock()
	if !role.Is(Follower) {
		fmt.Println("ROLE:", role.String())
		panic("debug: heartbeats should keep node followers")
		return
	}
	if len(arg.Entries) != 0 {
		fmt.Println("RECV APPEND")
		fmt.Printf("%+v\n", arg)
	}
	// todo: check conflicts, append new entries
	if len(arg.Entries) > 0 {
		r.Log.patch(arg.PrevLogIndex, arg.Entries)
		// log heartbeats
		//if len(arg.Entries) == 0 {
		//log("recv: heartbeats")
		//}
		r.Log.UpdateCommitIndexFromLeader(arg.LeaderCommit)
	}
	reply.Success = true
	return
}

func (r *Raft) requestVotes(arg RequestVotesArg, reply *RequestVotesReply) (err error) {
	r.mu.RLock()
	reply.Term = r.CurrentTerm
	r.mu.RUnlock()
	if !checkReqTerm(r, arg.Term) {
		return
	}
	lastLogEntry := r.Log.last()
	hasMoreUpToDateLog := arg.LastLogTerm > lastLogEntry.Term ||
		(arg.LastLogTerm == lastLogEntry.Term && arg.LastLogIndex >= lastLogEntry.Index)
	r.mu.Lock()
	if r.VotedFor.IsEmptyOrEqualTo(arg.CandidateId) && hasMoreUpToDateLog {
		r.VotedFor = arg.CandidateId
		reply.VoteGranted = true
	}
	r.mu.Unlock()
	return
}
