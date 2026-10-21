package main

import (
	"errors"
	"io"
	"log"
	"net"
	"strings"

	"tap/internal/protocol"
)

const (
	listenAddr = ":4242"
	greeting   = "OK hello proto=1" // RFC 42TAP
)

func main() {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s", listenAddr)

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handle(c) // one goroutine per client
	}
}


func handle(c net.Conn) {
	defer c.Close()

	if !send(c, greeting) {
		return
	}

	lr := protocol.NewLineReader(c, protocol.DefaultMaxLine)
	for {
		line, err := lr.ReadLine()
		if err != nil {
			if perr, ok := protocol.AsError(err); ok {
				send(c, protocol.ERR(perr))
			} else if !errors.Is(err, io.EOF) {
				log.Printf("read %s: %v", c.RemoteAddr(), err)
			}
			return
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		cmd, err := protocol.ParseCommand(line)
		if err != nil {
			perr, ok := protocol.AsError(err)
			if !ok {
				perr = protocol.ErrUnknownCommand
			}
			if !send(c, protocol.ERR(perr)) {
				return
			}
			continue
		}

		log.Printf("%s: %s %q (not implemented)", c.RemoteAddr(), cmd.Name, cmd.Args)
		if !send(c, protocol.ERR(protocol.ErrUnknownCommand)) {
			return
		}
	}
}

func send(c net.Conn, line string) bool {
	_, err := io.WriteString(c, protocol.Encode(line))
	return err == nil
}
