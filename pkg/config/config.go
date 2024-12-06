package config

import (
	"fmt"
	"io"
	"os"

	"github.com/docker/go-units"
	"gopkg.in/yaml.v2"
)

type Upstream struct {
	Servers   []string `yaml:"servers"`
	KeepAlive int      `yaml:"keepalive"`
	TLS       bool     `yaml:"tls"`
}

type MiddlewareCompression struct {
	Algorithm string         `yaml:"algorithm"`
	Config    map[string]any `yaml:"config"`
}

const (
	MiddlewareCompressionConfigKey_CompressionLevel = "compressionLevel"
)

type MiddlewareDecompression struct {
	Algorithm string         `yaml:"algorithm"`
	Config    map[string]any `yaml:"config"`
}

type MiddlewareStep struct {
	Compression   *MiddlewareCompression   `yaml:"compression"`
	Decompression *MiddlewareDecompression `yaml:"decompression"`
}

type MiddlewareChain struct {
	Steps []*MiddlewareStep `yaml:"steps"`
}

type Upstreams map[string]*Upstream
type MiddlewareChains map[string]*MiddlewareChain
type Log struct {
	Level string `yaml:"level"`
}

type ServerListen struct {
	Addresses []string `yaml:"addresses"`
	ReusePort bool     `yaml:"reusePort"`
}

type ServerUpstreams []string
type ServerMiddleware struct {
	Name string `yaml:"name"`
}
type ServerMiddlewareUpstream struct {
	PreWrite []*ServerMiddleware `yaml:"prewrite"`
	PostRead []*ServerMiddleware `yaml:"postread"`
}
type ServerMiddlewares struct {
	Upstream *ServerMiddlewareUpstream `yaml:"upstream"`
}

type ServerTLS struct {
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type Server struct {
	Listen      *ServerListen      `yaml:"listen"`
	TLS         *ServerTLS         `yaml:"tls"`
	Upstreams   ServerUpstreams    `yaml:"upstreams"`
	Middlewares *ServerMiddlewares `yaml:"middlewares"`
}

type Runtime struct {
	Upstreams        Upstreams        `yaml:"upstreams"`
	MiddlewareChains MiddlewareChains `yaml:"middleware_chains"`
	Servers          []*Server        `yaml:"servers"`
	Log              *Log             `yaml:"log"`
}

func (rt *Runtime) Validate() error {
	if rt == nil {
		return makeConfigValidationError("where's the config?")
	}

	upstreams := rt.Upstreams
	for name, upstream := range upstreams {
		if len(upstream.Servers) > 1 {
			return makeConfigValidationError(fmt.Sprintf("'servers' field in upstreams[%s] can only accept one server.s", name))
		}
	}

	if rt.Servers == nil {
		return makeConfigValidationError("'servers' field is required")
	}
	servers := rt.Servers
	for i, svr := range servers {
		if svr.Listen == nil {
			return makeConfigValidationError(fmt.Sprintf("'listen' field in servers[%d] is not defined. Server must listen to at least one address.", i))
		}
		if len(svr.Listen.Addresses) == 0 {
			return makeConfigValidationError(fmt.Sprintf("'listen.addresses' field in servers[%d] is of length 0. Server must listen to at least one address.", i))
		}
		if len(svr.Upstreams) == 0 {
			return makeConfigValidationError(fmt.Sprintf("'upstreams' field in servers[%d] must have at least one upstream.", i))
		}
		svrUpstream := svr.Upstreams[0]
		if _, ok := rt.Upstreams[svrUpstream]; !ok {
			return makeConfigValidationError(fmt.Sprintf("specified upstream %s in servers[%d] does not exist in the global upstreams list.", svrUpstream, i))
		}
	}

	return nil
}

func makeConfigValidationError(msg string) error {
	return fmt.Errorf("%s", msg)
}

const maxConfigFileSize = 10 * units.MiB

func ReadConfig(file string) (*Runtime, error) {
	stat, statErr := os.Stat(file)
	if statErr != nil {
		return nil, fmt.Errorf("failed to stat() config file: %w", statErr)
	}
	isFile := !stat.IsDir()
	if !isFile {
		return nil, fmt.Errorf("provided config file %s is actually not a file", file)
	}
	if stat.Size() >= maxConfigFileSize {
		return nil, fmt.Errorf("config file size is larger than %s", units.HumanSize(maxConfigFileSize))
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed opening config file: %w", err)
	}
	defer f.Close()

	configBytes, readErr := io.ReadAll(f)
	if readErr != nil {
		return nil, fmt.Errorf("failed reading config file: %w", readErr)
	}

	rt := Runtime{}
	unmarshalErr := yaml.Unmarshal(configBytes, &rt)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed yaml unmarshalling config file: %w", unmarshalErr)
	}

	if err := rt.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &rt, nil
}
