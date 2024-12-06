package lantas

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"path"
	"time"

	"github.com/bagaswh/lantas/pkg/config"
	"github.com/bagaswh/lantas/pkg/handler"
	"github.com/bagaswh/lantas/pkg/handler/chain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type handlerInfo struct {
	handler    handler.ConnHandler
	chainCount int
}

/* Server */
type Server struct {
	lantas *Lantas

	runtime   *config.Runtime
	config    *config.Server
	listeners []net.Listener
	stopCh    chan struct{}

	// conn handlers
	upstreamPreWrite, upstreamPostRead []handlerInfo
}

func (svr *Server) Start() {
	for _, ln := range svr.listeners {
		go svr.acceptLoop(ln)
	}
}

func (svr *Server) acceptLoop(ln net.Listener) {
	log.Info().Msgf("Server %s is ready to serve", ln.Addr().String())

	done := make(chan struct{})

	go func() {
		<-svr.stopCh

		if err := ln.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close listener")
		}
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		default:
			conn, acceptErr := ln.Accept()
			if acceptErr != nil {
				if errors.Is(acceptErr, net.ErrClosed) {
					return
				}
				log.Error().Err(acceptErr).Msg("failed accepting conn")
				continue
			}

			go svr.proxy(conn.(*net.TCPConn))
		}
	}
}

func (svr *Server) proxy(conn *net.TCPConn) {
	defer conn.Close()

	upstreamServer := svr.config.Upstreams[0]
	upstreamMgr := svr.lantas.upstreamManager
	upstreamConn := upstreamMgr.upstreams[upstreamServer]
	upstreamDialErr := upstreamConn.Dial()
	if upstreamDialErr != nil {
		log.Error().Err(upstreamDialErr).Str("upstream", upstreamServer).Msg("failed to dial upstream")
		return
	}
	defer upstreamConn.Close()

	done := make(chan struct{})

	// from upstream back to conn
	// go func() {
	// 	go svr.readLoop(upstreamConn, conn, done, svr.upstreamPostRead)
	// }()

	if len(svr.upstreamPostRead) == 0 {
		go func() {
			_, err := io.Copy(upstreamConn, conn)
			if err != nil {
				log.Debug().Err(err).Msg("copying from client to upstream errors")
			}
		}()
	} else {

	}

	if len(svr.upstreamPreWrite) == 0 {
		_, err := io.Copy(conn, upstreamConn)
		if err != nil {
			log.Debug().Err(err).Msg("copying from upstream to client errors")
		}
	} else {
		svr.readLoop(conn, upstreamConn, done, svr.upstreamPreWrite)
	}
}

func (svr *Server) readLoop(r, w net.Conn, done chan struct{}, handlers []handlerInfo) {

	bufpool := svr.lantas.bufferpool

	rbuf := bufpool.Get()
	rb := rbuf.Bytes()

	chainCount := 0
	for _, h := range handlers {
		chainCount += h.chainCount
	}
	bufChainSize := len(handlers)*chainCount + 1

	writeBufChain := make([]*bytes.Buffer, bufChainSize)
	for i := 0; i < bufChainSize; i++ {
		writeBufChain[i] = bufpool.Get()
	}

	ctx := handler.NewHandlerContext(r, rb, 0, nil, w, writeBufChain)

	for {
		select {
		case <-done:
			break
		default:
		}

		n, readErr := ctx.ReadConn.Read(ctx.ReadBytes)
		if readErr != nil {
			log.Error().Err(readErr).Msg("failed to read conn")
			return
		}
		ctx.ReadN = n
		ctx.ReadNBytes = ctx.ReadBytes[:n]
		var writebuf *bytes.Buffer
		for _, h := range handlers {
			err := h.handler.Handle(ctx)
			if err != nil {
				log.Error().Err(err).Msg("error handling connection")
				done <- struct{}{}
			}
		}
		writebuf = ctx.WriteBufChainCurrent
		ctx.WriteConn.SetWriteDeadline(time.Now().Add(30 * time.Second))
		upstreamWriteN, upstreamWriteErr := ctx.WriteConn.Write(writebuf.Bytes())
		_ = upstreamWriteN
		if upstreamWriteErr != nil {
			log.Error().Err(upstreamWriteErr).Str("upstream_addr", ctx.WriteConn.RemoteAddr().String()).Msgf("failed to write to upstream connection")
			// drop the fucking connection
			break
		}
		ctx.Reset()
	}
}

func (svr *Server) Stop() {
	svr.stopCh <- struct{}{}
}

/* Servers Manager */
type ServerManager struct {
	Servers []*Server
	logger  *zerolog.Logger
	rootDir string
}

func (sm *ServerManager) StartAllServers() error {
	for _, svr := range sm.Servers {
		svr.Start()
	}
	return nil
}

func newServerManager(lantas *Lantas, connHandler defaultHandler) (*ServerManager, error) {
	cfg := lantas.cfg
	rootDir := lantas.rootDir
	cfgServers := cfg.Servers
	servers := []*Server{}
	for i, svr := range cfgServers {
		var upstreamPreWrite, upstreamPostRead []handlerInfo
		if svr.Middlewares != nil && svr.Middlewares.Upstream != nil {

			if svr.Middlewares.Upstream.PreWrite != nil {
				preWrites := svr.Middlewares.Upstream.PreWrite
				for _, p := range preWrites {
					chain, chainErr := chain.New(lantas.middlewareChains[p.Name]...).Then(connHandler)
					if chainErr != nil {
						return nil, fmt.Errorf("failed building middleware chain for upstream prewrite: %w", chainErr)
					}
					handlerInfo := handlerInfo{
						handler:    chain,
						chainCount: len(lantas.middlewareChains[p.Name]) + 1,
					}
					upstreamPreWrite = append(upstreamPreWrite, handlerInfo)
				}
			} else {
				upstreamPreWrite = append(upstreamPreWrite, handlerInfo{
					handler:    defaultHandler{},
					chainCount: 1,
				})
			}

			if svr.Middlewares.Upstream.PostRead != nil {
				postReads := svr.Middlewares.Upstream.PostRead
				for _, p := range postReads {
					chain, chainErr := chain.New(lantas.middlewareChains[p.Name]...).Then(connHandler)
					if chainErr != nil {
						return nil, fmt.Errorf("failed building middleware chain for upstream postread: %w", chainErr)
					}
					handlerInfo := handlerInfo{
						handler:    chain,
						chainCount: len(lantas.middlewareChains[p.Name]) + 1,
					}
					upstreamPostRead = append(upstreamPostRead, handlerInfo)
				}
			} else {
				upstreamPostRead = append(upstreamPostRead, handlerInfo{
					handler:    defaultHandler{},
					chainCount: 1,
				})
			}

		}

		addrs := svr.Listen.Addresses
		server := Server{
			config:  svr,
			runtime: cfg,
			lantas:  lantas,

			upstreamPreWrite: upstreamPreWrite,
			upstreamPostRead: upstreamPostRead,
		}
		listeners := []net.Listener{}
		for j, addr := range addrs {
			ln, createLnErr := createListener(addr, false, svr.TLS, rootDir)
			if createLnErr != nil {
				return nil, fmt.Errorf("failed to create listener: %w (server index: %d, address index: %d)", createLnErr, i, j)
			}
			listeners = append(listeners, ln)
		}
		server.listeners = listeners
		servers = append(servers, &server)
	}
	svrMgr := &ServerManager{
		Servers: servers,
		rootDir: rootDir,
	}
	return svrMgr, nil
}

func createListener(address string, reusePort bool, tlsConfig *config.ServerTLS, rootDir string) (net.Listener, error) {
	tcpAddr, resolveAddrErr := net.ResolveTCPAddr("tcp", address)
	if resolveAddrErr != nil {
		return nil, fmt.Errorf("failed resolving TCP addr %s: %w", address, resolveAddrErr)
	}

	var ln net.Listener
	if tlsConfig != nil {
		certFile := tlsConfig.CertFile
		keyFile := tlsConfig.KeyFile
		if certFile == "" || keyFile == "" {
			return nil, errors.New("TLS configuration is invalid: either CertFile or KeyFile is empty")
		}
		cert, loadCertErr := tls.LoadX509KeyPair(path.Join(rootDir, certFile), path.Join(rootDir, keyFile))
		if loadCertErr != nil {
			return nil, fmt.Errorf("failed loading certificate %s and key %s", certFile, keyFile)
		}
		listenerTLSConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		tlsLn, listenErr := tls.Listen("tcp", tcpAddr.String(), listenerTLSConfig)
		if listenErr != nil {
			return nil, fmt.Errorf("failed to tls.Listen: %w", listenErr)
		}
		ln = tlsLn
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
