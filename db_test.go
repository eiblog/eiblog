package main

import (
	"testing"
)

func TestPageListBack(t *testing.T) {
	_, artcs := PageListBack(0, "", false, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(0, "", false, false, 2, 10)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(3, "", false, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(3, "19", false, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(0, "", false, true, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(0, "", true, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
}

func TestAddSerie(t *testing.T) {
	err := AddSerie("测试", "nothing", "这里是描述")
	if err != nil {
		t.Error(err)
	}
}
