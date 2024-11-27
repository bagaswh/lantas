package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func main() {
	raw := "ghuwe ganteng banget coy ngalahin suparno gundala"
	rawRW := bytes.NewBuffer([]byte(raw))
	compressedRW := new(bytes.Buffer)

	gzipWriter, err := gzip.NewWriterLevel(compressedRW, gzip.BestCompression)
	mustNotErr(err, "gzip.NewWriterLevel")

	_, err = io.Copy(gzipWriter, rawRW)
	mustNotErr(err, "io.Copy to gzipWriter")

	err = gzipWriter.Close()
	mustNotErr(err, "gzipWriter.Close")

	gunzipReader, err := gzip.NewReader(compressedRW)
	mustNotErr(err, "gzip.NewReader")
	defer gunzipReader.Close()

	fmt.Println("Compressed data:")
	_, err = io.Copy(os.Stdout, compressedRW)
	mustNotErr(err, "io.Copy compressed data")

	fmt.Println("\nDecompressed data:")
	// Copy decompressed data to stdout
	_, err = io.Copy(os.Stdout, gunzipReader)
	mustNotErr(err, "io.Copy decompressed data")
}

func mustNotErr(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
}
