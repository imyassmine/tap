package main

import (
	"log/slog"
	"math/rand"
	"slices"
)

func (s *Server) cmdTalk(p *Player, args string) string {
	if args == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	npcs := s.npcsInRoomLocked(p.Room)
	i := slices.IndexFunc(npcs, func(id string) bool {
		return matchRef(args, id, s.world.NPCs[id].Name)
	})
	if i == -1 {
		return "ERR 404 NPC_NOT_FOUND"
	}
	id := npcs[i]
	npc := s.world.NPCs[id]
	line := npc.Dialogue[rand.Intn(len(npc.Dialogue))]
	slog.Info("npc talk", "player", p.Name, "npc", id)
	return "OK " + cleanText(line)
}
