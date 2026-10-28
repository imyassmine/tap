package main

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"tap/internal/world"
	"unicode"
	"unicode/utf8"
)

const (
	port         = ":4242"
	greeting     = "OK hello proto=1"
	maxLineSize  = 1024
	playerMaxHP  = 100
)

type Player struct {
	Name      string
	Room      string
	Inventory []string
	HP        int
	MaxHP     int
	out       chan string
}

type Server struct {
	mtx       sync.Mutex
	world     *world.World
	players   map[string]*Player
	roomItems map[string][]string
	npcs      map[string]*NPCState
}

func NewServer(w *world.World) *Server {
	s := &Server{
		world:     w,
		players:   make(map[string]*Player),
		roomItems: make(map[string][]string),
		npcs:      make(map[string]*NPCState),
	}
	for id, r := range w.Rooms {
		s.roomItems[id] = slices.Clone(r.Items)
	}
	for id, n := range w.NPCs {
		s.npcs[id] = &NPCState{HP: n.HP, MaxHP: n.HP, Alive: true}
	}
	return s
}

func main() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	w, err := world.Load("data/world.json")
	if err != nil {
		slog.Error("failed to load world", "error", err)
		os.Exit(1)
	}
	slog.Info("world loaded", "rooms", len(w.Rooms), "items", len(w.Items), "npcs", len(w.NPCs))

	ln, err := net.Listen("tcp", port)
	if err != nil {
		slog.Error("failed to start server", "port", port, "error", err)
		os.Exit(1)
	}
	defer ln.Close()
	slog.Info("server listening", "port", port)

	srv := NewServer(w)
	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("accept error", "error", err)
			continue
		}
		go srv.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	remote := conn.RemoteAddr().String()
	name := ""
	// p reste nil tant que CONNECT n'a pas réussi : toute autre commande est
	// alors refusée, donc les cmdXxx reçoivent toujours un joueur enregistré.
	var p *Player

	out := make(chan string)
	done := make(chan struct{})
	go writeLoop(conn, remote, out, done)
	reply := func(msg string) { out <- msg }

	slog.Info("client connected", "remote", remote)
	defer func() {
		if p != nil {
			s.unregister(p)
		}
		close(out)
		<-done
		_ = conn.Close()
		slog.Info("client disconnected", "remote", remote, "player", name)
	}()

	reply(greeting)

	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, maxLineSize)
	scanner.Buffer(buf, maxLineSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			slog.Warn("empty line", "remote", remote)
			continue
		}
		command, args, _ := strings.Cut(line, " ")
		command = strings.ToUpper(command)
		args = strings.TrimSpace(args)
		slog.Info("command received", "remote", remote, "player", name,
			"command", command, "args", args)

		if p == nil && command != "CONNECT" && command != "QUIT" {
			reply("ERR 202 NOT_AUTHENTICATED")
			continue
		}

		switch command {
		case "CONNECT":
			if p != nil || !validName(args) {
				reply("ERR 400 BAD_REQUEST")
			} else if p = s.register(args, out); p == nil {
				reply("ERR 201 NAME_IN_USE")
			} else {
				name = p.Name
				slog.Info("player connected", "remote", remote, "player", name)
				reply("OK connected")
				s.announceJoin(p)
			}
		case "QUIT":
			reply("OK bye")
			return
		case "LOOK":
			reply(s.cmdLook(p, args))
		case "MOVE":
			reply(s.cmdMove(p, args))
		case "CHAT":
			reply(s.cmdChat(p, args))
		case "WHO":
			reply(s.cmdWho(args))
		case "TAKE":
			reply(s.cmdTake(p, args))
		case "DROP":
			reply(s.cmdDrop(p, args))
		case "INVENTORY":
			reply(s.cmdInventory(p, args))
		case "QUESTS":
			reply(s.cmdQuests(p, args))
		case "QUEST":
			reply(s.cmdQuest(p, args))
		case "TALK":
			reply(s.cmdTalk(p, args))
		case "ATTACK":
			reply(s.cmdAttack(p, args))
		default:
			reply("ERR 400 BAD_REQUEST")
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("read error", "remote", remote, "error", err)
		if errors.Is(err, bufio.ErrTooLong) {
			reply("ERR 400 BAD_REQUEST")
		}
	}
}

func writeLoop(conn net.Conn, remote string, out <-chan string, done chan<- struct{}) {
	defer close(done)
	for msg := range out {
		if _, err := io.WriteString(conn, msg+"\n"); err != nil {
			slog.Warn("send failed", "remote", remote, "error", err)
			_ = conn.Close()
			for range out {
			}
			return
		}
		slog.Info("response sent", "remote", remote, "response", msg)
	}
}

func (s *Server) pushRoomLocked(roomID, except, msg string) {
	for _, p := range s.players {
		if p.Room == roomID && p.Name != except {
			p.push(msg)
		}
	}
}

func (s *Server) pushAllLocked(msg string) {
	for _, p := range s.players {
		p.push(msg)
	}
}

func (player *Player) push(msg string) {
	select {
	case player.out <- msg:
	default:
		slog.Warn("message dropped", "player", player.Name)
	}
}

func (s *Server) statsEventLocked() string {
	return "EVT STATS players=" + strconv.Itoa(len(s.players))
}

// register renvoie le joueur créé, ou nil si le nom est déjà pris.
func (s *Server) register(name string, out chan string) *Player {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, taken := s.players[name]; taken {
		return nil
	}
	p := &Player{
		Name:  name,
		Room:  s.world.Start,
		HP:    playerMaxHP,
		MaxHP: playerMaxHP,
		out:   out,
	}
	s.players[name] = p
	return p
}

func (s *Server) announceJoin(p *Player) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.pushRoomLocked(p.Room, p.Name, "EVT ROOM PRESENCE ENTER "+p.Name)
	s.pushAllLocked(s.statsEventLocked())
}

func (s *Server) unregister(p *Player) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.players, p.Name)
	s.pushRoomLocked(p.Room, "", "EVT ROOM PRESENCE LEAVE "+p.Name)
	s.pushAllLocked(s.statsEventLocked())
}

func validName(name string) bool {
	if name == "" || !utf8.ValidString(name) {
		return false
	}
	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}
