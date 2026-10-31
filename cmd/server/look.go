package main

import (
	"encoding/json"
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
	data, _ := json.Marshal(lookReply{
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
	return "OK " + string(data)
}

func (s *Server) npcsInRoomLocked(roomID string) []string {
	npcs := []string{}
	for _, id := range s.world.Rooms[roomID].NPCs {
		if st, ok := s.npcs[id]; ok && st.Alive {
			npcs = append(npcs, id)
		}
	}
	return npcs
}
