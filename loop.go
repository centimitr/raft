package raft

import "time"

func (r *Raft) heartbeatLoop() {
	for {
		// exit if not leader
		if !r.IsLeader() {
			return
		}
		// reset election timer
		r.election.reset()
		select {
		case <-r.shutdown:
			r.log("shutdown: heartbeatLoop")
			return
		default:
			r.forEachPeer(func(peerIndex NodeIndex) {
				go r.callSync(peerIndex)
			})
		}
		time.Sleep(r.Config.HeartbeatTimeout)
	}
}

func (r *Raft) electionLoop() {
	for {
		select {
		case <-r.shutdown:
			r.log("shutdown: electionLoop")
			return
		// reset election timer
		case <-r.election.resetCh:
			if !r.election.timer.Stop() {
				<-r.election.timer.C
			}
			r.election.timer.Reset(r.Config.ElectionTimeout)
		case <-r.election.timer.C:
			r.log("election: timeout -> restart")
			go r.callElection()
			r.election.timer.Reset(r.Config.ElectionTimeout)
		}
	}
}

func (r *Raft) applyLoop() {
	for {
		// cache log entries to apply
		var entries []LogEntry

		r.mu.Lock()
		for r.Log.LastApplied == r.Log.CommitIndex {
			r.Log.commitCond.Wait()
			select {
			case <-r.shutdown:
				r.mu.Unlock()
				r.log("shutdown: loop")
				close(r.apply)
				return
			default:
			}
		}

		lastApplied := r.Log.LastApplied
		commitIndex := r.Log.CommitIndex

		// read committed log entries that have not been applied
		if lastApplied < commitIndex {
			r.Log.LastApplied = commitIndex
			entries = make([]LogEntry, commitIndex-lastApplied)
			copy(entries, r.Logs[r.Log.realIndex(lastApplied+1):r.Log.realIndex(commitIndex)+1])
		}
		r.mu.Unlock()

		// emit these committed entries to apply channel
		for i, log := range entries {
			index := lastApplied + i + 1
			msg := NewApplyMsg(index, log.Command)
			r.log("apply: %v", msg)
			r.apply <- msg
		}
	}
}
