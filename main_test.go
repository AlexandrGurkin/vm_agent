package main

import (
	"fmt"
	"testing"
)

func TestHostName(t *testing.T) {
	res := hostName()
	fmt.Println(res)
}

func TestNets(t *testing.T) {
	res := networks()
	fmt.Println(res)
}
