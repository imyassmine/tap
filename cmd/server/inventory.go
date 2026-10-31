package main

import (
	"encoding/json"
	"log/slog"
)

func (s *Server) cmdInventory(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	data, err := json.Marshal(append([]string{}, p.Inventory...))
	if err != nil {
		slog.Error("inventory failed", "player", p.Name, "error", err)
		return "ERR 500 INTERNAL_ERROR"
	}
	return "OK " + string(data)
}
