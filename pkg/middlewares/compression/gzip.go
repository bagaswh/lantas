package compression

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
)

type Gzip struct {
	buf   bytes.Buffer
	w     gzip.Writer
	level string
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

func NewGzip(level string, w io.Writer) *Gzip {
	compLevel, err := getGzipLevel(level)
	if err != nil {
		log.Error().Err(err).Msg("failed to get gzip level, using default (DefaultCompression)")
		compLevel = flate.DefaultCompression
	}
	gzipWriter, _ := gzip.NewWriterLevel(w, compLevel)
	return &Gzip{
		level: level,
		w:     *gzipWriter,
	}
}

type Gunzip struct {
}
