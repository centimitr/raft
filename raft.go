package raft

import (
	"errors"
	"fmt"
	"time"
)

const (
	//DefaultRPCAddr = ":3456"
	DefaultRPCAddr = ""

	HeartbeatTimeout = 100 * time.Millisecond
)

type Config struct {
	RPCAddr string
}

type Events struct {
	OnRoleChange handler
}

type Raft struct {
	*State
	*Events
	Config       *Config
	Connectivity *Connectivity
	Election     *Election
	Quit         chan struct{}
}

func New(c Config) *Raft {
	return new(Raft).Init(&c)
}

func (r *Raft) Init(c *Config) *Raft {
	r.State = NewState()
	r.Config = c
	r.Connectivity = NewConnectivity()
	r.Election = NewElection(r)
	r.Quit = make(chan struct{})
	return r
}

func (r *Raft) BindStateMachine(stateMachine StateMachine) {
	r.Log = NewLog(stateMachine)
	r.Leader = newLeaderState(r.Log)
	stateMachine.Delegate(r)
}

func (r *Raft) setupConnectivity() (err error) {
	addr := DefaultString(r.Config.RPCAddr, DefaultRPCAddr)
	err = r.Connectivity.ListenAndServe("Raft", NewRPCDelegate(r), addr)
	return
}

func (r *Raft) Start() (err error) {
	if r.Log == nil {
		err = errors.New("raft: must bind a state machine before run")
		return
	}
	// callbacks have been in locked state
	r.Role.didSet(func() {
		log("role:", r.Role)
		switch r.Role.typ {
		case Follower:
			r.VotedFor = ""
		case Candidate:
			win := r.Election.Start()
			if win {
				r.Role.set(Leader)
			}
		case Leader:
			r.Leader.Reset()
			go func() {
				for {
					// todo: heartbeat
					r.callDeclareLeader()
					time.Sleep(HeartbeatTimeout)
				}
			}()
		}
	})
	// default start with a follower
	r.mu.Lock()
	r.Role.set(Follower)
	r.mu.Unlock()

	// connectivity
	err = r.setupConnectivity()
	if err != nil {
		err = fmt.Errorf("raft: %s", err)
		return
	}

	// start election timeout
	r.Election.ResetTimer()
	go func() {
		for {
			select {
			case <-r.Election.Timer.C:
				println("TIMEOUT")
				r.mu.Lock()
				r.Role.set(Candidate)
				r.mu.Unlock()
			case <-r.Quit:
				break
			}
		}
	}()
	return
}

func (r *Raft) Apply(command interface{}) (err error) {
	r.mu.RLock()
	role := r.Role
	term := r.CurrentTerm
	r.mu.RUnlock()
	// refuse if not a leader
	if !role.Is(Leader) {
		return errors.New("raft: Apply requires being a leader")
	}
	// create log append entry transaction
	tx, err := r.Log.append(term, command)
	println(r.Log.Entries)
	if err != nil {
		return
	}
	// call peers to append entries
	go r.callAppendEntries(tx)
	select {
	case err := <-tx.Cancel:
		return err
	case <-tx.Done:
		return
	}
}
