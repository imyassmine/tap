package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Obtainable  bool   `json:"obtainable"`
}

type NPC struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Dialogue    []string `json:"dialogue"`
	HP          int      `json:"hp"`
	Hostile     bool     `json:"hostile"`
}

type Room struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Items       []string          `json:"items"`
	NPCs        []string          `json:"npcs"`
}

type World struct {
	Start string          `json:"start"`
	Rooms map[string]Room `json:"rooms"`
	Items map[string]Item `json:"items"`
	NPCs  map[string]NPC  `json:"npcs"`
}

func Load(path string) (*World, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var w World
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&w); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := w.Validate(); err != nil {
		return nil, fmt.Errorf("invalid world %s: %w", path, err)
	}
	return &w, nil
}

// Validate vérifie la cohérence du monde et renvoie TOUTES les erreurs d'un coup.
func (w *World) Validate() error {
	var errs []error
	bad := func(format string, a ...any) {
		errs = append(errs, fmt.Errorf(format, a...))
	}
	checkID := func(kind, id, prefix string) {
		if !strings.HasPrefix(id, prefix) || strings.ContainsAny(id, " \t") {
			bad("%s %q: id must start with %q and contain no space", kind, id, prefix)
		}
	}

	if _, ok := w.Rooms[w.Start]; !ok {
		bad("start room %q does not exist", w.Start)
	}

	itemIn := map[string]string{} // ID d'item -> salle où on l'a déjà vu
	npcIn := map[string]string{}  // ID de NPC -> salle où on l'a déjà vu

	for id, r := range w.Rooms {
		checkID("room", id, "loc.")
		if r.Name == "" {
			bad("room %q: empty name", id)
		}
		for dir, target := range r.Exits {
			if dir == "" || dir != strings.ToLower(dir) || strings.ContainsAny(dir, " \t") {
				bad("room %q: invalid direction %q (lowercase, no space)", id, dir)
			}
			if _, ok := w.Rooms[target]; !ok {
				bad("room %q: exit %q leads to unknown room %q", id, dir, target)
			}
		}
		for _, it := range r.Items {
			if _, ok := w.Items[it]; !ok {
				bad("room %q: unknown item %q", id, it)
			} else if prev, dup := itemIn[it]; dup {
				bad("item %q is in both %q and %q (items must be unique)", it, prev, id)
			} else {
				itemIn[it] = id
			}
		}
		for _, n := range r.NPCs {
			if _, ok := w.NPCs[n]; !ok {
				bad("room %q: unknown npc %q", id, n)
			} else if prev, dup := npcIn[n]; dup {
				bad("npc %q is in both %q and %q", n, prev, id)
			} else {
				npcIn[n] = id
			}
		}
	}

	for id, it := range w.Items {
		checkID("item", id, "item.")
		if it.Name == "" {
			bad("item %q: empty name", id)
		}
	}
	for id, n := range w.NPCs {
		checkID("npc", id, "npc.")
		if n.Name == "" {
			bad("npc %q: empty name", id)
		}
		if n.HP < 0 || (n.Hostile && n.HP == 0) {
			bad("npc %q: invalid hp %d", id, n.HP)
		}
	}
	return errors.Join(errs...)
}
