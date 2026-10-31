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
	i := slices.IndexFunc(items, func(id string) bool {
		return matchRef(args, id, s.world.Items[id].Name)
	})
	if i == -1 || !s.world.Items[items[i]].Obtainable {
		return "ERR 404 ITEM_NOT_FOUND"
	}
	id := items[i]
	s.roomItems[p.Room] = slices.Delete(items, i, i+1)
	p.Inventory = append(p.Inventory, id)
	slog.Info("item taken", "player", p.Name, "item", id, "room", p.Room)
	return "OK taken=" + id
}
