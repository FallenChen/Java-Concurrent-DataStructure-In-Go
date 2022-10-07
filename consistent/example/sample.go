package main

import (
	"fmt"
	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash"
)

// implement consistent.Member interface
type myMember string

func (m myMember) String() string {
	return string(m)
}

// default hashing function
type hasher struct{}

func (h hasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func main() {
	cfg := consistent.Config{
		PartitionCount:    7,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            hasher{},
	}

	c := consistent.New(nil, cfg)

	// Add some members to the consistent hash table.
	// Add function calculates average load and distributes partitions over members
	node1 := myMember("node1.olric.com")
	c.Add(node1)

	node2 := myMember("node2.olric.com")
	c.Add(node2)

	key := []byte("my-key")
	// calculates partition id for the given key
	// partID := hash(key) % partitionCount
	// the partitions is already distributed among members by Add function.
	owner := c.LocateKey(key)
	fmt.Println(owner.String())
}
