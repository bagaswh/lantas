package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/rs/zerolog/log"
)

var (
	mode   = flag.String("mode", "gunzip", "client mode")
	listen = flag.String("l", ":7812", "listen addr")
)

func main() {
	flag.Parse()

	listener, listenErr := net.Listen("tcp", *listen)
	if listenErr != nil {
		log.Fatal().Err(listenErr).Msgf("failed to listen on addr %s", *listen)
	}

	buf := make([]byte, 8)
	outbuf := make([]byte, 8)
	_ = outbuf

	for {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			log.Error().Err(acceptErr).Msgf("failed to listen on addr %s", *listen)
			continue
		}
		go func() {
			for {
				n, readErr := conn.Read(buf)
				if readErr != nil {
					if readErr != io.EOF {
						log.Error().Err(acceptErr).Msgf("failed to listen on addr %s", *listen)
					}
					return
				}
				br := bytes.NewReader(buf[:n])
				gunzipReader, gunzipReaderErr := gzip.NewReader(br)
				if gunzipReaderErr != nil {
					log.Error().Err(gunzipReaderErr).Msgf("failed to create gzip reader, here's the raw content instead: ")
					fmt.Println(buf[:n])
					continue
				}
				gunzipReadN, gunzipReadErr := gunzipReader.Read(outbuf)
				if gunzipReadErr != nil {
					log.Error().Err(gunzipReaderErr).Msgf("failed to create read from gzip reader, here's the raw content instead: ")
					fmt.Println(buf[:n])
					continue
				}
				fmt.Println(string(outbuf[:gunzipReadN]))
			}
		}()
	}
}
