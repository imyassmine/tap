package main

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"tap/internal/world"
)

const port = ":4242"
const greeting = "OK hello proto=1"
const maxLineSize = 1024

type Server struct {
	mtx     sync.Mutex
	players map[string]net.Conn
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	w, err := world.Load("data/world.json")
	if err != nil {
		slog.Error("failed to load world", "error", err)
		os.Exit(1)
	}
	slog.Info("world loaded", "rooms", len(w.Rooms), "items", len(w.Items), "npcs", len(w.NPCs))
	
	ln, err := net.Listen("tcp", port)
	if err != nil {
		slog.Error("failed to start server", "port", port, "error", err)
		os.Exit(1)
	}
	defer ln.Close()
	slog.Info("server listening", "port", port)

	srv := new(Server)
	srv.players = make(map[string]net.Conn)
	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("accept error", "error", err)
			continue
		}
		go srv.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	name := ""

	slog.Info("client connected", "remote", remote)
	defer func() {
		s.unregister(name)
		slog.Info("client disconnected", "remote", remote, "player", name)
	}()

	if err := send(conn, greeting); err != nil {
		return
	}

	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, maxLineSize)
	scanner.Buffer(buf, maxLineSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			slog.Warn("empty line", "remote", remote)
			continue
		}
		command, args, _ := strings.Cut(line, " ")
		command = strings.ToUpper(command)
		args = strings.TrimSpace(args)
		slog.Info("command received", "remote", remote, "player", name,
			"command", command, "args", args)

		if name == "" && command != "CONNECT" && command != "QUIT" {
			if err := send(conn, "ERR 202 NOT_AUTHENTICATED"); err != nil {
				return
			}
			continue
		}

		switch command {
		case "CONNECT":
			if name != "" || !validName(args) {
				if err := send(conn, "ERR 400 BAD_REQUEST"); err != nil {
					return
				}
			} else if !s.register(args, conn) {
				if err := send(conn, "ERR 201 NAME_IN_USE"); err != nil {
					return
				}
			} else {
				name = args
				slog.Info("player connected", "remote", remote, "player", name)
				if err := send(conn, "OK connected"); err != nil {
					return
				}
			}
		case "QUIT":
			_ = send(conn, "OK bye")
			return
		default:
			if err := send(conn, "ERR 400 BAD_REQUEST"); err != nil {
				return
			}
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("read error", "remote", remote, "error", err)
	}
}

func (s *Server) register(name string, conn net.Conn) bool {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, taken := s.players[name]; taken {
		return false
	}
	s.players[name] = conn
	return true
}

func (s *Server) unregister(name string) {
	if name == "" {
		return
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.players, name)
}

func validName(name string) bool {
	if name == "" || !utf8.ValidString(name) {
		return false
	}
	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func send(conn net.Conn, msg string) error {
	remote := conn.RemoteAddr().String()
	if _, err := io.WriteString(conn, msg+"\n"); err != nil {
		slog.Warn("send failed", "remote", remote, "error", err)
		return err
	}
	slog.Info("response sent", "remote", remote, "response", msg)
	return nil
}
