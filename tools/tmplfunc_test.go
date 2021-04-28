// Package tools provides ...
package tools

import (
	"testing"
	"time"
)

func TestDateFormat(t *testing.T) {
	now := time.Now()
	layout := "2006-01-02 15:04:05"
	str := DateFormat(now, layout)
	t.Log(str)

	var err error
	TimeLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	str = DateFormat(now, layout)
	t.Log(str)
}

func TestParseInLocation(t *testing.T) {
	date := "2021-04-27 15:33"
	layout := "2006-01-02 15:04"
	tm, err := time.Parse(layout, date)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tm)

	TimeLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	tm, err = time.ParseInLocation(layout, date, TimeLocation)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tm.UTC())
}
