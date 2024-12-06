package chain

import (
	"fmt"

	mdw "github.com/bagaswh/lantas/pkg/handler"
)

// from: https://github.com/justinas/alice/blob/master/chain.go

type Constructor func(mdw.ConnHandler) (mdw.ConnHandler, error)

type Chain struct {
	constructors []Constructor
}

func New(constructors ...Constructor) Chain {
	return Chain{append(([]Constructor)(nil), constructors...)}
}

func (c Chain) Then(h mdw.ConnHandler) (mdw.ConnHandler, error) {
	var err error
	for i := range c.constructors {
		h, err = c.constructors[len(c.constructors)-1-i](h)
		if err != nil {
			return nil, fmt.Errorf("failed building middleware chain: %w", err)
		}
	}

	return h, nil
}

func (c Chain) Append(constructors ...Constructor) Chain {
	newCons := make([]Constructor, 0, len(c.constructors)+len(constructors))
	newCons = append(newCons, c.constructors...)
	newCons = append(newCons, constructors...)

	return Chain{newCons}
}

func (c Chain) Extend(chain Chain) Chain {
	return c.Append(chain.constructors...)
}
