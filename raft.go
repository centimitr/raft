package raft

import (
	"errors"
	"fmt"
	"net/rpc"
)

const (
	//DefaultRPCAddr = ":3456"
	DefaultRPCAddr = ""
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
	Log          *Log
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
	r.Election = NewElection(r.State, r.Connectivity)
	r.Quit = make(chan struct{})
	return r
}

func (r *Raft) BindStateMachine(stateMachine StateMachine) {
	r.Log = NewLog(stateMachine)
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
	err = r.setupConnectivity()
	if err != nil {
		err = fmt.Errorf("raft: %s", err)
		return
	}
	r.Election.Init()
	go func() {
		for {
			select {
			case <-r.Election.Timer.C:
				r.Election.Start()
			case <-r.Quit:
				break
			}
		}
	}()
	return
}

func (r *Raft) Apply(command interface{}) (err error) {
	tx, err := r.append(r.CurrentTerm, command)
	if err != nil {
		return
	}

	go func() {
		arg := NewAppendEntriesArg(r.State)
		peers := make(chan *rpc.Client, len(r.Connectivity.Peers))
		for _, peer := range r.Connectivity.Peers {
			peers <- peer
		}
		cnt := 0
		for peer := range peers {
			var reply AppendEntriesReply
			// todo: modify to concurrent call after debug
			err := peer.Call("Raft.AppendEntries", arg, &reply)
			if err != nil || !reply.Success {
				peers <- peer
			}
			// todo: check how to handle reply.Term
			cnt++
			if cnt >= len(r.Connectivity.Peers) {
				close(tx.Apply)
			}
		}
	}()
	<-tx.Done
	return
}
