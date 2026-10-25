package main

type NPCState struct {
	HP    int
	MaxHP int
	Alive bool
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
