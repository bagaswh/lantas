package compression

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"

	"github.com/bagaswh/lantas/pkg/handler"
)

func NewGzipMiddleware(level string, h handler.ConnHandler) (handler.ConnHandler, error) {
	flateLevel, err := getFlateCompressionLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip middleware: %w", err)
	}
	return handler.HandlerFunc(func(c *handler.HandlerContext) error {
		writebuf := c.WriteBufChainCurrent
		if writebuf == nil {
			return fmt.Errorf("gzip middleware: write buffer, where the gzipp'ed data goes, is nil")
		}
		w, _ := gzip.NewWriterLevel(writebuf, flateLevel)
		data := c.ReadNBytes
		if c.WriteBufChainCurrent != nil {
			data = c.WriteBufChainCurrent.Bytes()
		}
		_, err := w.Write(data)
		if err != nil {
			w.Close()
			return fmt.Errorf("failed writing gzipped bytes into context WriteBuf: %w", err)
		}
		w.Close()
		return h.Handle(c)
	}), nil
}

func NewGunzipMiddleware(h handler.ConnHandler) handler.ConnHandler {
	return handler.HandlerFunc(func(c *handler.HandlerContext) error {
		writebuf := c.WriteBufChainCurrent
		readbuf := bytes.NewBuffer(c.ReadNBytes)
		r, err := gzip.NewReader(readbuf)
		if err != nil {
			return fmt.Errorf("failed creating gunzipper: %w", err)
		}
		_, err = writebuf.ReadFrom(r)
		if err != nil {
			return fmt.Errorf("failed reading from gzip reader into context WriteBuf: %w", err)
		}
		return h.Handle(c)
	})
}

func getFlateCompressionLevel(level string) (int, error) {
	switch level {
	case "NoCompression":
		return flate.NoCompression, nil
	case "BestSpeed":
		return flate.BestSpeed, nil
	case "BestCompression":
		return flate.BestCompression, nil
	case "HuffmanOnly":
		return flate.HuffmanOnly, nil
	default:
		return 0, fmt.Errorf("unknown compression level %s", level)
	}
}
