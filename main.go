package main

import (
	"flag"
	"log"
	"os"

	"golang.org/x/net/websocket"
	"golang.org/x/term"
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
	go func() {
		state, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err == nil {
			defer term.Restore(int(os.Stdin.Fd()), state)
		}
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				break
			}
			if _, err := ws.Write(buf[:n]); err != nil {
				break
			}
		}
	}()
	buf := make([]byte, 4096)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			break
		}
		if _, err := os.Stdout.Write(buf[:n]); err != nil {
			break
		}
	}
}
