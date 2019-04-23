package main

import "log"

func check(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	} else {
		return false
	}
}
