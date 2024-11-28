package run

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/bagaswh/lantas/pkg/config"
	"github.com/bagaswh/lantas/pkg/middlewares"
	"github.com/docker/go-units"
	"github.com/rs/zerolog"
)

/* conntrack */
type conntrack struct {
	mu    sync.RWMutex
	track map[net.Conn]struct{}
}

func (tr *conntrack) remove(c net.Conn) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	delete(tr.track, c)
}

func (tr *conntrack) add(c net.Conn) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.track[c] = struct{}{}
}

func (tr *conntrack) count() int {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return len(tr.track)
}

/* Server */
type Server struct {
	config    *config.Server
	listeners []*net.TCPListener
	stopCh    chan struct{}
	conntrack *conntrack
}

func (svr *Server) Start() {
	// middlewareChain :=
	middlewares := svr.config.Middlewares

	for _, ln := range svr.listeners {
		go svr.acceptLoop(ln)
	}
}

func (svr *Server) acceptLoop(ln *net.TCPListener) {
	b := make([]byte, 16*units.KiB)
	connBuf := bytes.NewBuffer(b)

	// connHandler :=

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			_, readErr := connBuf.ReadFrom(conn)
			// really, what do i want?
			// i want to first read the conn, put into conn buffer
			// then i want to throw the read data into a series of middlewares (middleware chain)
			// middlewares can mutate the buffer however they want
		}
	}
}

func (svr *Server) buildConnHandler() middlewares.ConnHandler {}

func (svr *Server) Wait() {
	<-svr.stopCh
}

// connHandler
type connHandler struct {
	buf    *bytes.Buffer
	leftC  *net.TCPConn
	rightC *net.TCPConn
}

func (h *connHandler) Handle() {

}

/* Servers Manager */
type ServerManager struct {
	Servers []*Server
	logger  *zerolog.Logger
}

func NewServerManager(cfg *config.Runtime) (*ServerManager, error) {
	cfgServers := cfg.Servers
	servers := []*Server{}
	for i, svr := range cfgServers {
		addrs := svr.Listen.Addresses
		server := Server{
			config: svr,
			conntrack: &conntrack{
				track: make(map[net.Conn]struct{}),
			},
		}
		listeners := []*net.TCPListener{}
		for j, addr := range addrs {
			ln, createLnErr := createListener(addr, false, svr.TLS)
			if createLnErr != nil {
				return nil, fmt.Errorf("failed to create listener: %w (server index: %d, address index: %d)", createLnErr, i, j)
			}
			listeners = append(listeners, ln)
		}
		server.listeners = listeners
	}
	svrMgr := &ServerManager{
		Servers: servers,
	}
	return svrMgr, nil
}

func createListener(address string, reusePort bool, tlsConfig *config.ServerTLS) (*net.TCPListener, error) {
	tcpAddr, resolveAddrErr := net.ResolveTCPAddr("tcp", address)
	if resolveAddrErr != nil {
		return nil, fmt.Errorf("failed resolving TCP addr %s: %w", address, resolveAddrErr)
	}

	var ln *net.TCPListener
	if tlsConfig != nil {
		certFile := tlsConfig.CertFile
		keyFile := tlsConfig.KeyFile
		if certFile == "" || keyFile == "" {
			return nil, errors.New("TLS configuration is invalid: either CertFile or KeyFile is empty")
		}
		cert, loadCertErr := tls.LoadX509KeyPair(certFile, keyFile)
		if loadCertErr != nil {
			return nil, fmt.Errorf("failed loading certificate %s and key %s", certFile, keyFile)
		}
		listenerTLSConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		tlsLn, listenErr := tls.Listen("tcp", tcpAddr.String(), listenerTLSConfig)
		if listenErr != nil {
			return nil, fmt.Errorf("failed to tls.Listen: %w", listenErr)
		}
		ln = tlsLn.(*net.TCPListener)
	} else {
		var listenErr error
		ln, listenErr = net.ListenTCP("tcp", tcpAddr)
		if listenErr != nil {
			return nil, fmt.Errorf("failed to net.Listen: %w", listenErr)
		}
	}

	if reusePort {
		// TODO: set SO_REUSEPORT sockopt
	}

	return ln, nil
}

/* Lantas */
func NewLantas(cfg *config.Runtime) *Lantas {
	return &Lantas{
		cfg: cfg,
	}
}

type Lantas struct {
	cfg *config.Runtime
}

func (l *Lantas) Init() error {
	serversManager, srvMgrErr := NewServerManager(l.cfg)
	if srvMgrErr != nil {
		return fmt.Errorf("failed creating server manager: %w", srvMgrErr)
	}

	// middlewareChains :=
}

func buildMiddlewareChain(cfg *config.MiddlewareChain) middlewares.Chain {
	steps := make([]*config.MiddlewareStep, len(cfg.Steps))
	for i, step := range cfg.Steps {
		steps[i] = step
	}
}

func buildConnHandlerFromMiddlewareStep(step *config.MiddlewareStep) middlewares.ConnHandler {
	if step.Compression != nil {
		// handler :=
	}
}
