package raft

import (
	"errors"
	"fmt"
	"time"
)

const (
	//DefaultRPCAddr = ":3456"
	DefaultRPCAddr = ""
	
	HeartbeatTimeout = 50 * time.Millisecond
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
	r.Role.didSet(func() {
		// todo: remove debug
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
	r.Role.set(Follower)
	err = r.setupConnectivity()
	r.Election.ResetTimer()
	if err != nil {
		err = fmt.Errorf("raft: %s", err)
		return
	}
	go func() {
		for {
			select {
			case <-r.Election.Timer.C:
				r.Role.set(Candidate)
			case <-r.Quit:
				break
			}
		}
	}()
	return
}

func (r *Raft) Apply(command interface{}) (err error) {
	tx, err := r.Log.append(r.CurrentTerm, command)
	if err != nil {
		return
	}
	go r.callAppendEntries(tx.Apply, tx.Cancel)
	select {
	case err := <-tx.Cancel:
		return err
	case <-tx.Done:
		return
	}
}
