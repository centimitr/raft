package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Test struct {
	Id int
	Content []string
}

type ColorGroup struct {
	ID     int
	Name   string
	Colors []string
}

func main()  {
	m := Test{
		Id:1,
		Content: []string{"a", "b"},
	}

	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("error: ", err)
	}
	os.Stdout.Write(b)

	var m1 Test
	err1 := json.Unmarshal(b, &m1)
	if err1 != nil {
		fmt.Println("error: ", err1)
	}
	fmt.Println("\n", m1)
}