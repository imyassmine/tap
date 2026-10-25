package main

import (
	"log/slog"
)

func (s *Server) cmdLook(name, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	data, err := s.lookJSON(name)
	if err != nil {
		slog.Error("look failed", "player", name, "error", err)
		return "ERR 500 INTERNAL_ERROR"
	}
	return "OK " + string(data)
}