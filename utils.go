package raft

import (
	l2 "log"
)

func log(vs ...interface{}) {
	l2.Println(vs...)
}

type handler func()

func (fn *handler) Handle() {
	if (*fn) != nil {
		(*fn)()
	}
}

type OnError func(err error)

func (fn *OnError) Check(err error) {
	if err != nil && (*fn) != nil {
		(*fn)(err)
	}
}

func DefaultString(s string, d string) string {
	if s != "" {
		return s
	} else {
		return d
	}
}
