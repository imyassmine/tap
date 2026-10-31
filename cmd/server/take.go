package main

import (
	"log/slog"
	"slices"
	"strings"
)

func matchRef(query, id, name string) bool {
	return strings.EqualFold(query, id) || strings.EqualFold(query, name)
}

func (s *Server) cmdTake(p *Player, args string) string {
	if args == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	items := s.roomItems[p.Room]
	for i, id := range items {
		def := s.world.Items[id]
		if !matchRef(args, id, def.Name) {
			continue
		}
		if !def.Obtainable {
			return "ERR 404 ITEM_NOT_FOUND"
		}
		s.roomItems[p.Room] = slices.Delete(items, i, i+1)
		p.Inventory = append(p.Inventory, id)
		slog.Info("item taken", "player", p.Name, "item", id, "room", p.Room)
		return "OK taken=" + id
	}
	return "ERR 404 ITEM_NOT_FOUND"
}
