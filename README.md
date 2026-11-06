*This activity has been created as part of the 42 curriculum by hrezvani, yassabda.*

# The Answer Protocol (TAP) - A Shared-World Retro Text Adventure

## Description

**The Answer Protocol (TAP)** is a small multiplayer text adventure inspired by classic MUDs (Multi-User Dungeons). A TCP server hosts a shared virtual world; several players connect at the same time to explore rooms, chat, pick up and drop items, talk to NPCs, fight a hostile NPC and form groups.

The goal of the activity is to implement the line-based, UTF-8 protocol described in **RFC 42TAP** and to provide two interchangeable clients:

1. a **CLI client** (terminal),
2. a **GUI client** (web interface served locally by a small Go program).

Everything is written in **Go**. World data (rooms, items, NPCs, quests) is static JSON in `data/world.json`; there is no persistence, the state resets when the server restarts.

## Instructions

```bash
make install          # download dependencies
make run-server       # terminal 1: start the server (port 4242)
make run-client       # terminal 2: CLI client
make run-client-gui   # terminal 3: GUI client, then open http://localhost:8080
```

Run every command **from the repository root**: the server loads `data/world.json` with a relative path. See [Building and Running](#building-and-running) for all targets and prerequisites.

## Architecture

- **Dispatcher.** `handleConn` (`cmd/server/main.go`) reads one line, splits it into a command and its arguments, upper-cases the command (commands are case-insensitive) and dispatches it with a `switch` to one handler per command (`cmdLook`, `cmdMove`, `cmdTake`, ...). Each handler lives in its own file and returns the reply line. We chose a dispatcher over inline handling so that a command can be added or read in isolation.
- **Concurrency model.** The listener accepts connections in a loop and starts one goroutine per client (`handleConn`). Each client also has a writer goroutine (`writeLoop`) fed by a buffered channel (`chan string`, 64 messages).
- **Replies and events share one queue.** Replies to commands and asynchronous `EVT` messages are put in the same per-player channel, so a client always receives its messages in the order they were produced and a slow client never blocks the server: if its queue is full the message is dropped and a WARN is logged.
- **Shared state.** The world state (players, room items, NPC state, groups) is protected by a single `sync.Mutex`. Every handler takes the lock for its whole duration, so each command is atomic and the code is easy to reason about. Helper functions that need the lock already held are suffixed `Locked`.
- **Disconnection.** On `QUIT` or a dropped connection the player is removed from the player map first, then the inventory is dropped in the current room, the player leaves their group, and `EVT ROOM PRESENCE LEAVE` / `EVT STATS` are broadcast.
- **CLI client** (`cmd/client`). It sends what the user types **as-is**, using the RFC syntax (for example `MOVE north`, `CHAT GLOBAL hello`). A goroutine prints incoming lines (replies and events) while the main loop waits for user input, so events are received at any time. It connects to `127.0.0.1:4242`. We chose raw RFC syntax because it keeps the client interchangeable with other groups' servers and makes the protocol visible while testing.
- **GUI client** (`cmd/gui`). A small Go HTTP server (port 8080) that opens the TCP connection to the game server (address and port are entered in the page) and relays commands from buttons to the server. The page has quick-action buttons, item/NPC fields, a free command box, and separate boxes for room information, chat and server log.

## Protocol Implementation

The server implements every command and event of RFC 42TAP. Where the RFC or the subject leaves room, or where we differ, it is listed here.

**Framing and robustness**
- One message per line (`\n`), UTF-8. TCP fragmentation and coalescing are handled by `bufio.Scanner`.
- Maximum line length: 1024 bytes. A longer line gets `ERR 400 BAD_REQUEST`, the event is logged and the connection is closed.
- Control characters are removed from chat and dialogue text.
- Commands are case-insensitive. Arguments such as `MOVE` directions and `CHAT` scopes are case-sensitive (`north`, `GLOBAL`); item and NPC references (ID or display name) are case-insensitive and may contain spaces.

**Additions to the RFC**
- Error codes not listed in the RFC table: `202 NOT_AUTHENTICATED` (command before `CONNECT`), `400 BAD_REQUEST` (unknown command, missing/extra argument, malformed input), `404 QUEST_NOT_FOUND`, `500 INTERNAL_ERROR`.
- Additional commands (allowed by RFC section 6.1.1): `DEFEND` and `FLEE`.
- Additional event: `EVT ROOM COMBAT <player> <npc-id> <damage> <npc-hp> <status>`, sent to the other players of the room where a fight happens (see [Combat System](#combat-system)).
- `EVT ROOM PRESENCE LEAVE/ENTER` is also sent when a player flees or respawns after being defeated.

**Choices where the RFC and the subject differ or are open**
- `TALK` replies `OK <dialogue>` and `WHO` replies `OK players=<count>` as defined by the RFC, not the JSON shown in the subject's examples.
- `GROUP JOIN <leader>` does not require a pending invitation: `GROUP INVITE` is only a notification. A player can join any existing group by naming a member.
- `STATUS` replies `OK {"hp": <n>, "max_hp": <n>, "status": "<healthy|wounded|defending>"}`.
- `QUEST <quest-id>` takes a quest identifier (not an NPC name) and `QUESTS` returns the identifiers of all quests of the world, not the player's own quests. See [Quest System](#quest-system).

## Combat System

Players start with **100 HP** (`playerMaxHP`). Only NPCs flagged `hostile` in `world.json` can be attacked (`ERR 405 NPC_NOT_HOSTILE` otherwise).

| Command | Effect |
|---|---|
| `ATTACK <npc>` | Deals **15** damage to the NPC (never more than its remaining HP). If the NPC survives, it immediately counter-attacks for **10** damage. |
| `DEFEND` | The next counter-attack is halved (**5** damage). The effect stays until a counter-attack consumes it or the player flees. |
| `FLEE` | Leaves combat at once and teleports the player to the start room (`loc.square`). Clears the defending state. |
| `STATUS` | Shows `hp`, `max_hp` and `status` (`healthy`, `wounded` when HP < max, `defending`). |

**Turn order (initiative).** One `ATTACK` is one round: the player always acts first, then the NPC answers with a counter-attack if it is still alive. NPCs never attack on their own. A killed NPC (`status: "victory"`) deals no counter-damage and is not respawned.

**Defeat.** When the player's HP reaches 0 the attack reply has `status: "defeated"`; the player respawns in the start room with **50 HP**.

**Reply and broadcast.** The attacker receives `OK {"attacker_hp":..,"target_hp":..,"damage":..,"status":"combat|victory|defeated"}`. The other players in the room receive `EVT ROOM COMBAT <player> <npc-id> <damage> <npc-hp> <status>`, followed by presence events if the attacker was defeated and moved to the start room. Every attack is also logged (INFO, WARN on player death).

**Why these choices.** Fixed damage and an immediate counter-attack make combat deterministic, easy to test and fully resolved inside one command, which fits the single-mutex design (no timers or tick loop, no combat state shared between commands except the defending flag). `DEFEND` and `FLEE` give the player a real decision without needing a full turn queue.

**Limits.** A killed NPC stays dead until the server restarts.

## Quest System

- Quests are static data in `data/world.json`: `title`, `description`, `target` (an item or NPC identifier) and `reward` (an item identifier). `Validate()` checks at startup that the target and the reward exist.
- `QUESTS` returns the JSON list of quest identifiers defined in the world.
- `QUEST <quest-id>` returns the quest definition as JSON (`title`, `description`, `target`, `reward`), or `ERR 404 QUEST_NOT_FOUND`.
- Current quests: `quest.baker` (*The Baker's Key*: bring the rusty key from the cellar) and `quest.herbs` (*Gathering Herbs*: collect healing herbs from the dark alley). The Master Baker's dialogue hints at the first one.

## World Design

The world has **8 rooms** (`data/world.json`), starting in the Village Square:

```
                 Quiet Garden
                      |
 Old Wishing Well --- Village Square --- General Store
                      |        ^               |
                 The Prancing  |          Shop Storage
                   Pony        +--- Dark Alley ---+
                      |
                 Dark Cellar
```

- **Loop:** Village Square -> General Store -> Shop Storage -> Dark Alley -> Village Square (the way back from the alley is one-way; the alley also connects west to the storage).
- **Branches:** Tavern -> Dark Cellar, and Well -> Quiet Garden (both dead ends).

| NPC | Room | HP | Role |
|---|---|---|---|
| Village Guard | Village Square | 20 | Dialogue / flavor |
| Master Baker | The Prancing Pony | 25 | Dialogue, hints at the Baker's Key quest |
| Shop Keeper | General Store | 15 | Dialogue / merchant theme |
| Goblin | Dark Alley | 30 | Hostile enemy |

| Item | Room | Obtainable |
|---|---|---|
| Frothy Ale | The Prancing Pony | yes |
| Loaf of Bread | General Store | yes |
| Rusty Key | Dark Cellar | yes |
| Healing Herbs | Dark Alley | yes |

**Item management.** Each item is a unique instance: `TAKE` removes it from the room and puts it in the player's inventory, `DROP` puts it back in the room (visible to others), and items of a disconnecting player are dropped in their room. Items can be referenced by ID (`item.herbs`) or display name (`Healing Herbs`, multi-word names supported).

**Validation.** At startup the world is loaded with unknown JSON fields rejected and `Validate()` checks the start room, every exit target, every item and NPC referenced by a room, NPC dialogue, and quest targets and rewards. If anything is wrong the server logs an error and exits with status 1.

## Server Logging

Logging uses Go's `log/slog` with the **JSON handler** on **stderr**. Every record has a precise `time`, a `level` (INFO, WARN, ERROR), a `msg` and structured attributes.

| Category | Examples of `msg` | Level |
|---|---|---|
| Connections | `client connected`, `client disconnected` (with remote `ip:port`), `player connected` | INFO |
| Commands | `command received` (remote, player, command, args) | INFO |
| Responses | `response sent` (every reply and event, including `ERR` codes) | INFO |
| World changes | `player moved`, `item taken`, `item dropped`, `npc talk`, `combat attack`, `player fled from combat`, `group created`, `user joined`, `inventory dropped on disconnect` | INFO |
| Combat deaths | `player died and respawned` | WARN |
| Errors and anomalies | `empty line`, `read error`, `message dropped`, `response channel full`, `send failed` | WARN |
| Failures | `failed to load world`, `failed to start server`, `accept error`, `... marshal failed` | ERROR |

**Abuse detection** (monitor and log only, nothing is blocked):

- **Command flooding:** more than **20 lines per second** on one connection (empty lines included) logs `command flooding detected` (WARN) once per window, with remote address and player.
- **Rapid connections:** more than **10 connections within 10 seconds** from the same IP logs `rapid connections detected` (WARN) with the IP and the count.
- **Oversized lines:** a line over 1024 bytes logs `read error` (`bufio.ErrTooLong`), the client gets `ERR 400 BAD_REQUEST` and is disconnected.

**Monitoring.**

```bash
make run-server 2> server.log      # keep the logs in a file
grep '"level":"WARN"' server.log   # all warnings
grep 'flooding' server.log         # abuse patterns
grep '"player":"alice"' server.log # activity of one player
```

Example record:

```json
{"time":"2026-11-06T10:12:03.41+01:00","level":"INFO","msg":"combat attack","player":"alice","npc":"npc.goblin","damage":15,"target_hp":15,"attacker_hp":90}
```

## Group Contributions

- **yassabda:** server (protocol, connection handling, commands, logging) and world validation.
- **hrezvani:** combat and quest system, world data (`data/world.json`) and CLI and GUI client.

## Building and Running

**Prerequisites:** Go 1.26.8 or newer (as declared in `go.mod`), `make`, and a C compiler (the Makefile runs Go with `-race`, which needs cgo). `golangci-lint` is optional, only for `lint-strict`.

| Target | Description |
|---|---|
| `make install` | Download Go dependencies |
| `make run-server` | Run the server on port `4242` |
| `make run-client` | Run the CLI client (connects to `127.0.0.1:4242`) |
| `make run-client-gui` | Run the GUI client, then open `http://localhost:8080` and connect to the server |
| `make fmt` | Format all Go files with `gofmt` |
| `make lint` | Check `gofmt` formatting and run `go vet` |
| `make lint-strict` | `lint`, then `golangci-lint` |
| `make test` | Run `go test -race -v ./...` |
| `make clean` | Clear the test cache |
| `make help` | List the targets |

## Testing

Tests are manual: start the server, then use two or more clients (`make run-client`, or `nc 127.0.0.1 4242`). With the CLI, type the commands exactly as in the RFC.

**Multiplayer.** Open two clients A and B.
1. A: `CONNECT alice`, B: `CONNECT bob`. A receives `EVT ROOM PRESENCE ENTER bob` and both receive `EVT STATS players=2`. A third client using `CONNECT alice` gets `ERR 201 NAME_IN_USE`.
2. A: `CHAT GLOBAL hi` -> A gets `OK` and `EVT GLOBAL CHAT alice hi`; B gets the `EVT`. Try `CHAT ROOM ...` with the clients in different rooms.
3. A: `MOVE north` -> `OK room=loc.tavern`, B receives `EVT ROOM PRESENCE LEAVE alice`. A: `TAKE Frothy Ale`, then B moves north and `LOOK`: the ale is no longer there. A: `DROP item.ale`: B sees it again.
4. A: `GROUP CREATE`, A: `GROUP INVITE bob` (B receives `EVT GROUP INVITE alice`), B: `GROUP JOIN alice`, then `CHAT GROUP ...`, `GROUP LEAVE`.
5. A: `QUIT` -> B receives the leave event and the new player count.

**Combat.** Walk to the goblin: `MOVE east`, `MOVE north`, `MOVE east` (Dark Alley).
- `ATTACK goblin` -> `{"attacker_hp":90,"target_hp":15,"damage":15,"status":"combat"}`; a second attack gives `target_hp` 0 and `"victory"`. With `DEFEND` before the first attack the counter-attack costs 5 HP (`attacker_hp` 95). `STATUS` shows the player's state.
- A second player in the alley receives `EVT ROOM COMBAT alice npc.goblin 15 15 combat`.
- `FLEE` returns to `loc.square`.
- The respawn path cannot be reached with the default goblin (it dies in two hits). To test it, temporarily raise the goblin's `hp` to 500 in `data/world.json`: the tenth `ATTACK` returns `"status":"defeated"` and the player is back in the square with 50 HP.
- `ATTACK guard` gives `ERR 405 NPC_NOT_HOSTILE`.

**Quests.** `QUESTS` lists the quest identifiers, `QUEST quest.baker` returns its definition, `QUEST unknown` gives `ERR 404 QUEST_NOT_FOUND`.

**Protocol errors.** Sending a command before `CONNECT` gives `ERR 202 NOT_AUTHENTICATED`; `FOO` gives `ERR 400 BAD_REQUEST`; `MOVE up` gives `ERR 301 NO_EXIT`; `TAKE nothing` gives `ERR 404 ITEM_NOT_FOUND`.

**Abuse detection.** Each check should produce a WARN in the server logs:
```bash
# command flooding
(echo "CONNECT flooder"; yes WHO | head -100) | nc 127.0.0.1 4242
# rapid connections
for i in $(seq 12); do echo QUIT | nc 127.0.0.1 4242 > /dev/null; done
# oversized line
head -c 2000 /dev/zero | tr '\0' 'a' | nc 127.0.0.1 4242
```

## Resources

**References**
- RFC 42TAP (project specification) and the subject.
- [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119) - requirement levels; [RFC 5234](https://www.rfc-editor.org/rfc/rfc5234) - ABNF; [RFC 793](https://www.rfc-editor.org/rfc/rfc793) - TCP; [RFC 3629](https://www.rfc-editor.org/rfc/rfc3629) - UTF-8.
- Go documentation: [`net`](https://pkg.go.dev/net), [`bufio`](https://pkg.go.dev/bufio), [`log/slog`](https://pkg.go.dev/log/slog), [`sync`](https://pkg.go.dev/sync), [A Tour of Go](https://go.dev/tour/), [Effective Go](https://go.dev/doc/effective_go), [Structured Logging with slog](https://go.dev/blog/slog).
- [Multi-user dungeon (Wikipedia)](https://en.wikipedia.org/wiki/MUD).

**Use of AI**
- AI was used to review the server code against the subject and RFC 42TAP and this README.
