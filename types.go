package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

func toJsonString(x interface{}) (string, error) {
	var b = new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(x)
	return b.String(), err
}



func ActionString(amt USD) string {
	if amt < 0 {
		return "spent " + amt.Abs().String() + " on"
	}
	return "received " + amt.Abs().String() + " from"
}

type TagBalance struct {
	usrs  map[string]USD
	total USD
	tag   string
}

func NewTagBalance(tag string) *TagBalance {
	return &TagBalance{
		make(map[string]USD),
		USD(0),
		tag,
	}
}

func (tb *TagBalance) Add(usr string, bal USD) {
	tb.usrs[usr] = bal
	tb.total += bal
}

func (tb TagBalance) String() string {
	str := fmt.Sprintln(ActionString(tb.total), tb.tag)
	for usr, bal := range tb.usrs {
		percent := bal.InDollars() / tb.total.InDollars() * 100
		str += ">@" + usr + ": " + bal.Abs().String() + fmt.Sprintf(" (%.1f", percent) + "%%)\n"
	}
	return str
}
