package main

import "github.com/bits-and-blooms/bloom/v3"

func main() {
	filter := bloom.NewWithEstimates(1000000, 0.01)

	filter.Add([]byte("Love"))

	if filter.Test([]byte("Love")) {
		println("Hit")
	} else {
		println("Not Hit")
	}
}
