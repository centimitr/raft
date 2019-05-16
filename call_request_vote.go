package raft

type RequestVoteArg struct {
	Term         Term
	CandidateID  NodeIndex
	LastLogIndex LogEntryIndex
	LastLogTerm  Term
}

type RequestVoteReply struct {
	CurrentTerm Term
	VoteGranted bool
}

func (r *Raft) RequestVote(arg *RequestVoteArg, reply *RequestVoteReply) {
	select {
	case <-r.shutdown:
		r.log("shutdown: RequestVote RPC")
		return
	default:
	}

	r.log("RequestVote: %d, term: %d", arg.CandidateID, arg.Term)
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.Store()

	// reject previous term request
	if arg.Term < r.CurrentTerm {
		reply.CurrentTerm = r.CurrentTerm
		return
	}

	// update term
	if r.CurrentTerm < arg.Term {
		r.CurrentTerm = arg.Term
		r.becomeFollower()
	}

	lastLogIndex := r.Log.lastIndex()
	lastLogTerm := r.Log.last().Term

	available := r.VotedFor == -1
	atLeastAsUpdate := (arg.LastLogTerm == lastLogTerm && arg.LastLogIndex >= lastLogIndex) || arg.LastLogTerm > lastLogTerm
	if available && atLeastAsUpdate {
		r.log("vote for: %d", arg.CandidateID)
		r.election.reset()
		r.Role = Follower
		r.VotedFor = arg.CandidateID
		reply.VoteGranted = true
	}
}
