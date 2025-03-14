package main

import (
	"bufio"
	"flag"
	"io"
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
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			if _, err := ws.Write(line[:len(line)-1]); err != nil {
				break
			}
		}
	}()

	buf := make([]byte, 16*4096)
read:
	for {
		frame, err := ws.NewFrameReader()
		if err != nil {
			break
		}
		frame, err = ws.HandleFrame(frame)
		if err != nil {
			break
		}
		if frame == nil {
			continue
		}
		for {
			n, err := frame.Read(buf)
			if err != nil {
				break
			}
			if _, err := os.Stdout.Write(buf[:n]); err != nil {
				break read
			}
		}
		if _, err := os.Stdout.Write([]byte("\n")); err != nil {
			break
		}
		if trailer := frame.TrailerReader(); trailer != nil {
			io.Copy(io.Discard, trailer)
		}
	}
}
