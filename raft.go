package raft

const (
	DefaultRPCAddr = ":3456"
)

type Config struct {
	RPCAddr string
}

type Raft struct {
	*State
	Config       *Config
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
	r.Election = NewElection(r.State)
	r.Quit = make(chan struct{})
	return r
}

func (r *Raft) SetupConnectivity(peerAddrs []string, onPeerConnectError OnError) (err error) {
	err = r.Connectivity.ListenAndServe(r, DefaultString(r.Config.RPCAddr, DefaultRPCAddr))
	if err != nil {
		return
	}
	go r.Connectivity.ConnectPeers(peerAddrs, onPeerConnectError)
	return
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
