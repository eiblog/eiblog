package main

import (
	"testing"
)

func TestCheckEmail(t *testing.T) {
	e := "xx@email.com"
	e1 := "xxxxemail.com"
	e2 := "xxx#email.com"

	t.Log(CheckEmail(e))
	t.Log(CheckEmail(e1))
	t.Log(CheckEmail(e2))
}

func TestCheckDomain(t *testing.T) {
	d := "123.com"
	d1 := "http://123.com"
	d2 := "https://123.com"
	d3 := "123#.com"
	d4 := "123.coooom"

	t.Log(CheckDomain(d))
	t.Log(CheckDomain(d1))
	t.Log(CheckDomain(d1))
	t.Log(CheckDomain(d1))
	t.Log(CheckDomain(d4))
}
