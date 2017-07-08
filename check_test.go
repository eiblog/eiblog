package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckEmail(t *testing.T) {
	emails := []string{
		"xx@email.com",
		"xxxxemail.com",
		"xxx#email.com",
	}

	for i, v := range emails {
		if i == 0 {
			assert.True(t, CheckEmail(v))
		} else {
			assert.False(t, CheckEmail(v))
		}
	}
}

func TestCheckDomain(t *testing.T) {
	domains := []string{
		"123.com",
		"http://123.com",
		"https://123.com",
		"123#.com",
		"123.coooom",
	}

	for i, v := range domains {
		if i > 2 {
			assert.False(t, CheckDomain(v))
		} else {
			assert.True(t, CheckDomain(v))
		}
	}
}
