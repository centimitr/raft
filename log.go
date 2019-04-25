package raft

type Command interface{}

type StateMachine interface {
	Apply(cmd Command)
}

type LogEntryIndex int

type LogEntry struct {
	Index   LogEntryIndex
	Term    Term
	Command Command
}

func newLogEntry(l *Log, term Term, cmd Command) *LogEntry {
	return &LogEntry{
		Index:   LogEntryIndex(len(l.Entries) + 1),
		Term:    term,
		Command: cmd,
	}
}

type Log struct {
	Entries     []*LogEntry
	CommitIndex LogEntryIndex
	LastApplied LogEntryIndex

	sm    StateMachine
	apply chan LogEntry
}

func NewLog(stateMachine StateMachine) *Log {
	return &Log{
		sm:    stateMachine,
		apply: make(chan LogEntry),
	}
}

func (l *Log) Run() () {
	go func() {
		for entry := range l.apply {
			l.sm.Apply(entry.Command)
			l.LastApplied = entry.Index
		}
	}()
}

func (l *Log) Append(term Term, cmd Command) (err error) {
	e := newLogEntry(l, term, cmd)
	// todo: save disk or make entries stable, committed
	l.Entries = append(l.Entries, e)
	// todo: stable storage error
	//if err != nil {
	//	return
	//}
	l.CommitIndex = e.Index
	go func() {
		l.apply <- *e
	}()
	return
}
