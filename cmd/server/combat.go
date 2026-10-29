package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

const (
	defaultPlayerAttackDamage = 15
	defaultNPCCounterDamage   = 10
	respawnHP                 = 50
)

type attackReply struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}

func (s *Server) cmdStatus(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return fmt.Sprintf("OK {\"hp\": %d, \"max_hp\": %d, \"defending\": %t}", p.HP, p.MaxHP, p.IsDefending)
}

func (s *Server) cmdDefend(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	p.IsDefending = true
	slog.Info("player defending", "player", p.Name)
	return "OK you brace yourself, preparing for incoming attacks."
}

func (s *Server) cmdAttack(p *Player, args string) string {
	query := normalize(args)
	if query == "" {
		return "ERR 400 BAD_REQUEST"
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	room := s.world.Rooms[p.Room]
	for _, id := range room.NPCs {
		state, ok := s.npcs[id]
		if !ok || !state.Alive || !matchRef(query, id, s.world.NPCs[id].Name) {
			continue
		}
		npcDef := s.world.NPCs[id]
		if !npcDef.Hostile {
			return "ERR 405 NPC_NOT_HOSTILE"
		}

		damageToNPC := defaultPlayerAttackDamage
		if damageToNPC > state.HP {
			damageToNPC = state.HP
		}
		state.HP -= damageToNPC
		status := "combat"
		if state.HP <= 0 {
			state.HP = 0
			state.Alive = false
			status = "victory"
		}

		counterDamage := 0
		if state.Alive {
			counterDamage = defaultNPCCounterDamage
			if p.IsDefending {
				counterDamage = defaultNPCCounterDamage / 2
				p.IsDefending = false
			}
			p.HP -= counterDamage
			if p.HP < 0 {
				p.HP = 0
			}
		}

		if p.HP == 0 {
			p.HP = respawnHP
			p.Room = s.world.Start
			slog.Warn("player died and respawned", "player", p.Name, "respawn_room", s.world.Start)
			status = "defeated"
		}

		data, err := json.Marshal(attackReply{
			AttackerHP: p.HP,
			TargetHP:   state.HP,
			Damage:     damageToNPC,
			Status:     status,
		})
		if err != nil {
			slog.Error("attack response marshal failed", "player", p.Name, "npc", id, "error", err)
			return "ERR 500 INTERNAL_ERROR"
		}

		slog.Info("combat attack", "player", p.Name, "npc", id, "damage", damageToNPC, "target_hp", state.HP, "attacker_hp", p.HP)
		return "OK " + string(data)
	}
	return "ERR 404 NPC_NOT_FOUND"
}

func (s *Server) cmdFlee(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	from := p.Room
	p.Room = s.world.Start
	p.IsDefending = false

	s.pushRoomLocked(from, p.Name, "EVT ROOM PRESENCE LEAVE "+p.Name)
	s.pushRoomLocked(s.world.Start, p.Name, "EVT ROOM PRESENCE ENTER "+p.Name)
	slog.Info("player fled from combat", "player", p.Name, "from", from, "to", s.world.Start)
	return fmt.Sprintf("OK room=%s", s.world.Start)
}
