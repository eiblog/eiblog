// Package setting provides ...
package setting

import (
	"encoding/json"
	"testing"
)

func TestInit(t *testing.T) {
	data, err := json.Marshal(Conf)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}
