package middlewares

import (
	"bytes"
	"net"
)

type ConnHandler interface {
	Handle(net.Conn, net.Conn)
}

type DefaultConnHandler struct {
	buf bytes.Buffer
}

func (h *DefaultConnHandler) Handle(l net.Conn, r net.Conn) {}

func (h *DefaultConnHandler) ResetBuffer() error {
	h.buf.Reset()
	return nil
}
