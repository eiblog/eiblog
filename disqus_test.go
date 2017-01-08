package main

import (
	"testing"
)

func TestDisqus(t *testing.T) {
	PostsCount()
}

func TestPostComment(t *testing.T) {
	pc := &PostCreate{
		Message:     "hahahaha",
		Thread:      "52799014",
		AuthorEmail: "deepzz.qi@gmail.com",
		AuthorName:  "deepzz",
	}

	id := PostComment(pc)
	if id == "" {
		t.Error("post failed")
		return
	}
	t.Log("post success")
}
