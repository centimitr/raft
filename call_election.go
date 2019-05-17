package raft

// reset all peers' callSync progress
func (r *Raft) resetProgress() {
	l := r.Log.len()
	r.forEachPeer(func(peerIndex NodeIndex) {
		r.progress.reset(peerIndex, l)
	})
	r.progress.MatchIndex[r.Id] = l - 1
}

func (r *Raft) callElection() {
	arg := new(RequestVoteArg)

	r.mu.Lock()
	r.VotedFor = r.Id
	r.CurrentTerm += 1
	r.Role = Candidate

	arg.Term = r.CurrentTerm
	arg.CandidateID = r.Id
	arg.LastLogIndex = r.Log.lastIndex()
	arg.LastLogTerm = r.Log.last().Term
	r.mu.Unlock()

	votes := 1
	handle := func(reply *RequestVoteReply) {
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.Role != Candidate {
			return
		}
		if reply.CurrentTerm > arg.Term {
			r.CurrentTerm = reply.CurrentTerm
			r.becomeFollower()
			r.Store()
			r.election.reset()
			return
		}
		if reply.VoteGranted {
			if votes >= r.peersCount/2 {
				r.log("become leader")
				r.Role = Leader
				r.resetProgress()
				go r.heartbeatLoop()
				return
			}
			votes++
		}
	}
	r.forEachPeer(func(peerIndex NodeIndex) {
		go func() {
			var reply RequestVoteReply
			p := r.peers[peerIndex]
			err := p.Call("Raft.RequestVote", arg, &reply)
			if err != nil {
				r.log("Call: %s", err)
				return
			}
			handle(&reply)
		}()
	})
}
