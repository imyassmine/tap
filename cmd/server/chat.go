package main

import (
	"log/slog"
	"strings"
	"unicode"
)

func (s *Server) cmdChat(p *Player, args string) string {
	scope, text, _ := strings.Cut(args, " ")
	text = cleanText(text)
	if text == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	evt := "EVT " + scope + " CHAT " + p.Name + " " + text
	switch scope {
	case "GLOBAL":
		s.pushAllLocked(evt)
	case "ROOM":
		s.pushRoomLocked(p.Room, "", evt)
	case "GROUP":
		if p.Group == nil {
			return "ERR 401 NOT_IN_GROUP"
		}
		s.pushGroupLocked(p.Group, evt)
	default:
		return "ERR 400 BAD_REQUEST"
	}
	return "OK"
}

func cleanText(text string) string {
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, text)
	return strings.TrimSpace(text)
}

func (s *Server) pushAllLocked(msg string) {
	for _, p := range s.players {
		p.push(msg)
	}
}

func (s *Server) pushRoomLocked(roomID, except, msg string) {
	for _, p := range s.players {
		if p.Room == roomID && p.Name != except {
			p.push(msg)
		}
	}
}

func (player *Player) push(msg string) {
	select {
	case player.out <- msg:
	default:
		slog.Warn("message dropped", "player", player.Name)
	}
}
