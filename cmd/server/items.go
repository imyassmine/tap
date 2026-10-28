package main

import (
	"encoding/json"
	"log/slog"
	"slices"
	"strings"
)

func matchRef(query, id, name string) bool {
	_, short, _ := strings.Cut(id, ".")
	return strings.EqualFold(query, id) ||
		strings.EqualFold(query, short) ||
		strings.EqualFold(query, name)
}

func normalize(args string) string {
	return strings.Join(strings.Fields(args), " ")
}

func (s *Server) cmdTake(p *Player, args string) string {
	query := normalize(args)
	if query == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	items := s.roomItems[p.Room]
	for i, id := range items {
		def := s.world.Items[id]
		if !matchRef(query, id, def.Name) {
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

func (s *Server) cmdDrop(p *Player, args string) string {
	query := normalize(args)
	if query == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for i, id := range p.Inventory {
		if !matchRef(query, id, s.world.Items[id].Name) {
			continue
		}
		p.Inventory = slices.Delete(p.Inventory, i, i+1)
		s.roomItems[p.Room] = append(s.roomItems[p.Room], id)
		slog.Info("item dropped", "player", p.Name, "item", id, "room", p.Room)
		return "OK dropped=" + id
	}
	return "ERR 404 ITEM_NOT_IN_INVENTORY"
}

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
