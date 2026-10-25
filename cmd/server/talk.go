package main

import (
	"log/slog"
	"math/rand"
)

func (s *Server) cmdTalk(name, args string) string {
	query := normalize(args)
	if query == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	p, ok := s.players[name]
	if !ok {
		slog.Error("talk failed: unknown player", "player", name)
		return "ERR 500 INTERNAL_ERROR"
	}
	for _, id := range s.npcsInRoomLocked(p.Room) {
		npc := s.world.NPCs[id]
		if !matchRef(query, id, npc.Name) {
			continue
		}
		line := "..."
		if len(npc.Dialogue) > 0 {
			line = npc.Dialogue[rand.Intn(len(npc.Dialogue))]
		}
		slog.Info("npc talk", "player", name, "npc", id)
		return "OK " + cleanText(line)
	}
	return "ERR 404 NPC_NOT_FOUND"
}
