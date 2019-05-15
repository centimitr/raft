package raft

import (
	"fmt"
)

const (
	methodAppendEntries = "Raft.AppendEntries"
	methodRequestVotes  = "Raft.RequestVotes"
)

func NewHeartbeatArg(r *State, peer *Peer) *AppendEntriesArg {
	idx := r.Leader.NextIndex(peer.Id) - 1
	r.mu.RLock()
	defer r.mu.RUnlock()
	return &AppendEntriesArg{
		Term:         r.CurrentTerm,
		LeaderId:     r.Id,
		PrevLogIndex: idx,
		PrevLogTerm:  r.Log.retrieve(idx).Term,
		LeaderCommit: r.Log.CommitIndex,
	}
}

func NewAppendEntriesArg(state *State, peer *Peer) *AppendEntriesArg {
	arg := NewHeartbeatArg(state, peer)
	arg.Entries = state.Log.slice(state.Leader.NextIndex(peer.Id))
	return arg
}

func NewRequestVotesArg(r *State) *RequestVotesArg {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return &RequestVotesArg{
		Term:        r.CurrentTerm,
		CandidateId: r.Id,
		// todo: log index and term
		LastLogIndex: 0,
		LastLogTerm:  0,
	}
}

func checkRespTerm(r *Raft, term Term) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if term > r.CurrentTerm {
		r.CurrentTerm = term
		return false
	}
	return true
}

// callRequestVotes requests votes from peers
func (r *Raft) callRequestVotes(v *Voting) {
	for _, peer := range r.Connectivity.Peers {
		go func(peer *Peer) {
			var reply RequestVotesReply
			args := NewRequestVotesArg(r.State)
			err := peer.Call(methodRequestVotes, args, &reply)
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

// callDeclareLeader sends heartbeats to followers to keep them followers
func (r *Raft) callDeclareLeader() {
	log("broadcast: heartbeats")
	for _, p := range r.Connectivity.Peers {
		var arg = NewHeartbeatArg(r.State, p)
		var reply AppendEntriesReply
		err := p.Call(methodAppendEntries, arg, &reply)
		// todo: err
		_ = err
	}
}

// callAppendEntries tries to append log entries to followers
func (r *Raft) callAppendEntries(apply chan<- struct{}, cancel chan<- error) {
	peers := make(chan *Peer, len(r.Connectivity.Peers))
	for _, peer := range r.Connectivity.Peers {
		peers <- peer
	}
	cnt := 0
	for peer := range peers {
		arg := NewAppendEntriesArg(r.State, peer)
		var reply AppendEntriesReply
		// todo: modify to concurrent call after debug
		err := peer.Call(methodAppendEntries, arg, &reply)
		if checkRespTerm(r, reply.Term) {
			r.mu.Lock()
			r.Role.set(Follower)
			r.mu.Unlock()
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
