package raft

type Raft struct {
	*State
	Election *Election
	Quit     chan struct{}
}

func New() *Raft {
	return new(Raft).Init()
}

func (r *Raft) Init() *Raft {
	r.State = NewState()
	r.Election = NewElection(r.State)
	r.Quit = make(chan struct{})
	return r
}

func (r *Raft) Run() {
	for {
		select {
		case <-r.Election.Timer.C:
			r.Election.Start()
		case <-r.Quit:
			break
		}
	}
}
