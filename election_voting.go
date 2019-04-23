package raft

import (
	"sync"
	"time"
)

type Voting struct {
	Done    chan struct{}
	Timeout chan struct{}
	Cancel  chan struct{}
}

func (v *Voting) Request() {
	v.Done = make(chan struct{})
	v.Timeout = make(chan struct{})
	v.Cancel = make(chan struct{})
	time.AfterFunc(VotingTimeout, func() {
		close(v.Timeout)
	})

	var wg sync.WaitGroup
	// request votes
	wg.Wait()
	close(v.Done)
}

func (v *Voting) Win() bool {
	return false
}
