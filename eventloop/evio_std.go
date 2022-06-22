package eventloop

import "errors"

var errClosing = errors.New("closing")
var errCloseConns = errors.New("close conns")
