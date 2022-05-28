
package internal

type Poll struct {
	fd	int	// epoll fd
	wfd 	int	// wake  fd
	notes 	noteQueue
}

// OpenPoll ...
func OpenPoll() *Poll {

	l := new(Poll)
	p, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	l.fd = p

	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if e0 != 0 {
		syscall.Close(p)
		panic(err)
	}
	l.wfd = int(r0)
	l.AddRead(l.wfd)
	return l
}