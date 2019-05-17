package raft

import (
	"errors"
)

type cmdType string

const (
	cmdGet cmdType = "get"
	cmdSet cmdType = "set"
)

var (
	ErrKVNotLeader = errors.New("kv: current node not a leader")
	ErrKeyNotExist = errors.New("kv: key not exist")
	ErrShutdown    = errors.New("kv: shutdown")
)

type cmd struct {
	Type  cmdType
	Key   string
	Value string
}

func (kv *KV) request(typ cmdType, k, v string) (value string, err error) {
	receipt := kv.raft.Apply(cmd{Type: typ, Key: k, Value: v})
	if receipt.Err != nil {
		err = receipt.Err
		return
	}

	notify := make(chan struct{})

	kv.mu.Lock()
	kv.notify[receipt.Index] = notify
	kv.mu.Unlock()

	select {
	case <-notify:
		if !kv.raft.CheckLeadership(receipt.Term) {
			err = ErrKVNotLeader
			return
		}
		if typ == cmdGet {
			kv.mu.Lock()
			var ok bool
			value, ok = kv.m[k]
			if !ok {
				err = ErrKeyNotExist
			}
			kv.mu.Unlock()
		}
	case <-kv.shutdown:
		err = ErrShutdown
	}
	return
}

func (kv *KV) Get(key string) (value string, err error) {
	return kv.request(cmdGet, key, "")
}

func (kv *KV) Set(key, value string) (err error) {
	_, err = kv.request(cmdGet, key, value)
	return
}
