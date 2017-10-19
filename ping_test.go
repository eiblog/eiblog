// Package main provides ...
package main

import (
	"testing"

	"github.com/eiblog/eiblog/setting"
)

func TestSuperFeedr(t *testing.T) {
	sf := superfeedr{}
	sf.PingFunc("https://deepzz.com/rss.html")
}

func TestPingRPC(t *testing.T) {
	pr := pingRPC{
		MethodName: "weblogUpdates.extendedPing",
	}
	pr.Params.Param = [4]rpcValue{
		rpcValue{Value: Ei.BTitle},
		rpcValue{Value: "https://" + setting.Conf.Mode.Domain},
		rpcValue{Value: "https://deepzz.com/post/gdb-debug.html"},
		rpcValue{Value: "https://deepzz.com/rss.html"},
	}
	pr.PingFunc("https://deepzz.com/post/gdb-debug.html")
}
