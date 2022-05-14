package eventloop

import (
	"io"
	"net"
	"os"
	"strings"
	"time"
)

// occurs after the completion of an event
type Action int 

const(
	None	Action = iota
	
	Detach

	Close

	Shutdown
)

type Options struct {

	TCPKeepAlive	time.Duration

	ReuseInputBuffer	bool
}

// Server represents a server context wichi provides information about the
// running server and has control functions for managing state
type Server struct {

	Addrs	[]net.Addr

	NumLoops	int
}

type Conn interface {
	// user-defined context
	Context()	interface{}

	SetContext(interface{})
	// the index of server address that was passed to the Serve call
	AddrIndex()	int

	LocalAddr()	net.Addr

	RemoteAddr()	net.Addr
	// Wake triggers a Data event for this connection
	Wake()
}

type LoadBalance	int

const (
	Random	LoadBalance = iota

	RoundRobin
	// the next accepted connection to the loop with 
	// the least number of active connections
	LeastConnections
)

// Events represents the server events for the Serve call
// Each event has an Action return value that is used manage the state
// of the connection and server
type Events struct {

	// Setting this to a value greater than 1 will effectively make
	// the server multithreaded for multi-core machines.Which means you must
	// take care wiht synchonizing memory between all event callbacks.
	NumLoops	int

	LoadBalance	LoadBalance

	Serving func(server Server) (action Action)

	Opened func(c Conn) (out []byte, opts Options, action Action)

	Closed func(c Conn, err error) (action Action)

	Detached func(c Conn, rwc io.ReadWriteCloser) (action Action)

	PreWrite func()

	Data func(c Conn, in []byte) (out []byte, action Action)

	Tick func() (delay time.Duration, action Action)
}

// InputStream is a helper type for managing input streams from inside
// the Data event.
type InputStream struct{ b []byte }

// Begin accepts a new packet and returns a working sequence of
// unprocessed bytes.
func (is *InputStream) Begin(packet []byte) (data []byte) {
	data = packet
	if len(is.b) > 0 {
		is.b = append(is.b, data...)
		data = is.b
	}
	return data
}

// End shifts the stream to match the unprocessed data.
func (is *InputStream) End(data []byte) {
	if len(data) > 0 {
		if len(data) != len(is.b) {
			is.b = append(is.b[:0], data...)
		}
	} else if len(is.b) > 0 {
		// empty is.b
		is.b = is.b[:0]
	}
}

type listener struct {
	ln      net.Listener
	lnaddr  net.Addr
	pconn   net.PacketConn
	opts    addrOpts
	f       *os.File
	fd      int
	network string
	addr    string
}

type addrOpts struct {
	reusePort bool
}


func parseAddr(addr string) (network, address string, opts addrOpts, stdlib bool) {
	network = "tcp"
	address = addr
	opts.reusePort = false
	if strings.Contains(address, "://") {
		network = strings.Split(address, "://")[0]
		address = strings.Split(address, "://")[1]
	}
	if strings.HasSuffix(network, "-net") {
		stdlib = true
		network = network[:len(network)-4]
	}
	q := strings.Index(address, "?")
	if q != -1 {
		for _, part := range strings.Split(address[q+1:], "&") {
			kv := strings.Split(part, "=")
			if len(kv) == 2 {
				switch kv[0] {
				case "reuseport":
					if len(kv[1]) != 0 {
						switch kv[1][0] {
						default:
							opts.reusePort = kv[1][0] >= '1' && kv[1][0] <= '9'
						case 'T', 't', 'Y', 'y':
							opts.reusePort = true
						}
					}
				}
			}
		}
		address = address[:q]
	}
	return
}

func Serve(events Events, addr ...string) error {

	var lns []*listener
	defer func() {
		for _, ln := range lns {
			ln.close()
		}
	}()
	var stdlib bool
	for _, addr := range addr {
		var ln listener
		var stdlibt bool
		ln.network, ln.addr, ln.opts, stdlibt = parseAddr(addr)
		if stdlibt {
			stdlib = true
		}
		if ln.network == "unix" {
			os.RemoveAll(ln.addr)
		}
		var err error
		if ln.network == "udp" {
			if ln.opts.reusePort {
				ln.pconn, err = reuseportListenPacket(ln.network, ln.addr)
			} else {
				ln.pconn, err = net.ListenPacket(ln.network, ln.addr)
			}
		} else {
			if ln.opts.reusePort {
				ln.ln, err = reuseportListen(ln.network, ln.addr)
			} else {
				ln.ln, err = net.Listen(ln.network, ln.addr)
			}
		}
		if err != nil {
			return err
		}
		if ln.pconn != nil {
			ln.lnaddr = ln.pconn.LocalAddr()
		} else {
			ln.lnaddr = ln.ln.Addr()
		}
		if !stdlib {
			if err := ln.system(); err != nil {
				return err
			}
		}
		lns = append(lns, &ln)
	}
	if stdlib {
		return stdserve(events, lns)
	}
	return serve(events, lns)
}
