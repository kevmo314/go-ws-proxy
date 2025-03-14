package main

import (
	"bytes"
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
	go func() {
		buf := make([]byte, 4096)
		var frame io.WriteCloser
	write:
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				break write
			}
			for m := 0; m < n; {
				if frame == nil {
					frame, err = ws.NewFrameWriter(ws.PayloadType)
					if err != nil {
						break write
					}
				}
				p := bytes.Index(buf[m:n], []byte("\n"))
				if p != -1 {
					// there is a newline in the buffer, write m to p and close the frame
					if _, err := frame.Write(buf[m : m+p]); err != nil {
						break write
					}
					if err := frame.Close(); err != nil {
						break write
					}
					m = p + 1
					frame = nil
				} else {
					// there is no newline in the buffer, write m to n and continue
					if _, err := frame.Write(buf[m:n]); err != nil {
						break write
					}
					m = n
				}
			}
		}
		if frame != nil {
			if err := frame.Close(); err != nil {
				return
			}
		}
		if err := ws.WriteClose(1000); err != nil {
			return
		}
	}()

	buf := make([]byte, 4096)
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
