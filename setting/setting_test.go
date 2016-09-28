// Package setting provides ...
package setting

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	init()
	fmt.Printf("%v\n", *Conf)
}
