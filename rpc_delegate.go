package raft

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
