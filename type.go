package raft

import "github.com/google/uuid"

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type NodeId string

func (id *NodeId) String() string {
	return string(*id)
}

func (id *NodeId) IsEmptyOrEqualTo(id2 NodeId) bool {
	return *id == "" || *id == id2
}

func NewNodeId() NodeId {
	return NodeId(uuid.New().String())
}

type Term int

func (t Term) LaterThan(t2 Term) bool {
	return t > t2
}

func (t Term) NotEarlierThan(t2 Term) bool {
	return t >= t2
}
