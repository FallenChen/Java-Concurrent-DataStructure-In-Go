package http2

import (
	"fmt"
	"sync"
	"time"
)

type Conn struct {
	config *Config

	handshakeL        sync.Mutex
	handshakeComplete bool
	handshakeErr      error

	writeQueue *writeQueue

	*connState
}

type Config struct {

	// HandshakeTimeout specifies the duration for the handshake to complete
	HandshakeTimeout time.Duration

	// whether a server allows the client's
	// TLSVersion is lower than TLS 1.2
	AllowLowTLSVersion bool
}

type connState struct {
	server bool
}

type writeQueue struct {
	sync.Mutex
}

// HandshakeError represents connection handshake error
type HandshakeError string

func (e HandshakeError) Error() string {
	return fmt.Sprintf("http2: %s", string(e))
}

// Handshake runs the client or server handshake
// protocol if it has not yet been run.
// Most uses of this package need not call Handshake
// explicitly: the first Read or Write will call it automatically

func (c *Conn) Handshake() error {
	c.handshakeL.Lock()
	defer c.handshakeL.Unlock()

	if err := c.handshakeErr; err != nil {
		return err
	}

	if c.handshakeComplete {
		return nil
	}

	if timeout := c.config.HandshakeTimeout; timeout > 0 {
		errCh := make(chan error, 2)
		time.AfterFunc(timeout, func() {
			errCh <- HandshakeError("handshake timed out")
		})

		go func() {
			if c.server {
				errCh <- c.serverHandshake()
			} else {
				errCh <- c.clientHandshake()
			}
		}()
		c.handshakeErr = <-errCh
	} else {
		if c.server {
			c.handshakeErr = c.serverHandshake()
		} else {
			c.handshakeErr = c.clientHandshake()
		}
	}

	if err := c.handshakeErr; err != nil {
		switch err.(type) {
		case HandshakeError:
			c.close()
			return err
		}

		// todo
		// Clients and server MUST treat an invalid connection preface as a
		// connection error of type PROTOCOL_ERROR
	} else {
		c.handshakeComplete = true
	}
	return c.handshakeErr
}

func (c *Conn) close() error {

	return nil
}
