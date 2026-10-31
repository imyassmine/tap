package main

import (
	"log/slog"
	"slices"
)

func (s *Server) cmdDrop(p *Player, args string) string {
	if args == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	i := slices.IndexFunc(p.Inventory, func(id string) bool {
		return matchRef(args, id, s.world.Items[id].Name)
	})
	if i == -1 {
		return "ERR 404 ITEM_NOT_IN_INVENTORY"
	}
	id := p.Inventory[i]
	p.Inventory = slices.Delete(p.Inventory, i, i+1)
	s.roomItems[p.Room] = append(s.roomItems[p.Room], id)
	slog.Info("item dropped", "player", p.Name, "item", id, "room", p.Room)
	return "OK dropped=" + id
}
