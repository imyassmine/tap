package main

import "strconv"

func (s *Server) cmdWho(args string) string {
	if args != "" {
		return "ERR 400 BAD_REQUEST"
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return "OK players=" + strconv.Itoa(len(s.players))
}
