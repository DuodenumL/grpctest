package testsuite

import (
	"fmt"
	"testing"
)

func TestCombine(t *testing.T) {
	c := [][]string{
		{"0", "1", "2"},
		{"a", "b", "c"},
	}
	fmt.Println(combine(c))
}
