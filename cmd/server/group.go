package main

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

type Group struct {
	ID      string
	Members []string
}

func (s *Server) cmdGroup(p *Player, args string) string {
	sub, text, _ := strings.Cut(args, " ")
	sub = strings.ToUpper(sub)
	switch sub {
	case "CREATE":
		return s.cmdGroupCreate(p, text)
	case "INVITE":
		return s.cmdGroupInvite(p, text)
	case "JOIN":
		return s.cmdGroupJoin(p, text)
	case "LEAVE":
		return s.cmdGroupLeave(p, text)
	default:
		return "ERR 400 BAD_REQUEST"
	}
}

func (s *Server) cmdGroupCreate(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if p.Group != nil {
		return "ERR 402 ALREADY_IN_GROUP"
	}
	s.nextGroupID++
	id := strconv.Itoa(s.nextGroupID)
	p.Group = &Group{
		ID:      id,
		Members: []string{p.Name},
	}
	slog.Info("group created", "player", p.Name, "group", id)
	return "OK group=" + id
}

func (s *Server) cmdGroupInvite(p *Player, args string) string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	target, ok := s.players[args]
	if !ok {
		return "ERR 400 BAD_REQUEST"
	}
	if p.Group == nil {
		return "ERR 401 NOT_IN_GROUP"
	}
	if target.Group != nil {
		return "ERR 402 ALREADY_IN_GROUP"
	}
	target.push("EVT GROUP INVITE " + p.Name)
	slog.Info("user invited", "player", p.Name, "target", target.Name, "group", p.Group.ID)
	return "OK"
}

func (s *Server) pushGroupLocked(g *Group, msg string) {
	for _, m := range g.Members {
		s.players[m].push(msg)
	}
}

func (s *Server) cmdGroupJoin(p *Player, args string) string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	target, ok := s.players[args]
	if !ok || target.Group == nil {
		return "ERR 400 BAD_REQUEST"
	}
	if p.Group != nil {
		return "ERR 402 ALREADY_IN_GROUP"
	}
	g := target.Group
	s.pushGroupLocked(g, "EVT GROUP JOIN "+p.Name)
	g.Members = append(g.Members, p.Name)
	p.Group = g
	slog.Info("user joined", "player", p.Name, "group", g.ID)
	return "OK group=" + g.ID
}

func (s *Server) cmdGroupLeave(p *Player, args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	g := p.Group
	if g == nil {
		return "ERR 401 NOT_IN_GROUP"
	}
	g.Members = slices.DeleteFunc(g.Members, func(n string) bool {
		return n == p.Name
	})
	p.Group = nil
	s.pushGroupLocked(g, "EVT GROUP LEAVE "+p.Name)
	slog.Info("user left", "player", p.Name, "group", g.ID)
	return "OK"
}
