package world

import (
	"encoding/json"
	"fmt"
	"os"
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
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&w); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &w, nil
}
