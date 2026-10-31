package main

import (
	"encoding/json"
)

func (s *Server) cmdInventory(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	data, _ := json.Marshal(append([]string{}, p.Inventory...))
	return "OK " + string(data)
}
