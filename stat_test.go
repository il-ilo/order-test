package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestH(t *testing.T) {
	h := NewHistogram[int]([]int{10, 20, 30, 40, 50})
	for range 10000 {
		h.Add(rand.Intn(20) + rand.Intn(20) + rand.Intn(20))
	}
	fmt.Println(h.Describe())
}
