package compression

import (
	"fmt"

	"github.com/bagaswh/lantas/pkg/config"
	"github.com/bagaswh/lantas/pkg/handler"
)

func NewCompressionMiddleware(compression config.MiddlewareCompression, h handler.ConnHandler) (handler.ConnHandler, error) {
	var mdw handler.ConnHandler
	if compression.Algorithm == "gzip" {
		level := ""
		cfgLevel, ok := compression.Config[config.MiddlewareCompressionConfigKey_CompressionLevel]
		if ok {
			level = cfgLevel.(string)
		}
		var err error
		mdw, err = NewGzipMiddleware(level, h)
		if err != nil {
			return nil, fmt.Errorf("failed creating gzip middleware: %w", err)
		}
	} else {
		return nil, fmt.Errorf("invalid compression algorithm %q", compression.Algorithm)
	}

	return handler.HandlerFunc(func(c *handler.HandlerContext) error {
		return mdw.Handle(c)
	}), nil
}

func NewDecompressionMiddleware(compression config.MiddlewareCompression, h handler.ConnHandler) (handler.ConnHandler, error) {
	var mdw handler.ConnHandler
	if compression.Algorithm == "gzip" {
		var err error
		mdw = NewGunzipMiddleware(h)
		if err != nil {
			return nil, fmt.Errorf("failed creating gunzip middleware: %w", err)
		}
	} else {
		return nil, fmt.Errorf("invalid decompression algorithm %q", compression.Algorithm)
	}

	return handler.HandlerFunc(func(c *handler.HandlerContext) error {
		return mdw.Handle(c)
	}), nil
}
