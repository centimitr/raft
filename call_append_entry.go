package raft

type AppendEntriesArg struct {
	Term         Term
	LeaderID     NodeIndex
	PrevLogIndex LogEntryIndex
	PrevLogTerm  LogEntryIndex
	Entries      []LogEntry
	LeaderCommit LogEntryIndex
}

type AppendEntriesReply struct {
	CurrentTerm Term
	Success     bool
	// return conflict info to leader
	ConflictTerm  Term
	ConflictIndex LogEntryIndex
}

func (r *Raft) AppendEntries(arg *AppendEntriesArg, reply *AppendEntriesReply) {
	select {
	case <-r.shutdown:
		r.log("shutdown: AppendEntries RPC")
		return
	default:
	}

	r.log("AppendEntries: %d, term: %d", arg.LeaderID, arg.Term)
	r.mu.Lock()
	defer r.mu.Unlock()
	defer r.Store()

	// reject previous term request
	if arg.Term < r.CurrentTerm {
		reply.CurrentTerm = r.CurrentTerm
		return
	}

	// update term
	if r.CurrentTerm < arg.Term {
		r.CurrentTerm = arg.Term
		r.becomeFollower()
	}

	// for straggler (follower)
	if r.VotedFor != arg.LeaderID {
		r.VotedFor = arg.LeaderID
	}

	// reset election timer
	r.election.reset()

	// past heartbeat
	if arg.PrevLogIndex < r.Log.Snapshot.LastIncludedIndex {
		reply.CurrentTerm = r.CurrentTerm
		reply.ConflictTerm = r.Log.Snapshot.LastIncludedTerm
		reply.ConflictIndex = r.Log.Snapshot.LastIncludedIndex
		return
	}

	var prevLogIndex LogEntryIndex
	var preLogTerm Term
	if arg.PrevLogIndex < r.Log.len() {
		prevLogIndex = arg.PrevLogIndex
		preLogTerm = r.Log.retrieve(prevLogIndex).Term
	}

	// prev log entry match
	if prevLogIndex == arg.PrevLogIndex && preLogTerm == arg.PrevLogTerm {
		reply.Success = true

		// modify logs
		r.Logs = r.Logs[:r.Log.realIndex(prevLogIndex)+1]
		r.Logs = append(r.Logs, arg.Entries...)

		lastIndex := r.Log.len() - 1
		// update commitIndex
		if arg.LeaderCommit > r.Log.CommitIndex {
			r.Log.CommitIndex = min(arg.LeaderCommit, lastIndex)
			go r.commitCond.Broadcast()
		}

		// set conflict information
		reply.ConflictTerm = r.Log.retrieve(lastIndex).Term
		reply.ConflictIndex = lastIndex
		return
	}

	// if not match
	firstIndex := r.Log.firstIndex()
	reply.ConflictTerm = preLogTerm
	if reply.ConflictTerm == 0 {
		firstIndex = r.Log.len()
		reply.ConflictTerm = r.Log.retrieve(firstIndex - 1).Term
	} else {
		for i := prevLogIndex - 1; i >= r.Log.firstIndex(); i-- {
			if r.Log.retrieve(i).Term != preLogTerm {
				firstIndex = i + 1
				break
			}
		}
	}
	reply.ConflictIndex = firstIndex
}
