package raft

// followers' callSync progress
type progress struct {
	NextIndex  []LogEntryIndex
	MatchIndex []LogEntryIndex
}

func (p *progress) init(peersCount int) {
	p.NextIndex = make([]int, peersCount)
	p.MatchIndex = make([]int, peersCount)
}

// update make a node's process record to latest
func (p *progress) update(nodeIndex NodeIndex, index LogEntryIndex) {
	p.NextIndex[nodeIndex] = index + 1
	p.MatchIndex[nodeIndex] = index
}

// reset resets a node's consistency process
func (p *progress) reset(nodeIndex NodeIndex, nextIndex LogEntryIndex) {
	p.MatchIndex[nodeIndex] = 0
	p.NextIndex[nodeIndex] = nextIndex
}
