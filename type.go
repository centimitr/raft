package raft

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type NodeId string

type Term int

func (t Term) NotEarlierThan(t2 Term) bool {
	return t >= t2
}
