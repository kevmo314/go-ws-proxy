package main

import (
	"flag"
	"io"
	"log"
	"os"

	"golang.org/x/net/websocket"
)

func main() {
	origin := flag.String("origin", "http://localhost/", "Origin header to send to destination")
	protocol := flag.String("protocol", "", "Protocol to use")
	flag.Parse()

	if len(os.Args) < 2 {
		panic("Usage: go-ws-proxy <destination> [--origin <origin>] [--protocol <protocol>]")
	}
	dst := os.Args[1]

	ws, err := websocket.Dial(dst, *protocol, *origin)
	if err != nil {
		log.Fatal(err)
	}
	go io.CopyBuffer(ws, os.Stdin, make([]byte, ws.MaxPayloadBytes))
	if _, err := io.CopyBuffer(ws, os.Stdin, make([]byte, ws.MaxPayloadBytes)); err != nil {
		log.Fatal(err)
	}
}
