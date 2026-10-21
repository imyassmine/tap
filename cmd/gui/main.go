package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
)

var tcpConn net.Conn
var serverLogs = "Not connected to MUD server yet.\n"
var chatMessages = "--- Chat Room ---\n"
var roomInfo = "Room info will appear here after LOOK.\n"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="utf-8">
			<title>MUD Game GUI - 42 Project</title>
			<style>
				body { font-family: monospace; background: #1e1e1e; color: #d4d4d4; padding: 20px; }
				.box { background: #252526; border: 1px solid #333; padding: 10px; margin-bottom: 10px; height: 150px; overflow-y: scroll; white-space: pre-wrap; }
				input, button { padding: 8px; margin: 4px; font-family: monospace; }
				input[type="text"] { background: #3c3c3c; color: #fff; border: 1px solid #555; width: 250px; }
				button { background: #007acc; color: white; border: none; cursor: pointer; }
				button:hover { background: #005999; }
				.panel { margin-bottom: 15px; }
			</style>
		</head>
		<body>
			<h2>MUD Web Client (Project 42)</h2>
			
			<div class="panel">
				<form action="/connect" method="POST">
					IP: <input type="text" name="ip" value="127.0.0.1">
					Port: <input type="text" name="port" value="4242">
					<button type="submit">Connect</button>
				</form>
			</div>

			<div class="panel">
				<strong>Quick Actions:</strong><br>
				<button onclick="sendCmd('LOOK')">LOOK</button>
				<button onclick="sendCmd('INVENTORY')">INVENTORY</button>
				<button onclick="sendCmd('WHO')">WHO</button>
				<button onclick="sendCmd('STATUS')">STATUS</button>
				<button onclick="sendCmd('QUESTS')">QUESTS</button>
				<button onclick="sendCmd('GROUP')">GROUP</button>
				<button onclick="sendCmd('QUIT')">QUIT</button>
				<br>
				<button onclick="sendCmd('MOVE north')">North</button>
				<button onclick="sendCmd('MOVE south')">South</button>
				<button onclick="sendCmd('MOVE east')">East</button>
				<button onclick="sendCmd('MOVE west')">West</button>
			</div>

			<div class="panel">
				Item: <input type="text" id="itemName" placeholder="item name">
				<button onclick="sendCustom('TAKE')">TAKE</button>
				<button onclick="sendCustom('DROP')">DROP</button>
				<br>
				NPC: <input type="text" id="npcName" placeholder="npc name">
				<button onclick="sendCustom('TALK')">TALK</button>
				<button onclick="sendCustom('ATTACK')">ATTACK</button>
			</div>

			<div class="panel">
				<form action="/send" method="POST">
					Custom Command: <input type="text" name="cmd" style="width: 400px;" placeholder="e.g. CHAT GLOBAL hello">
					<button type="submit">Send</button>
				</form>
			</div>

			<h3>Room & Environment:</h3>
			<div class="box" id="roomBox">` + roomInfo + `</div>

			<h3>Chat Room:</h3>
			<div class="box" id="chatBox">` + chatMessages + `</div>

			<h3>Server Logs & Output:</h3>
			<div class="box" id="logBox">` + serverLogs + `</div>

			<script>
				function sendCmd(c) {
					fetch('/cmd?action=' + encodeURIComponent(c)).then(() => location.reload());
				}
				function sendCustom(actionType) {
					let val = "";
					if (actionType === 'TAKE' || actionType === 'DROP') {
						val = document.getElementById('itemName').value;
					} else {
						val = document.getElementById('npcName').value;
					}
					fetch('/cmd?action=' + encodeURIComponent(actionType + (val ? " " + val : ""))).then(() => location.reload());
				}
				// Auto scroll boxes to bottom
				window.onload = function() {
					['roomBox', 'chatBox', 'logBox'].forEach(id => {
						let el = document.getElementById(id);
						el.scrollTop = el.scrollHeight;
					});
				}
			</script>
		</body>
		</html>
		`
		w.Write([]byte(html))
	})

	// هندلر اتصال به سرور TCP
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		ip := r.FormValue("ip")
		port := r.FormValue("port")
		address := ip + ":" + port

		conn, err := net.Dial("tcp", address)
		if err != nil {
			serverLogs += "\n[Error]: Connection failed: " + err.Error()
		} else {
			tcpConn = conn
			serverLogs += "\n[Success]: Connected to MUD server at " + address
			
			go func() {
				reader := bufio.NewReader(tcpConn)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						serverLogs += "\n[Disconnected]: Lost connection to server."
						break
					}
					line = strings.TrimSpace(line)
					if line != "" {
						if strings.Contains(line, "CHAT") {
							chatMessages += "\n" + line
						} else {
							serverLogs += "\n<< " + line
							if strings.Contains(line, "room") || strings.Contains(line, "OK") {
								roomInfo = line
							}
						}
					}
				}
			}()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		cmd := r.FormValue("cmd")
		executeCommand(cmd)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/cmd", func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("action")
		executeCommand(cmd)
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Web GUI started! Open your browser and go to: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func executeCommand(cmd string) {
	if tcpConn == nil {
		serverLogs += "\n[Error]: Not connected to server yet!"
		return
	}
	if cmd != "" {
		_, err := fmt.Fprintf(tcpConn, "%s\n", cmd)
		if err != nil {
			serverLogs += "\n[Error sending]: " + err.Error()
		} else {
			serverLogs += "\n>> " + cmd
		}
	}
}