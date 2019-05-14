package raft

import (
	"math/rand"
	"sync"
	"time"
)

type Election struct {
	r *Raft

	Timer      *time.Timer
	timerLock  sync.Mutex
	Processing bool

	currentVoting *Voting
}

func NewElection(r *Raft) *Election {
	return &Election{r: r}
}

// randomElectionTimeout returns a time duration between 150~300ms
func randomElectionTimeout() time.Duration {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return time.Duration(150+r.Intn(150)) * time.Millisecond
}

func (e *Election) ResetTimer() {
	e.timerLock.Lock()
	timeout := randomElectionTimeout()
	if e.Timer == nil {
		e.Timer = time.NewTimer(timeout)
	} else {
		e.Timer.Reset(timeout)
	}
	e.timerLock.Unlock()
}

func (e *Election) Abandon() {
	close(e.currentVoting.Cancel)
}

func (e *Election) Start() (win bool) {
	log("elect: start")
	e.Processing = true
	e.r.CurrentTerm++
	e.r.VotedFor = e.r.Id

	v := new(Voting)
	e.currentVoting = v
	v.Start(len(e.r.Connectivity.Peers) + 1)
	v.Approve()
	e.r.callRequestVotes(v)

	select {
	case <-v.Done:
		win = v.Win()
		log("elect: done", v.Win(), v.total, v.approves, v.rejects)
	//case <-v.Timeout:
	//	log("elect: timeout")
	//	time.Sleep(randomElectionInterval())
	//	win = e.Start()
	case <-v.Cancel:
		log("elect: cancel")
	}
	if !win {
		e.r.VotedFor = ""
		e.ResetTimer()
	}
	log("elect: end")
	e.Processing = false
	return
}
