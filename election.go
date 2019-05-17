package raft

import "time"

type election struct {
	VotedFor NodeIndex

	timer   *time.Timer
	resetCh chan struct{}
}

func (e *election) init(timeout time.Duration) {
	e.timer = time.NewTimer(timeout)
	e.resetCh = make(chan struct{})
}

func (e *election) reset() {
	e.resetCh <- struct{}{}
}
