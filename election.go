package raft

import (
	"math/rand"
	"time"
)

type Election struct {
	r *Raft

	Timer      *time.Timer
	Processing bool

	currentVoting *Voting
}

func NewElection(r *Raft) *Election {
	return &Election{r: r}
}

func (e *Election) Init() {
	e.Timer = time.NewTimer(ElectionTimeout)
}

func (e *Election) ResetTimer() {
	e.Timer.Reset(ElectionTimeout)
}

// randomElectionInterval returns a time duration between 150~300ms
func randomElectionInterval() time.Duration {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return time.Duration(150+r.Intn(150)) * time.Millisecond
}

func (e *Election) Abandon() {
	close(e.currentVoting.Cancel)
}

func (e *Election) Start() (win bool) {
	log("elect: start")
	e.Processing = true
	e.r.CurrentTerm++

	v := new(Voting)
	e.currentVoting = v
	v.Start(len(e.r.Connectivity.Peers) + 1)
	v.Approve()
	e.r.callRequestVotes(v)

	select {
	case <-v.Done:
		log("elect: done")
		win = v.Win()
	case <-v.Timeout:
		log("elect: timeout")
		time.Sleep(randomElectionInterval())
		win = e.Start()
	case <-v.Cancel:
		log("elect: cancel")
	}
	log("elect: end")
	e.Processing = false
	return
}
