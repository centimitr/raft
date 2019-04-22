package raft

type NodeRole int

const (
	Candidate NodeRole = iota
	Follower
	Leader
)

type NodeId string

type Term int
