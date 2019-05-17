package main

import (
	"fmt"
	"log"
)

func PortString(port int) string {
	return fmt.Sprintf(":%d", port)
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
