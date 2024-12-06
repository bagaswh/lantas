package connpool

import (
	"fmt"
	"net"
	"sync"
)

// TODO: make it lock-free by using atomic value
// for head and tail index

type ConnPool struct {
	max   int
	off   int
	count int

	store []net.Conn
	mu    sync.Mutex

	New   func() (net.Conn, error)
	Close func(net.Conn) error
}

func New(max int, newFn func() (net.Conn, error), closeFn func(net.Conn) error) (*ConnPool, error) {
	p := &ConnPool{
		max:   max,
		store: make([]net.Conn, max),

		New:   newFn,
		Close: closeFn,
	}
	if newFn == nil {
		return nil, fmt.Errorf("New() function requires to be set")
	}
	return p, nil
}

func (p *ConnPool) create() (net.Conn, error) {
	if p.New == nil {
		return nil, fmt.Errorf("New() function requires to be set")
	}
	conn, err := p.New()
	if err != nil {
		return nil, fmt.Errorf("failed creating new pool item: %w", err)
	}
	return conn, nil
}

func (p *ConnPool) Put(c net.Conn) error {
	var lru net.Conn

	p.mu.Lock()
	if p.store[p.off] != nil {
		// we'll remove and close lru conn later
		lru = p.store[p.off]
	}
	p.store[p.off] = c
	p.off = (p.off + 1) % (p.max)
	if p.count < p.max {
		p.count++
	}
	p.mu.Unlock()

	// Close() outside the critical section so locking time is minimized
	if lru != nil && p.Close != nil {
		err := p.Close(lru)
		if err != nil {
			return err
		}
	}

	return nil
}

//

func (p *ConnPool) Get() (net.Conn, error) {
	p.mu.Lock()

	var conn net.Conn
	if p.count != 0 {
		off := (p.off - 1 + p.max) % p.max
		conn = p.store[off]
		p.store[off] = nil
		p.off = off
		p.count--
		p.mu.Unlock()
		return conn, nil
	}

	p.mu.Unlock()

	// connect() outside the critical section so locking time is minimized
	var err error
	conn, err = p.create()
	if err != nil {
		return nil, err
	}
	return conn, nil
}
