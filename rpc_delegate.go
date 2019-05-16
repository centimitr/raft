package raft

// RPCDelegate delegates Raft objects for RPC service
type RPCDelegate struct {
	r *Raft
}

func NewRPCDelegate(r *Raft) *RPCDelegate {
	return &RPCDelegate{r}
}

func (d *RPCDelegate) AppendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) error {
	d.r.AppendEntries(&arg, reply)
	return nil
}

func (d *RPCDelegate) RequestVotes(arg RequestVoteArg, reply *RequestVoteReply) error {
	d.r.RequestVote(&arg, reply)
	return nil
}
