package connpool

import (
	"fmt"
	"net"
	"sync"
	"testing"
)

var listener net.Listener

func putN(n int, pool *ConnPool, fn func(i int)) ([]net.Conn, error) {
	conns := make([]net.Conn, n)
	for i := 0; i < n; i++ {
		_, conn := net.Pipe()
		conns[i] = conn
		putErr := pool.Put(conn)
		if putErr != nil {
			return nil, putErr
		}
		fn(i)
	}
	return conns, nil
}

var newPipeConn = func() (net.Conn, error) {
	_, conn := net.Pipe()
	return conn, nil
}
var closePipeConn = func(conn net.Conn) error {
	return conn.Close()
}

func createPool() *ConnPool {
	pool, _ := New(4, newPipeConn, closePipeConn)
	return pool
}

func TestPut(t *testing.T) {
	pool := createPool()
	toPut := 10
	putN(toPut, pool, func(i int) {
		if pool.off != (i+1)%4 {
			t.Fatalf("pool.off is %d, expected %d after putting new conn into pool", pool.off, 1)
		}
		if pool.count != min(i+1, 4) {
			t.Fatalf("pool.count is %d, expected %d after putting new conn into pool", pool.count, 1)
		}
	})
}

func TestGet(t *testing.T) {
	pool := createPool()
	conn, err := pool.Get()
	if err != nil {
		t.Fatal(err)
	}
	if pool.off != 0 {
		t.Fatalf("pool.off is non zero: %d despite nothing is being put into the pool", pool.off)
	}
	if pool.count != 0 {
		t.Fatalf("pool.count is non zero: %d despite nothing is being put into the pool", pool.count)
	}

	fn := func(i int) {}

	putN(1, pool, fn)
	c2, _ := putN(1, pool, fn)
	conn, err = pool.Get()
	if err != nil {
		t.Fatal(err)
	}
	if conn == nil {
		t.Fatalf("conn retrieved after putting 2 conns in pool is nil")
	}
	if conn != c2[0] {
		t.Fatalf("last retrieved conn is not the equal value of last put conn")
	}
	if pool.off != 1 {
		t.Fatalf("pool.off is not 1, but %d", pool.off)
	}
	if pool.count != 1 {
		t.Fatalf("pool.count is not 1, but %d", pool.count)
	}

	putN(1, pool, fn)
	putN(1, pool, fn)
	clast2nd, _ := putN(1, pool, fn)
	clast, _ := putN(1, pool, fn)
	conn, err = pool.Get()
	if conn == nil {
		t.Fatalf("conn retrieved after putting 2 conns (2nd time) in pool is nil")
	}
	if conn != clast[0] {
		t.Fatalf("last retrieved conn is not the equal value of last put conn")
	}
	if pool.store[0] != nil {
		t.Fatalf("first index in the pool store should be nil by now, but actually got %v", pool.store[0])
	}
	if pool.off != 0 {
		t.Fatalf("pool.off is not 0, but %d", pool.off)
	}

	clast, _ = putN(1, pool, fn)
	if pool.off != 1 {
		t.Fatalf("pool.off is not 1, but %d", pool.off)
	}

	pool.Get()

	conn, err = pool.Get()
	if conn != clast2nd[0] {
		t.Fatalf("last retrieved conn is not the equal value of last put conn")
	}
	if pool.off != 3 {
		t.Fatalf("pool.off is not 3, but %d", pool.off)
	}
	if pool.count != 2 {
		t.Fatalf("pool.count is not 2, but %d", pool.count)
	}
}

func TestConcurrentGetPut(t *testing.T) {
	pool, _ := New(4, newPipeConn, closePipeConn)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, conn := net.Pipe()
			pool.Put(conn)
			pool.Get()
		}()
	}
	wg.Wait()
}

func TestConnectionCreationFailure(t *testing.T) {
	pool, _ := New(2, func() (net.Conn, error) {
		return nil, fmt.Errorf("connection creation failed")
	}, closePipeConn)

	_, err := pool.Get()
	if err == nil {
		t.Fatal("Expected error when connection creation fails")
	}
}

func TestNewPoolErrors(t *testing.T) {
	// Test nil New function
	_, err := New(4, nil, closePipeConn)
	if err == nil {
		t.Fatal("Expected error for nil New function")
	}
}
