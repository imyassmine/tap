package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
)

type lookReply struct {
	Room    roomInfo `json:"room"`
	Players []string `json:"players"`
	Items   []string `json:"items"`
	NPCs    []string `json:"npcs"`
}

type roomInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
}

func (s *Server) cmdLook(name, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	data, err := s.lookJSON(name)
	if err != nil {
		slog.Error("look failed", "player", name, "error", err)
		return "ERR 500 INTERNAL_ERROR"
	}
	return "OK " + string(data)
}

func (s *Server) lookJSON(name string) ([]byte, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	player, ok := s.players[name]
	if !ok {
		return nil, fmt.Errorf("unknown player %q", name)
	}
	room := s.world.Rooms[player.Room]
	players := []string{}
	for _, other := range s.players {
		if other.Room == player.Room {
			players = append(players, other.Name)
		}
	}
	slices.Sort(players)
	exits := room.Exits
	if exits == nil {
		exits = map[string]string{}
	}
	return json.Marshal(lookReply{
		Room: roomInfo{
			ID:          player.Room,
			Name:        room.Name,
			Description: room.Description,
			Exits:       exits,
		},
		Players: players,
		Items:   append([]string{}, s.roomItems[player.Room]...),
		NPCs:    s.npcsInRoomLocked(player.Room),
	})
}
