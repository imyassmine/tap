package main

import (
	"log/slog"
	"strings"
)

func (s *Server) cmdMove(p *Player, args string) string {
	if args == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	from := p.Room
	target, ok := s.world.Rooms[from].Exits[strings.ToLower(args)]
	if !ok {
		return "ERR 301 NO_EXIT"
	}
	p.Room = target
	s.pushRoomLocked(from, p.Name, "EVT ROOM PRESENCE LEAVE "+p.Name)
	s.pushRoomLocked(target, p.Name, "EVT ROOM PRESENCE ENTER "+p.Name)
	slog.Info("player moved", "player", p.Name, "from", from, "to", target)
	return "OK room=" + target
}
