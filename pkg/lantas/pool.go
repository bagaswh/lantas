package lantas

import (
	"bytes"
	"sync"
)

type bufferpool struct {
	pool sync.Pool
}

func (p *bufferpool) Get() *bytes.Buffer {
	if tmp := p.pool.Get(); tmp != nil {
		return tmp.(*bytes.Buffer)
	}

	var res *bytes.Buffer
	return res
}

func (p *bufferpool) Put(x *bytes.Buffer) {
	p.pool.Put(x)
}

func newBufferPool(initSize, bufsize int) *bufferpool {
	p := &bufferpool{
		pool: sync.Pool{
			New: func() any {
				b := make([]byte, bufsize)
				buf := bytes.NewBuffer(b)
				return buf
			},
		},
	}
	for i := 0; i < initSize; i++ {
		b := p.Get()
		p.Put(b)
	}
	return p
}
