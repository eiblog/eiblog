// Package main provides ...
package main

import (
	"fmt"
	"testing"

	"github.com/eiblog/eiblog/setting"
)

func TestPing(t *testing.T) {
	sf := Superfeedr{URL: fmt.Sprintf("https://%s.superfeedr.com", setting.Conf.Superfeedr)}
	sf.PingFunc("https://deepzz.com/rss.html")
}
