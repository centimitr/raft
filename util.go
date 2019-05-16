package raft

import (
	"fmt"
	"log"
)

var DEBUG = true

func debug(v ...interface{}) {
	log.Print("DEBUG:")
	log.Println(v...)
}

func (r *Raft) log(format string, v ...interface{}) {
	DEBUG = false
	//noinspection GoBoolExpressions
	if DEBUG {
		s := fmt.Sprintf(format, v...)
		log.Printf("[%d-%s]: %s\n", r.Id, r, s)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
