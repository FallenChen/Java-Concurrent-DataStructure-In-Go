package http2

import "sync"

type Conn struct {
	config	*Config

	writeQueue	*writeQueue
}

type Config struct {

	// whether a server allows the client's
	// TLSVersion is lower than TLS 1.2
	AllowLowTLSVersion	bool
}

type writeQueue struct {
	sync.Mutex
}