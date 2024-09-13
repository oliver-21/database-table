package main

import "testing"

func TestTime(t *testing.T) {
	for i := 0; i < 100; i++ {
		println(randTime("2006-01-02", 356*20))
	}
}
