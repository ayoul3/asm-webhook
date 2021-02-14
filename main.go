package main

import (
	"flag"

	"github.com/ayoul3/asm-webhook/server"
)

var tlsCrt, tlsKey string

func init() {
	flag.StringVar(&tlsCrt, "tls-crt", "./ssl/server.crt", "Path to the server certificate file")
	flag.StringVar(&tlsKey, "tls-key", "./ssl/key.pem", "Path to the private key file")
	flag.Parse()
}

func main() {
	server.Start(tlsCrt, tlsKey)
}
