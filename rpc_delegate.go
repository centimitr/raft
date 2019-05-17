package raft

// RPCDelegate delegates Raft objects for RPC service
type RPCDelegate struct {
	Raft *Raft
}

func NewRPCDelegate(r *Raft) *RPCDelegate {
	return &RPCDelegate{r}
}

func (d *RPCDelegate) AppendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) error {
	d.Raft.AppendEntries(&arg, reply)
	return nil
}

func (d *RPCDelegate) RequestVote(arg RequestVoteArg, reply *RequestVoteReply) error {
	d.Raft.RequestVote(&arg, reply)
	return nil
}
