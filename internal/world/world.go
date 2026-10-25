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
	decoder.DisallowUnknownFields()
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
	check := func(kind, id, prefix, name string) {
		if !strings.HasPrefix(id, prefix) || strings.ContainsAny(id, " \t") {
			bad("%s %q: id must start with %q and contain no space", kind, id, prefix)
		}
		if name == "" {
			bad("%s %q: empty name", kind, id)
		}
	}
	placed := map[string]string{}
	place := func(kind, roomID, id string, exists bool) {
		if !exists {
			bad("room %q: unknown %s %q", roomID, kind, id)
		} else if prev, dup := placed[id]; dup {
			bad("%s %q is in both %q and %q (must be unique)", kind, id, prev, roomID)
		} else {
			placed[id] = roomID
		}
	}

	if _, ok := w.Rooms[w.Start]; !ok {
		bad("start room %q does not exist", w.Start)
	}
	for id, r := range w.Rooms {
		check("room", id, "loc.", r.Name)
		for dir, target := range r.Exits {
			if dir == "" || dir != strings.ToLower(dir) || strings.ContainsAny(dir, " \t") {
				bad("room %q: invalid direction %q (lowercase, no space)", id, dir)
			}
			if _, ok := w.Rooms[target]; !ok {
				bad("room %q: exit %q leads to unknown room %q", id, dir, target)
			}
		}
		for _, it := range r.Items {
			_, ok := w.Items[it]
			place("item", id, it, ok)
		}
		for _, n := range r.NPCs {
			_, ok := w.NPCs[n]
			place("npc", id, n, ok)
		}
	}
	for id, it := range w.Items {
		check("item", id, "item.", it.Name)
	}
	for id, n := range w.NPCs {
		check("npc", id, "npc.", n.Name)
		if n.HP < 0 || (n.Hostile && n.HP == 0) {
			bad("npc %q: invalid hp %d", id, n.HP)
		}
	}
	return errors.Join(errs...)
}
