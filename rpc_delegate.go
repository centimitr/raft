package raft

import "time"

// RPCDelegate delegates Raft objects for RPC service
type RPCDelegate struct {
	Raft *Raft
}

func (d *RPCDelegate) AppendEntries(arg AppendEntriesArg, reply *AppendEntriesReply) error {
	//delaySimulation(d.Raft)
	d.Raft.AppendEntries(&arg, reply)
	return nil
}

func (d *RPCDelegate) RequestVote(arg RequestVoteArg, reply *RequestVoteReply) error {
	//delaySimulation(d.Raft)
	d.Raft.RequestVote(&arg, reply)
	return nil
}

// delaySimulation applies delay on the node whose index is 4
func delaySimulation(r *Raft) {
	if r.Id == 4 {
		time.Sleep(500 * time.Millisecond)
	}
}

// delayedPeer is a wrapper for Peer/RPC that make outbound RPC call slower
type delayedPeer struct {
	Peer
	r *Raft
}

func (p *delayedPeer) Call(method string, arg interface{}, reply interface{}) error {
	delaySimulation(p.r)
	return p.Call(method, arg, reply)
}

func (r *Raft) applyInboundDelay() {
	for i, p := range r.peers {
		r.peers[i] = &delayedPeer{Peer: p, r: r}
	}
}
