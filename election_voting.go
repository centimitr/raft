package raft

type Voting struct {
	total             int
	approves, rejects int
	Done              chan struct{}
	Cancel            chan struct{}
	//Timeout           chan struct{}
}

func (v *Voting) Start(total int) {
	v.Done = make(chan struct{})
	//v.Timeout = make(chan struct{})
	v.Cancel = make(chan struct{})
	//time.AfterFunc(VotingTimeout, func() {
	//	close(v.Timeout)
	//})
	v.total = total
	v.approves, v.rejects = 0, 0
}

func (v *Voting) checkDone() {
	if v.approves+v.rejects >= v.total {
		close(v.Done)
	}
}

func (v *Voting) Approve() {
	v.approves++
	v.checkDone()
}

func (v *Voting) Reject() {
	v.rejects++
	v.checkDone()
}

func (v *Voting) Fail(err error) {
	// todo: may change log solution
	log("voting: fail:", err)
	v.total--
	v.checkDone()
}

func (v *Voting) Win() bool {
	return v.approves*2 > v.total
}
