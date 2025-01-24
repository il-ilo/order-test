package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestH(t *testing.T) {
	h := NewHistogram[int]([]int{10, 20, 30, 40, 50})
	for range 10000 {
		//h.Add(rand.Intn(60))
		h.Add(rand.Intn(20) + rand.Intn(20) + rand.Intn(20))
	}
	//h.Add(123)
	//h.Add(2)
	//h.Add(5)
	//h.Add(1)
	//h.Add(0)
	//h.Add(60)
	//h.Add(60)
	//h.Add(60)
	//h.Add(60)
	fmt.Println(h.Describe())
}
