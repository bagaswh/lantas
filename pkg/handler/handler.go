package handler

import (
	"bytes"
	"net"
)

type HandlerContext struct {
	ReadConn   net.Conn
	ReadBytes  []byte
	ReadN      int
	ReadNBytes []byte

	WriteConn     net.Conn
	writeBufChain []*bytes.Buffer

	WriteBufChainPrev    *bytes.Buffer
	WriteBufChainCurrent *bytes.Buffer
}

func NewHandlerContext(readConn net.Conn, readBytes []byte, readN int, readNBytes []byte, writeConn net.Conn, writeBufChain []*bytes.Buffer) *HandlerContext {
	h := &HandlerContext{
		ReadConn:      readConn,
		ReadBytes:     readBytes,
		ReadN:         readN,
		ReadNBytes:    readNBytes,
		WriteConn:     writeConn,
		writeBufChain: writeBufChain,
	}
	if len(h.writeBufChain) > 0 {
		h.WriteBufChainCurrent = h.writeBufChain[0]
	}
	return h
}

func (c *HandlerContext) AdvanceBufChain() {
	if c.WriteBufChainCurrent != nil {
		c.WriteBufChainPrev = c.WriteBufChainCurrent
	}
	if len(c.writeBufChain) > 1 {
		c.writeBufChain = c.writeBufChain[1:]
	}
	c.WriteBufChainCurrent = c.writeBufChain[0]
}

// Reset buffers.
func (c *HandlerContext) Reset() {
	for _, buf := range c.writeBufChain {
		if buf != nil {
			buf.Reset()
		}
	}
}

type HandlerFunc func(*HandlerContext) error

func (f HandlerFunc) Handle(c *HandlerContext) error {
	return f(c)
}

type ConnHandler interface {
	Handle(*HandlerContext) error
}
