package compression

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"

	"github.com/bagaswh/lantas/pkg/middlewares"
	"github.com/rs/zerolog/log"
)

type Gzip struct {
	buf        *bytes.Buffer
	gzipWriter gzip.Writer
	level      string
}

// Write writes b to gzip writer. The compressed bytes are stored in buf.
func (gz *Gzip) Write(b []byte) (int, error) {
	return gz.gzipWriter.Write(b)
}

func NewGzip(level string, bufferInitialSize int) *Gzip {
	compLevel, err := getGzipLevel(level)
	if err != nil {
		log.Error().Err(err).Msg("failed to get gzip level, using default (DefaultCompression)")
		compLevel = flate.DefaultCompression
	}
	b := make([]byte, 0, bufferInitialSize)
	buf := bytes.NewBuffer(b)
	gzipWriter, _ := gzip.NewWriterLevel(buf, compLevel)
	return &Gzip{
		level:      level,
		gzipWriter: *gzipWriter,
	}
}

func NewGzipMiddleware(gz *Gzip) middlewares.Constructor {
	return func(ch middlewares.ConnHandler) middlewares.ConnHandler {

	}
}

type Gunzip struct {
}

func getGzipLevel(level string) (int, error) {
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
