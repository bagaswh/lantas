package lantas

import (
	"fmt"
	"slices"

	"github.com/bagaswh/lantas/pkg/config"
	"github.com/bagaswh/lantas/pkg/connpool"
	"github.com/bagaswh/lantas/pkg/handler"
	"github.com/bagaswh/lantas/pkg/handler/chain"
	"github.com/bagaswh/lantas/pkg/middlewares/compression"
)

type defaultHandler struct{}

func (h defaultHandler) Handle(c *handler.HandlerContext) error {
	// does nothing
	return nil
}

func NewLantas(cfg *config.Runtime, rootDir string) *Lantas {
	return &Lantas{
		cfg:              cfg,
		rootDir:          rootDir,
		middlewareChains: make(map[string][]chain.Constructor),
	}
}

type connPools map[string]*connpool.ConnPool

type Lantas struct {
	cfg     *config.Runtime
	rootDir string
	stopCh  chan struct{}

	middlewareChains map[string][]chain.Constructor

	upstreamManager *upstreamManager

	conntrack *conntrack

	bufferpool *bufferpool
}

// createBufferPool create pool of 64 16 KiB buffers
// to be used at the beginning.
// TODO: make it configurable
func (l *Lantas) createBufferPool() {
	l.bufferpool = newBufferPool(60, 16384)
}

func (l *Lantas) Init() error {
	connHandler := defaultHandler{}

	// build middleware chains
	l.buildMiddlewareChains(connHandler)

	serversManager, srvMgrErr := newServerManager(l, connHandler)
	if srvMgrErr != nil {
		return fmt.Errorf("failed to create server manager: %w", srvMgrErr)
	}
	serversManager.StartAllServers()

	upstreamMgr := newUpstreamManager(l)
	l.upstreamManager = upstreamMgr

	l.createBufferPool()

	return nil
}

func (l *Lantas) buildMiddlewareChains(h handler.ConnHandler) error {
	for _, svr := range l.cfg.Servers {
		if svr.Middlewares == nil || svr.Middlewares.Upstream == nil {
			continue
		}
		handler := slices.Concat(svr.Middlewares.Upstream.PostRead, svr.Middlewares.Upstream.PreWrite)
		for _, svrMdw := range handler {
			mdw, ok := l.cfg.MiddlewareChains[svrMdw.Name]
			if !ok {
				return fmt.Errorf("cannot find middleware chain %s in the middleware_chains definition", svrMdw.Name)
			}
			err := l.buildMiddlewareChain(svrMdw.Name, mdw, h)
			if err != nil {
				return fmt.Errorf("failed building middleware chain %q: %w", svrMdw.Name, err)
			}
		}
	}

	return nil
}

func (l *Lantas) buildMiddlewareChain(name string, mdw *config.MiddlewareChain, h handler.ConnHandler) error {
	if _, alreadyExists := l.middlewareChains[name]; alreadyExists {
		return nil
	}
	ctors := make([]chain.Constructor, 0)
	for _, step := range mdw.Steps {
		if step.Compression != nil {
			ctors = append(ctors, func(h handler.ConnHandler) (handler.ConnHandler, error) {
				mdw, err := compression.NewCompressionMiddleware(*step.Compression, h)
				if err != nil {
					return nil, fmt.Errorf("failed creating compression middleware %q: %w", name, err)
				}
				return handler.HandlerFunc(func(c *handler.HandlerContext) error {
					err := mdw.Handle(c)
					if err != nil {
						return err
					}
					c.AdvanceBufChain()
					return nil
				}), nil
			})
		}
	}
	l.middlewareChains[name] = ctors
	return nil
}

func (l *Lantas) Wait() {
	<-l.stopCh
}
