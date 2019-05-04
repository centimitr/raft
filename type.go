package raft

import "github.com/google/uuid"

type RoleType string

const (
	Follower  RoleType = "follower"
	Candidate RoleType = "candidate"
	Leader    RoleType = "leader"
)

type onSet func()

func (fn *onSet) apply() {
	if (*fn) != nil {
		(*fn)()
	}
}

type Role struct {
	typ   RoleType
	onSet onSet
}

func (r *Role) Is(typ RoleType) bool {
	return r.typ == typ
}

func (r *Role) set(typ RoleType) {
	if r.typ != typ {
		r.typ = typ
		r.onSet.apply()
	}
}

func (r *Role) didSet(fn onSet) {
	r.onSet = fn
}

func (r Role) String() string {
	return string(r.typ)
}

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

func (t Term) EarlierThan(t2 Term) bool {
	return t < t2
}

func (t Term) LaterThan(t2 Term) bool {
	return t > t2
}

func (t Term) NotEarlierThan(t2 Term) bool {
	return t >= t2
}
