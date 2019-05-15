package raft

//
//import (
//	"sync"
//)
//
//type kvCommandType int
//
//const (
//	set kvCommandType = iota
//	remove
//)
//
//type kvCommand struct {
//	command kvCommandType
//	key     interface{}
//	value   interface{}
//}
//
//func newKVCommand(typ kvCommandType, k interface{}, v interface{}) *kvCommand {
//	return &kvCommand{typ, k, v}
//}
//
//type KV struct {
//	m        sync.Map
//	delegate StateMachineDelegate
//}
//
//func (kv *KV) apply(cmd *kvCommand) {
//	switch cmd.command {
//	case set:
//		kv.m.Store(cmd.key, cmd.value)
//	case remove:
//		kv.m.Delete(cmd.key)
//	}
//}
//
//func (kv *KV) Apply(command interface{}) {
//	// todo: check if panic is wanted
//	cmd := command.(*kvCommand)
//	kv.apply(cmd)
//}
//
//func (kv *KV) Delegate(d StateMachineDelegate) {
//	kv.delegate = d
//}
//
//func (kv *KV) Get(key interface{}) (interface{}, bool) {
//	return kv.m.Load(key)
//}
//
//func (kv *KV) modify(typ kvCommandType, key interface{}, value interface{}) (err error) {
//	cmd := newKVCommand(typ, key, value)
//	if kv.delegate != nil {
//		err = kv.delegate.Apply(cmd)
//	} else {
//		kv.apply(cmd)
//	}
//	return
//}
//
//func (kv *KV) Set(key interface{}, value interface{}) error {
//	return kv.modify(set, key, value)
//}
//
//func (kv *KV) Remove(key interface{}) error {
//	return kv.modify(remove, key, nil)
//}
//
//func (kv *KV) MustGet(key interface{}) interface{} {
//	v, _ := kv.Get(key)
//	return v
//}
//
//func (kv *KV) GetDefault(key interface{}, value interface{}) (v interface{}, err error) {
//	v, ok := kv.Get(key)
//	if !ok {
//		err = kv.Set(key, value)
//		v = value
//	}
//	return
//}
