package world

import (
	"encoding/json"
	"errors"
	"os"
)

type World struct {
	Start  string           `json:"start"`
	Rooms  map[string]Room  `json:"rooms"`
	Items  map[string]Item  `json:"items"`
	NPCs   map[string]NPC   `json:"npcs"`
	Quests map[string]Quest `json:"quests"`
}

type Room struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Items       []string          `json:"items"`
	NPCs        []string          `json:"npcs"`
}

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

type Quest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Target      string `json:"target"`
	Reward      string `json:"reward"`
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
		return nil, err
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return &w, nil
}

func (world *World) Validate() error {
	var errs []error
	if _, ok := world.Rooms[world.Start]; !ok {
		errs = append(errs, errors.New("start room does not exist"))
	}
	for _, room := range world.Rooms {
		for _, target := range room.Exits {
			if _, ok := world.Rooms[target]; !ok {
				errs = append(errs, errors.New("exit doesn't lead to an existing room"))
			}
		}
		for _, itemID := range room.Items {
			if _, ok := world.Items[itemID]; !ok {
				errs = append(errs, errors.New("item does not exist"))
			}
		}
		for _, npcID := range room.NPCs {
			if _, ok := world.NPCs[npcID]; !ok {
				errs = append(errs, errors.New("NPC does not exist"))
			}
		}
	}
	return errors.Join(errs...)
}
