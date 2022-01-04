package httpx

import (
	"net"
)

func listen(addr string) net.Listener {
	ln, err := net.Listen("TCP", addr)
	if err != nil {
		return listenErr(err)
	}
	return ln
}

func listenErr(err error) net.Listener {
	return &errorListener{err: err}
}

type errorListener struct {
	err error
}

func (ln *errorListener) Accept() (net.Conn, error) {
	return nil, ln.err
}

func (ln *errorListener) Addr() net.Addr {
	// net.Listen failed. Return something non-nil in case callers
	// call Addr before Accept:
	return &net.TCPAddr{IP: net.IP{0, 0, 0, 0}, Port: 443}
}

func (ln *errorListener) Close() error {
	return ln.err
}
