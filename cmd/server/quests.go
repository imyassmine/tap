package main

import (
	"encoding/json"
	"log/slog"
)

func (s *Server) cmdQuests(p *Player, args string) string {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	name := p.Name
	_, ok := s.players[name]
	if !ok {
		slog.Error("quests failed: unknown player", "player", name)
		return "ERR 500 INTERNAL_ERROR"
	}

	var questIDs []string
	for id := range s.world.Quests {
		questIDs = append(questIDs, id)
	}

	data, err := json.Marshal(questIDs)
	if err != nil {
		slog.Error("quests marshal failed", "player", name, "error", err)
		return "ERR 500 INTERNAL_ERROR"
	}

	return "OK " + string(data)
}

func (s *Server) cmdQuest(p *Player, args string) string {
	query := normalize(args)
	if query == "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	name := p.Name
	_, ok := s.players[name]
	if !ok {
		slog.Error("quest failed: unknown player", "player", name)
		return "ERR 500 INTERNAL_ERROR"
	}

	quest, ok := s.world.Quests[query]
	if !ok {
		return "ERR 404 QUEST_NOT_FOUND"
	}

	data, err := json.Marshal(quest)
	if err != nil {
		slog.Error("quest marshal failed", "player", name, "error", err)
		return "ERR 500 INTERNAL_ERROR"
	}

	return "OK " + string(data)
}
