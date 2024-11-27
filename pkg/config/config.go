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
}

type MiddlewareCompression struct {
	Algorithm string         `yaml:"algorithm"`
	Config    map[string]any `yaml:"config"`
}

type MiddlewareStep struct {
	Compression *MiddlewareCompression `yaml:"compression"`
}

type MiddlewareChain struct {
	Steps []*MiddlewareStep
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
	PreWrite *ServerMiddleware
	PostRead *ServerMiddleware
}
type ServerMiddlewares struct {
	Upstream *ServerMiddlewareUpstream `yaml:"upstream"`
}

type ServerTLS struct {
	CertFile    string             `yaml:"certFile"`
	KeyFile     string             `yaml:"keyFile"`
	Upstreams   ServerUpstreams    `yaml:"upstreams"`
	Middlewares *ServerMiddlewares `yaml:"middlewares"`
}

type Server struct {
	Listen *ServerListen `yaml:"listen"`
	TLS    *ServerTLS    `yaml:"tls"`
}

type Runtime struct {
	Upstreams        Upstreams        `yaml:"upstreams"`
	MiddlewareChains MiddlewareChains `yaml:"middleware_chains"`
	Servers          []*Server        `yaml:"servers"`
	Log              *Log             `yaml:"log"`
}

func (rt *Runtime) Validate() error {
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

	configBytes, readErr := io.ReadAll(f)
	if readErr != nil {
		return nil, fmt.Errorf("failed reading config file: %w", err)
	}

	var rt Runtime
	unmarshalErr := yaml.Unmarshal(configBytes, rt)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed yaml unmarshalling config file: %w", err)
	}

	return &rt, nil
}
