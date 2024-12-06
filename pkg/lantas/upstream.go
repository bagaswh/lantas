package lantas

import (
	"net"

	"github.com/bagaswh/lantas/pkg/connpool"
)

type upstreamManager struct {
	lantas    *Lantas
	upstreams map[string]*upstreamConn
}

type upstreamConn struct {
	pool      *connpool.ConnPool
	address   string
	conntrack *conntrack
	net.Conn
}

func newUpstreamConn(address string, keepalive int, conntrack *conntrack) (*upstreamConn, error) {
	upstreamConn := &upstreamConn{
		address:   address,
		conntrack: conntrack,
	}
	if keepalive > 0 {
		var err error
		upstreamConn.pool, err = connpool.New(keepalive, func() (net.Conn, error) {
			conn, err := net.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}, func(c net.Conn) error {
			return c.Close()
		})
		if err != nil {
			return nil, err
		}
	}
	return upstreamConn, nil
}

func (c *upstreamConn) Dial() error {
	var conn net.Conn
	var err error
	if c.pool != nil {
		conn, err = c.pool.Get()
	} else {
		// no pool, create connection on demand
		conn, err = net.Dial("tcp", c.address)
		conn = newTrackedConnection(conn, c.conntrack)
	}
	if err != nil {
		return err
	}
	c.Conn = conn
	return nil
}

func (c *upstreamConn) Done() error {
	if c.Conn == nil {
		return nil
	}
	if c.pool != nil {
		return c.pool.Put(c.Conn)
	}
	return c.Conn.Close()
}

func newUpstreamManager(lantas *Lantas) *upstreamManager {
	m := &upstreamManager{
		upstreams: make(map[string]*upstreamConn),
	}

	for name, upstream := range lantas.cfg.Upstreams {
		upstreamConn, _ := newUpstreamConn(upstream.Servers[0], upstream.KeepAlive, lantas.conntrack)
		m.upstreams[name] = upstreamConn
	}

	return m
}
