package main

import (
	"log/slog"
	"strings"
)

func (s *Server) cmdMove(name, args string) string {
	if args == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	p, ok := s.players[name]
	if !ok {
		slog.Error("move failed: unknown player", "player", name)
		return "ERR 500 INTERNAL_ERROR"
	}
	from := p.Room
	target, ok := s.world.Rooms[from].Exits[strings.ToLower(args)]
	if !ok {
		return "ERR 301 NO_EXIT"
	}
	p.Room = target
	s.pushRoomLocked(from, name, "EVT ROOM PRESENCE LEAVE "+name)
	s.pushRoomLocked(target, name, "EVT ROOM PRESENCE ENTER "+name)
	slog.Info("player moved", "player", name, "from", from, "to", target)
	return "OK room=" + target
}
