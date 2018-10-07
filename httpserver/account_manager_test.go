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

func TestPassword(t *testing.T) {
	fmt.Println(GeneratePassword())
	fmt.Println(GeneratePassword())
	fmt.Println(GeneratePassword())
	fmt.Println(GeneratePassword())
	fmt.Println(GeneratePassword())
	t.Error("erro")
}
