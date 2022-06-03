package eventloop

import (
	"garry.org/data_structure/eventloop/internal"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type conn struct {
	fd         int              // file descriptor
	lnidx      int              // listener index in the server lns list
	out        []byte           // write buffer
	sa         syscall.Sockaddr // remote socket address
	reuse      bool             // should reuse input buffer
	opened     bool             // connection opened event fired
	action     Action           // next user action
	ctx        interface{}      // user-defined context
	addrIndex  int              // index of listening address
	localAddr  net.Addr         // local addr
	remoteAddr net.Addr         // remote addr
	loop       *loop            // connected loop
}

func (c *conn) Context() interface{}       { return c.ctx }
func (c *conn) SetContext(ctx interface{}) { c.ctx = ctx }
func (c *conn) AddrIndex() int             { return c.addrIndex }
func (c *conn) LocalAddr() net.Addr        { return c.localAddr }
func (c *conn) RemoteAddr() net.Addr       { return c.remoteAddr }
func (c *conn) Wake() {
	if c.loop != nil {
		c.loop.pool.Trigger(c)
	}
}

type server struct {
	events   Events             // user events
	loops    []*loop            // all the loops
	lns      []*listener        // all the listeners
	wg       sync.WaitGroup     // loop close waitgroup
	cond     *sync.Cond         // shutdown signal
	balance  LoadBalance        // load balancing method
	accepted uintptr            // accept counter
	tch      chan time.Duration // ticker channel
}

type loop struct {
	idx     int            // loop index in the server loops list
	pool    *internal.Poll // epoll or kqueue
	packet  []byte         // read packet buffer
	fdconns map[int]*conn  // loop connections fd -> conn
	count   int32          // connection count
}

// waitForShutdown waits for a signal to shutdown
func (s *server) waitForShutdown() {
	s.cond.L.Lock()
	s.cond.Wait()
	s.cond.L.Unlock()
}

// signalShutdown signals a shutdown an begins server closing
func (s *server) signalShutdown() {
	s.cond.L.Lock()
	s.cond.Signal()
	s.cond.L.Unlock()
}

func serve(events Events, listeners []*listener) error {

	numLoops := events.NumLoops
	if numLoops <= 0 {
		if numLoops == 0 {
			numLoops = 1
		} else {
			numLoops = runtime.NumCPU()
		}
	}

	s := &server{}
	s.events = events
	s.lns = listeners
	s.cond = sync.NewCond(&sync.Mutex{})
	s.balance = events.LoadBalance
	s.tch = make(chan time.Duration)

	// server starting
	if s.events.Serving != nil {
		var svr Server
		svr.NumLoops = numLoops
		svr.Addrs = make([]net.Addr, len(listeners))
		for i, ln := range listeners {
			svr.Addrs[i] = ln.lnaddr
		}
		action := s.events.Serving(svr)
		switch action {
		case None:
		case Shutdown:
			return nil
		}
	}

	defer func() {
		// wait on a singal for shutdown
		s.waitForShutdown()

		// notify all loops to close by closing all listeners
		for _, l := range s.loops {
			l.pool.Trigger(errClosing)
		}

		// wait on all loops to complete reading events
		s.wg.Wait()

		// close loops and all outstanding connections
		for _, l := range s.loops {
			for _, c := range l.fdconns {
				loopCloseConn(s, l, c, nil)
			}
			l.pool.Close()
		}
		// server stopped
	}()

	// create loops locally and bind the listeners
	for i := 0; i < numLoops; i++ {
		l := &loop{
			idx:     i,
			pool:    internal.OpenPoll(),
			packet:  make([]byte, 0xFFFF),
			fdconns: make(map[int]*conn),
		}
		for _, ln := range listeners {
			l.pool.AddRead(ln.fd)
		}
		s.loops = append(s.loops, l)
	}
	// start loops in backround
	s.wg.Add(len(s.loops))
	for _, l := range s.loops {
		go loopRun(s, l)
	}
	return nil
}

func loopCloseConn(s *server, l *loop, c *conn, err error) error {
	atomic.AddInt32(&l.count, -1)
	delete(l.fdconns, c.fd)
	syscall.Close(c.fd)
	if s.events.Closed != nil {
		switch s.events.Closed(c, err) {
		case None:
		case Shutdown:
			return errClosing
		}
	}
	return nil
}

func loopDetachConn(s *server, l *loop, c *conn, err error) error {
	if s.events.Detached == nil {
		return loopCloseConn(s, l, c, err)
	}
	l.pool.ModDetach(c.fd)

	atomic.AddInt32(&l.count, -1)
	delete(l.fdconns, c.fd)
	if err := syscall.SetNonblock(c.fd, false); err != nil {
		return err
	}
	switch s.events.Detached(c, &detachedConn{fd: c.fd}) {
	case None:
	case Shutdown:
		return errClosing
	}
	return nil
}

func loopRun(s *server, l *loop) {
	defer func() {
		s.signalShutdown()
		s.wg.Done()
	}()

	if l.idx == 0 && s.events.Tick != nil {
		go loopTicker(s, l)
	}

	l.poll.Wait(func(fd int, note interface{}) error {
		if fd == 0 {
			return loopNote(s, l, note)
		}
		c := l.fdconns[fd]
		switch {
		case c == nil:
			return loopAccept(s, l, fd)
		case !c.opened:
			return loopOpened(s, l, c)
		case len(c.out) > 0:
			return loopWrite(s, l, c)
		case c.action != None:
			return loopAction(s, l, c)
		default:
			return loopRead(s, l, c)
		}
	})
}
