package main

import (
	"encoding/json"
	"log/slog"
)

const npcAttackDamage = 10

type attackReply struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}

func (s *Server) cmdAttack(p *Player, args string) string {
	query := normalize(args)
	if query == "" {
		return "ERR 400 BAD_REQUEST"
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, id := range s.world.Rooms[p.Room].NPCs {
		state, ok := s.npcs[id]
		if !ok || !state.Alive || !matchRef(query, id, s.world.NPCs[id].Name) {
			continue
		}
		if !s.world.NPCs[id].Hostile {
			return "ERR 405 NPC_NOT_HOSTILE"
		}

		damage := npcAttackDamage
		if damage > state.HP {
			damage = state.HP
		}
		state.HP -= damage
		status := "combat"
		if state.HP <= 0 {
			state.HP = 0
			state.Alive = false
			status = "victory"
		}

		data, err := json.Marshal(attackReply{
			AttackerHP: p.HP,
			TargetHP:   state.HP,
			Damage:     damage,
			Status:     status,
		})
		if err != nil {
			slog.Error("attack response marshal failed", "player", p.Name, "npc", id, "error", err)
			return "ERR 500 INTERNAL_ERROR"
		}
		slog.Info("npc attacked", "player", p.Name, "npc", id, "damage", damage, "target_hp", state.HP)
		return "OK " + string(data)
	}
	return "ERR 404 NPC_NOT_FOUND"
}
