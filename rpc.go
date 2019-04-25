package raft

// AppendEntries RPC

type AppendEntriesArg struct {
	Term         Term
	LeaderId     NodeId
	PrevLogIndex LogEntryIndex
	PrevLogTerm  Term
	Entries      []LogEntry
	LeaderCommit LogEntryIndex
}

func NewAppendEntriesArg(state *State) *AppendEntriesArg {
	return &AppendEntriesArg{
		Term: state.CurrentTerm,
	}
}

type AppendEntriesReply struct {
	Term    Term
	Success bool
}

func (r *Raft) appendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) (err error) {
	reply.Term = r.CurrentTerm
	if r.Election.Processing {
		if arg.Term.NotEarlierThan(r.CurrentTerm) {
			r.Election.Abandon()
			return
		}
	}
	// validity
	r.Election.ResetTimer()
	return nil
}

// RequestVotes RPC

type RequestVotesArg struct {
	Term         Term
	CandidateId  NodeId
	LastLogIndex LogEntryIndex
	LastLogTerm  Term
}

func NewRequestVotesArg(state *State) *RequestVotesArg {
	return &RequestVotesArg{
		Term:        state.CurrentTerm,
		CandidateId: state.Id,
		// todo: log index and term
		LastLogIndex: 0,
		LastLogTerm:  0,
	}
}

type RequestVotesReply struct {
	Term        Term
	VoteGranted bool
}

func (r *Raft) requestVotes(arg RequestVotesArg, reply *AppendEntriesReply) (err error) {
	reply.Term = r.CurrentTerm
	if arg.Term.NotEarlierThan(r.CurrentTerm) {
		return
	}
	// todo: log validity
	var logValid bool
	if r.VotedFor.IsEmptyOrEqualTo(arg.CandidateId) && logValid {
		r.VotedFor = arg.CandidateId
		reply.Success = true
	}
	// todo: check if need reset election
	//r.Election.ResetTimer()
	return
}

// RPCDelegate delegates Raft objects for RPC service
type RPCDelegate struct {
	r *Raft
}

func NewRPCDelegate(r *Raft) *RPCDelegate {
	return &RPCDelegate{r}
}

func (d *RPCDelegate) AppendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) error {
	return d.r.appendEntries(arg, reply)
}

func (d *RPCDelegate) RequestVotes(arg RequestVotesArg, reply *AppendEntriesReply) error {
	return d.r.requestVotes(arg, reply)
}
