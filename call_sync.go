package raft

import "sort"

func (r *Raft) updateCommitIndex() {
	r.log("update commit index")
	indexes := make([]int, len(r.progress.MatchIndex))
	copy(indexes, r.progress.MatchIndex)

	sort.Ints(indexes)
	mid := indexes[r.peersCount/2]

	if r.Snapshot.LastIncludedIndex < mid && r.CommitIndex < mid && r.Log.retrieve(mid).Term == r.CurrentTerm {
		r.log("update commit index: %d -> %d", r.Log.CommitIndex, mid)
		r.Log.CommitIndex = mid
		go r.Log.commitCond.Broadcast()
	}
}

func (r *Raft) callSync(peerIndex LogEntryIndex) {
	r.mu.Lock()
	defer r.mu.Unlock()

	preLogIndex := r.progress.NextIndex[peerIndex] - 1

	// send Snapshot
	if preLogIndex < r.Snapshot.LastIncludedIndex {
		r.callSnapshot(peerIndex)
		return
	}

	// send entries
	var arg = &AppendEntriesArg{
		Term:         r.CurrentTerm,
		LeaderID:     r.Id,
		PrevLogIndex: preLogIndex,
		PrevLogTerm:  r.Log.retrieve(preLogIndex).Term,
		LeaderCommit: r.CommitIndex,
	}
	nextIndex := r.progress.NextIndex[peerIndex]
	logLength := r.Log.len()
	if nextIndex < logLength {
		entries := r.Logs[r.Log.realIndex(nextIndex):]
		arg.Entries = append(arg.Entries, entries...)
	}

	go func() {
		r.log("callSync: %d", peerIndex)
		var reply AppendEntriesReply
		err := r.peers[peerIndex].Call("Raft.AppendEntries", arg, &reply)
		if err != nil {
			return
		}
		r.handleSyncReply(peerIndex, &reply)
	}()
}

func (r *Raft) handleSyncReply(peerIndex NodeIndex, reply *AppendEntriesReply) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Role != Leader {
		return
	}

	// when success
	if reply.Success {
		r.progress.MatchIndex[peerIndex] = reply.ConflictIndex
		r.progress.NextIndex[peerIndex] = r.progress.MatchIndex[peerIndex] + 1
		r.updateCommitIndex()
		return
	}

	// when term check fail
	if r.Role == Leader && reply.CurrentTerm > r.CurrentTerm {
		r.log("later term found: leader -> follower")
		r.becomeFollower()
		r.Store()
		r.election.reset()
		return
	}

	// handle conflict
	foundIndex := -1
	if reply.ConflictTerm != 0 {
		for i := len(r.Logs) - 1; i > 0; i-- {
			if r.Logs[i].Term == reply.ConflictTerm {
				foundIndex = r.Log.realIndex(i)
				break
			}
		}
	}
	if foundIndex != -1 {
		r.progress.NextIndex[peerIndex] = min(foundIndex, reply.ConflictIndex)
	} else {
		r.progress.NextIndex[peerIndex] = reply.ConflictIndex
	}

	// send Snapshot or reset NextIndex
	lastSnapshotIndex := r.Log.Snapshot.LastIncludedIndex
	needSnapshot := lastSnapshotIndex != 0 && r.progress.NextIndex[peerIndex] <= lastSnapshotIndex
	if needSnapshot {
		r.callSnapshot(peerIndex)
	} else {
		nextIndex := r.progress.NextIndex[peerIndex]
		lowerBound := max(nextIndex, r.Log.firstIndex())
		r.progress.NextIndex[peerIndex] = min(lowerBound, r.len())
	}
}
