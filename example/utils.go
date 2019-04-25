package main

import (
	"log"
	"strings"
)

func check(err error, vs ...string) bool {
	if err != nil {
		if len(vs) > 0 {
			log.Println(strings.Join(vs, ": ")+":", err)
		} else {
			log.Println(err)
		}
		return true
	} else {
		return false
	}
}
