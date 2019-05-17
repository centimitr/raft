package raft

type InstallSnapshotArg struct {
	Term              Term
	LeaderID          NodeIndex
	LastIncludedIndex LogEntryIndex
	LastIncludedTerm  Term
	Snapshot          []byte
}

type InstallSnapshotReply struct {
	CurrentTerm Term
}

// after outer service perform a snapshot
func (r *Raft) DidSnapshot(index int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.Store()

	if r.CommitIndex < index || index <= r.Log.Snapshot.LastIncludedIndex {
		panic("debug: index invalid")
	}

	r.Logs = r.Logs[r.Log.realIndex(index):]

	r.Log.Snapshot.LastIncludedIndex = index
	r.Log.Snapshot.LastIncludedTerm = r.Logs[0].Term

	r.log("snapshot: %d %d", r.Log.Snapshot.LastIncludedIndex, r.Log.Snapshot.LastIncludedTerm)
}

func (r *Raft) InstallSnapshot(arg *InstallSnapshotArg, reply *InstallSnapshotReply) {
	select {
	case <-r.shutdown:
		r.log("shutdown: snapshot")
		return
	default:
	}

	r.log("InstallSnapshot: %d, term: %d", arg.LeaderID, arg.Term)
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.Store()

	reply.CurrentTerm = r.CurrentTerm

	// check term
	if arg.Term < r.CurrentTerm {
		return
	}

	// check snapshot index
	if arg.LastIncludedIndex <= r.Log.Snapshot.LastIncludedIndex {
		return
	}

	// reset election
	r.election.reset()

	r.Log.Snapshot.LastIncludedIndex = arg.LastIncludedIndex
	r.Log.Snapshot.LastIncludedTerm = arg.LastIncludedTerm
	r.Log.CommitIndex = r.Log.Snapshot.LastIncludedIndex
	r.Log.LastApplied = r.Log.Snapshot.LastIncludedIndex

	if arg.LastIncludedIndex >= r.Log.len()-1 {
		r.log("snapshot: receive complete log entries")
		r.Logs = []LogEntry{{Term: r.Log.Snapshot.LastIncludedTerm}}
	} else {
		r.log("snapshot: receive partial log entries")
		r.Logs = r.Logs[r.Log.realIndex(arg.LastIncludedIndex):]

	}
	r.apply <- ApplyMsg{Index: r.Log.Snapshot.LastIncludedIndex, UseSnapshot: true, Snapshot: arg.Snapshot}
}

func (r *Raft) callSnapshot(server int) {
	data, _ := r.StableStore.LoadSnapshot()
	arg := &InstallSnapshotArg{
		Term:              r.CurrentTerm,
		LastIncludedIndex: r.Log.Snapshot.LastIncludedIndex,
		LastIncludedTerm:  r.Log.Snapshot.LastIncludedTerm,
		LeaderID:          r.Id,
		Snapshot:          data,
	}
	handle := func(server int, reply *InstallSnapshotReply) {
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.Role != Leader {
			return
		}
		if r.CurrentTerm < reply.CurrentTerm {
			r.CurrentTerm = reply.CurrentTerm
			r.becomeFollower()
			return
		}
		r.progress.MatchIndex[server] = r.Log.Snapshot.LastIncludedIndex
		r.progress.NextIndex[server] = r.Log.Snapshot.LastIncludedIndex + 1
	}
	go func() {
		var reply InstallSnapshotReply
		err := r.peers[server].Call("Raft.InstallSnapshot", arg, &reply)
		if err != nil {
			r.log("Call: %s", err)
			return
		}
		handle(server, &reply)
	}()
}
