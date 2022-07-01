package http2

// section 5.1
type StreamState int32

const (
	StateIdle StreamState = iota
	StateReservedLocal
	StateReservedRemote
	StateOpen
	StateHalfClosedLocal
	StateHalfClosedRemote
	StateClosed
)

type stream struct {
	id    uint32
	state StreamState

	// 5.3.1
	weight   uint8
	parent   *stream
	children map[uint32]*stream

	recvFlow *flowController
}
