package main

import (
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

func (s *Server) cmdChat(p *Player, args string) string {
	scope, text, _ := strings.Cut(args, " ")
	scope = strings.ToUpper(scope)
	text = cleanText(text)
	if text == "" || (scope != "GLOBAL" && scope != "ROOM") {
		return "ERR 400 BAD_REQUEST"
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	evt := "EVT " + scope + " CHAT " + p.Name + " " + text
	if scope == "GLOBAL" {
		s.pushAllLocked(evt)
	} else {
		s.pushRoomLocked(p.Room, "", evt)
	}
	return "OK"
}
