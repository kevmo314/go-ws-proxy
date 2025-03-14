package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:       func(r *http.Request) bool { return true },
	EnableCompression: true,
} // use default options
func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv (%d)", len(message))
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func TestMainProgram(t *testing.T) {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	go http.ListenAndServe("localhost:23487", nil)

	stdoutr, stdoutw := io.Pipe()
	stdinr, stdinw := io.Pipe()

	cmd := exec.Command("go", "run", "main.go", "ws://localhost:23487/echo")
	cmd.Stdout = stdoutw
	cmd.Stdin = stdinr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	reader := bufio.NewReader(stdoutr)

	stdinw.Write([]byte("hello world\n"))

	line, err := reader.ReadBytes('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte("hello world\n"), line) {
		t.Fatal("stdout mismatch")
	}

	for size := 1024; size <= 256*1024*1024; size *= 2 {
		buf := make([]byte, size)
		for i := 0; i < size; i++ {
			buf[i] = 'a'
		}
		buf[len(buf)-1] = '\n'

		if n, err := stdinw.Write(buf); err != nil || n != len(buf) {
			t.Fatal(err)
		}

		line, err = reader.ReadBytes('\n')
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, line) {
			t.Fatalf("stdout mismatch, want len=%d, got len=%d", len(buf), len(line))
		}
	}

	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
}
