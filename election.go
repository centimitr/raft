package raft

import (
	"math/rand"
	"net/rpc"
	"sync"
	"time"
)

var (
	ElectionTimeout = 10 * time.Second
	VotingTimeout   = 5 * time.Second
)

type Election struct {
	state        *State
	connectivity *Connectivity

	Timer      *time.Timer
	Processing bool

	currentVoting *Voting
}

func NewElection(state *State, connectivity *Connectivity) *Election {
	return &Election{
		state:        state,
		connectivity: connectivity,
	}
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

func (e *Election) announceBeingLeader() {
	var wg sync.WaitGroup
	// notify all
	wg.Wait()
}

func (e *Election) requestVotes(v *Voting) {
	for _, peer := range e.connectivity.Peers {
		go func(peer *rpc.Client) {
			var reply AppendEntriesReply
			// todo: create voting request
			args := &AppendEntriesArg{}
			err := peer.Call("Raft.AppendEntries", args, &reply)
			// todo: check if vote response valid
			if err != nil {
				v.Fail(err)
				return
			}
			if reply.Success {
				v.Approve()
			} else {
				v.Reject()
			}
		}(peer)
	}
}

func (e *Election) Start() {
	log("elect: start")
	e.Processing = true
	e.state.CurrentTerm++
	e.state.Role = Candidate

	v := new(Voting)
	e.currentVoting = v
	v.Start(len(e.connectivity.Peers))
	e.requestVotes(v)

	select {
	case <-v.Done:
		log("elect: done")
		if v.Win() {
			log("elect: win")
			e.state.Role = Leader
			go e.announceBeingLeader()
		}
	case <-v.Timeout:
		log("elect: timeout")
		time.Sleep(randomElectionInterval())
		e.Start()
	case <-v.Cancel:
		log("elect: cancel")
		e.state.Role = Follower
	}
	log("elect: end")
	e.Processing = false
}
