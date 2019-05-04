package raft

import (
	"fmt"
)

func NewAppendEntriesArg(state *State) *AppendEntriesArg {
	return &AppendEntriesArg{
		Term:     state.CurrentTerm,
		LeaderId: state.Id,
	}
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

func checkRespTerm(r *Raft, term Term) bool {
	if term > r.CurrentTerm {
		r.CurrentTerm = term
		return false
	}
	return true
}

func (r *Raft) callRequestVotes(v *Voting) {
	for _, peer := range r.Connectivity.Peers {
		go func(peer *Peer) {
			var reply RequestVotesReply
			// todo: create voting request
			args := NewRequestVotesArg(r.State)
			err := peer.Call("Raft.RequestVotes", args, &reply)
			// todo: check if vote response valid
			if err != nil {
				v.Fail(err)
				return
			}
			if reply.VoteGranted {
				v.Approve()
			} else {
				v.Reject()
			}
		}(peer)
	}
}

func (r *Raft) callDeclareLeader() {
	log("heartbeats")
	// todo: implement real callDeclareLeader
	tx := newLogTx()
	r.callAppendEntries(tx.Apply, tx.Cancel)
}

func (r *Raft) callAppendEntries(apply chan<- struct{}, cancel chan<- error) {
	arg := NewAppendEntriesArg(r.State)
	peers := make(chan *Peer, len(r.Connectivity.Peers))
	for _, peer := range r.Connectivity.Peers {
		peers <- peer
	}
	cnt := 0
	for peer := range peers {
		var reply AppendEntriesReply
		// todo: modify to concurrent call after debug
		err := peer.Call("Raft.AppendEntries", arg, &reply)
		if checkRespTerm(r, reply.Term) {
			r.Role.set(Follower)
			cancel <- fmt.Errorf("raft.callAppendEntries: currentTerm: %d, reply.term: %d", r.CurrentTerm, reply.Term)
			close(cancel)
		}
		if err != nil || !reply.Success {
			peers <- peer
		}
		// todo: check how to handle reply.Term
		cnt++
		if cnt >= len(r.Connectivity.Peers) {
			close(apply)
		}
	}
}
