package raft

import (
	"fmt"
	"time"
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

func NewAppendEntriesArg(r *State, peer *Peer) *AppendEntriesArg {
	arg := NewHeartbeatArg(r, peer)
	arg.Entries = r.Log.slice(r.Leader.NextIndex(peer.Id))
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

func checkRespTerm(r *Raft, term Term) (ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if term > r.CurrentTerm {
		r.CurrentTerm = term
		r.Role.set(Follower)
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
	//log("broadcast: heartbeats")
	for _, p := range r.Connectivity.Peers {
		var arg = NewHeartbeatArg(r.State, p)
		var reply AppendEntriesReply
		err := p.Call(methodAppendEntries, arg, &reply)
		// todo: err
		_ = err
	}
}

// callAppendEntries tries to append log entries to followers
func (r *Raft) callAppendEntries(tx *LogTx) {
	peersCount := r.Connectivity.PeersCount()
	// when there is no other node, try commit directly
	if peersCount == 0 {
		ok := r.UpdateCommitIndex(r.CurrentTerm, r.Connectivity.PeersCount(), r.Leader)
		if ok {
			close(tx.Done)
		}
	}
	peers := make(chan *Peer, peersCount)
	for _, peer := range r.Connectivity.Peers {
		peers <- peer
	}
	for p := range peers {
		// skip if a peer is disconnected
		if !r.Connectivity.HasConnectedPeer(p) {
			continue
		}
		// start call append
		arg := NewAppendEntriesArg(r.State, p)
		println("NEW APPEND")
		fmt.Printf("%+v\n", arg)
		var reply AppendEntriesReply
		// todo: modify to concurrent call after debug
		err := p.Call(methodAppendEntries, arg, &reply)
		// communication fail
		if err != nil {
			go func(p *Peer) {
				time.Sleep(50 * time.Millisecond)
				peers <- p
			}(p)
			continue
		}
		// if follower responds higher term, convert to follower
		if checkRespTerm(r, reply.Term) {
			tx.Cancel <- fmt.Errorf("raft.callAppendEntries: currentTerm: %d, reply.term: %d", r.CurrentTerm, reply.Term)
			close(tx.Cancel)
			break
		}
		// inconsistency
		if !reply.Success {
			r.Leader.DecreaseNextIndex(p.Id)
		} else
		//	append successfully
		{
			lastEntry := arg.Entries[len(arg.Entries)-1]
			r.Leader.Update(p.Id, lastEntry.Index+1, lastEntry.Index)
			ok := r.UpdateCommitIndex(arg.Term, len(r.Connectivity.Peers), r.Leader)
			if ok {
				close(tx.Done)
			}
		}
	}
}
