package middlewares

// from: https://github.com/justinas/alice/blob/master/chain.go

type Constructor func(ConnHandler) ConnHandler

type Chain struct {
	constructors []Constructor
}

func NewChain(constructors ...Constructor) Chain {
	return Chain{append(([]Constructor)(nil), constructors...)}
}

func (c Chain) Then(h ConnHandler) ConnHandler {
	if h == nil {
		h = &DefaultConnHandler{}
	}

	for i := range c.constructors {
		h = c.constructors[len(c.constructors)-1-i](h)
	}

	return h
}

func (c Chain) ThenFunc(fn ConnHandler) ConnHandler {
	// This nil check cannot be removed due to the "nil is not nil" common mistake in Go.
	// Required due to: https://stackoverflow.com/questions/33426977/how-to-golang-check-a-variable-is-nil
	if fn == nil {
		return c.Then(nil)
	}
	return c.Then(fn)
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
