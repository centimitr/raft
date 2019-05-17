package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Message struct {
	Id int
	Content []string
}


func MakeRequest()  {
	m := Message{
		Id:1,
		Content: []string{"a","b"},
	}
	fmt.Println(m)

	b, err := json.Marshal(m)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(b)

	resp, err := http.Post("http://localhost:8080/","application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Fatalln(err)
	}

	b1, err := ioutil.ReadAll(resp.Body)
	fmt.Println(b1)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalln(err)
	}

	var result Message
	err = json.Unmarshal(b1, &result)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(result)
}

func main() {
	MakeRequest()
}