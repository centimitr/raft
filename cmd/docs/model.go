package main

import "encoding/json"

type DocEntry struct {
	String string
}

type Doc struct {
	Entries []DocEntry
}

func (d *Doc) String() string {
	b, _ := json.Marshal(d)
	return string(b)
}

func NewDoc(s string) *Doc {
	var doc Doc
	_ = json.Unmarshal([]byte(s), &doc)
	return &doc
}

func (d *Doc) Merge(d2 Doc) {

}
