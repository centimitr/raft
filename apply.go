package raft

type ApplyMsg struct {
	Index   LogEntryIndex
	Command interface{}

	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

func NewApplyMsg(index LogEntryIndex, command interface{}) ApplyMsg {
	return ApplyMsg{
		Index:   index,
		Command: command,
	}
}
