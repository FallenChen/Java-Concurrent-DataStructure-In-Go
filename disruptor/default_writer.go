package disruptor

import "runtime"


const SpinMask = 1024*16 - 1

type DefaultWriter struct {
	written		*Cursor  // the ring buffer has been written up to this sequence
	upstream	Barrier	 // all of the readers have advanced up to this sequence
	capacity	int64
	previous	int64
	gate		int64
}

func NewWriter(written *Cursor, upstream Barrier, capacity int64) *DefaultWriter {
	return &DefaultWriter{
		upstream: upstream,
		written: written,
		capacity: capacity,
		previous: defaultCursorValue,
		gate: defaultCursorValue,
	}
}

func (this *DefaultWriter) Reserve(count int64) int64 {
	if count <= 0 {
		panic(ErrMinimumReservationSize)
	}
	this.previous += count
	for spin:=int64(0); this.previous- this.capacity > this.gate; spin++ {
		if spin & SpinMask == 0 {
			runtime.Gosched() //LockSupport.parkNanos(1L)
		}
		this.gate = this.upstream.Load()
	}
	return this.previous
}

func (this *DefaultWriter) Commit(_, upper int64) {
	this.written.Store(upper)
}

