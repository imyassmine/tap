package main

import (
	"log/slog"
	"strconv"
	"strings"
	"unicode"
)

func cleanText(text string) string {
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, text)
	return strings.TrimSpace(text)
}

func (s *Server) cmdWho(args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return "OK players=" + strconv.Itoa(len(s.players))
}

func (s *Server) cmdChat(name, args string) string {
	scope, text, _ := strings.Cut(args, " ")
	scope = strings.ToUpper(scope)
	text = cleanText(text)
	if text == "" || (scope != "GLOBAL" && scope != "ROOM") {
		return "ERR 400 BAD_REQUEST"
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	p, ok := s.players[name]
	if !ok {
		slog.Error("chat failed: unknown player", "player", name)
		return "ERR 500 INTERNAL_ERROR"
	}
	evt := "EVT " + scope + " CHAT " + name + " " + text
	if scope == "GLOBAL" {
		s.pushAllLocked(evt)
	} else {
		s.pushRoomLocked(p.Room, "", evt)
	}
	return "OK"
}
