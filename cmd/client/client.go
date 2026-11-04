package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:4242")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("\n[Disconnected from server]")
				os.Exit(0)
			}
			fmt.Printf("\n%s\n> ", strings.TrimSpace(line))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		text := scanner.Text()
		_, err := fmt.Fprintf(conn, "%s\n", text)
		if err != nil {
			fmt.Println("Send error:", err)
			break
		}

		if strings.EqualFold(strings.TrimSpace(text), "QUIT") {
			break
		}
	}
}
