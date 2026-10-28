package main

import (
	"encoding/json"
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

func (s *Server) cmdLook(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	data, err := s.lookJSON(p)
	if err != nil {
		slog.Error("look failed", "player", p.Name, "error", err)
		return "ERR 500 INTERNAL_ERROR"
	}
	return "OK " + string(data)
}

func (s *Server) lookJSON(p *Player) ([]byte, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	room := s.world.Rooms[p.Room]
	players := []string{}
	for _, other := range s.players {
		if other.Room == p.Room {
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
			ID:          p.Room,
			Name:        room.Name,
			Description: room.Description,
			Exits:       exits,
		},
		Players: players,
		Items:   append([]string{}, s.roomItems[p.Room]...),
		NPCs:    s.npcsInRoomLocked(p.Room),
	})
}
