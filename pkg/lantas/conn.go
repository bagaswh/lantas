package lantas

import (
	"net"
	"sync"
)

/* conntrack */
type conntrack struct {
	mu    sync.RWMutex
	track map[net.Conn]struct{}
}

func (tr *conntrack) remove(c net.Conn) {
	tr.mu.Lock()
	delete(tr.track, c)
	tr.mu.Unlock()
}

func (tr *conntrack) add(c net.Conn) {
	tr.mu.Lock()
	tr.track[c] = struct{}{}
	tr.mu.Unlock()
}

func (tr *conntrack) count() int {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return len(tr.track)
}

type trackedConnection struct {
	tracker *conntrack
	net.Conn
}

func newTrackedConnection(conn net.Conn, tracker *conntrack) *trackedConnection {
	tracker.add(conn)
	return &trackedConnection{
		tracker: tracker,
		Conn:    conn,
	}
}
