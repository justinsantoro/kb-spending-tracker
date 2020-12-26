package main

import (
	"fmt"
	"testing"
)

func TestSingleTagParsing(t *testing.T) {
	tlist := []string{"tag"}
	tags, ntags := parseTagInput(tlist)
	if ntags < 1 {
		t.Error("error parsing tags expected ntags 1 got", ntags)
	}
	if tags[0] != "tag" {
		t.Error("error parsing tags expected taglist ", tlist, "got", tags)
	}

	tlist = []string{"tag", "this", "is", "a", "note"}
	tags, note := parseTagsAndNote(tlist)
	if tags == nil {
		t.Error("unexpected tags == nil")
	}
	if note != "this is a note" {
		t.Error("error parsing note from taglist got", note, "expected: this is a note")
	}
}

func TestMultiTagParsing(t *testing.T) {
	tlist := []string{"tag1,", "tag2,", "tag3"}
	tags, ntags := parseTagInput(tlist)
	if ntags < 3 {
		t.Error("error parsing tags expected ntags 3 got", ntags)
	}
	if len(tags) != 3 {
		t.Error("unexpected tag length:", len(tags), "expected: 3")
	}
	for i, tag := range tags {
		if tag != fmt.Sprintf("tag%v", i+1) {
			t.Error("error parsing tags: unexpected value: ", tag)
			break
		}
	}

	tlist = []string{"tag1,", "tag2,", "tag3", "this", "is", "a", "note"}
	tags, note := parseTagsAndNote(tlist)
	if tags == nil {
		t.Error("unexpected tags == nil")
	}
	if len(tags) != 3 {
		t.Error("ParseMultiTagAndNote: unexpected tag length:", len(tags), "expected: 3")
	}
	for i, tag := range tags {
		if tag != fmt.Sprintf("tag%v", i+1) {
			t.Error("ParseMultiTagAndNote: error parsing tags: unexpected value: ", tag)
			break
		}
	}
	if note != "this is a note" {
		t.Error("error parsing note from taglist got", note, "expected: this is a note")
	}
}
