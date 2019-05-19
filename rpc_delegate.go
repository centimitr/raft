package raft

import "time"

// RPCDelegate delegates Raft objects for RPC service
type RPCDelegate struct {
	Raft *Raft
}

func NewRPCDelegate(r *Raft) *RPCDelegate {
	return &RPCDelegate{r}
}

func (d *RPCDelegate) AppendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) error {
	delaySimulation(d.Raft)
	d.Raft.AppendEntries(&arg, reply)
	return nil
}

func (d *RPCDelegate) RequestVote(arg RequestVoteArg, reply *RequestVoteReply) error {
	delaySimulation(d.Raft)
	d.Raft.RequestVote(&arg, reply)
	return nil
}

func delaySimulation(r *Raft) {
	if r.Id == 4 {
		time.Sleep(500 * time.Millisecond)
	}
}
