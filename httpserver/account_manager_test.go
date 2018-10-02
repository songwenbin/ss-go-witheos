package httpserver

import (
	"fmt"
	"testing"
)

func TestRand(t *testing.T) {
	fmt.Println(RandomPort())
	fmt.Println(RandomPort())
	fmt.Println(RandomPort())
	fmt.Println(RandomPort())
	fmt.Println(RandomPort())
	fmt.Println(RandomPort())
	if RandomPort() != "" {
		t.Error("erro")
	}
}
