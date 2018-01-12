package main

import (
	"testing"
)

func TestDisqus(t *testing.T) {
	PostsCount()
}

func TestPostCreate(t *testing.T) {
	pc := &PostComment{
		Message:     "hahahaha",
		Thread:      "52799014",
		AuthorEmail: "deepzz.qi@gmail.com",
		AuthorName:  "deepzz",
	}

	id, err := PostCreate(pc)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("post success", id)
}

func TestThreadCreate(t *testing.T) {
	tc := &Article{
		Title: "测试test7",
		Slug:  "test7",
	}
	err := ThreadCreate(tc)
	if err != nil {
		t.Fatal(err)
	}
}
