*This activity has been created as part of the 42 curriculum by hrezvani, yassabda.*

# The Answer Protocol (TAP) - A Shared-World Retro Text Adventure

## Description
**The Answer Protocol (TAP)** is a multi-player retro text adventure game engine and network server inspired by classic Multi-User Dungeons (MUDs). Built using **Go** (Golang), TAP enables multiple players to connect concurrently to a persistent virtual world where they can explore interconnected rooms, chat globally or locally, trade items, complete quests, engage in turn-based combat with hostile NPCs, and form groups.

The project implements a line-oriented, UTF-8 encoded TCP server adhering strictly to the **RFC 42TAP** specification. The server is accompanied by two distinct client implementations:
1. **CLI Client:** A lightweight, terminal-based asynchronous client for low-overhead text interaction.
2. **GUI Client (Web Interface):** A responsive, browser-based graphical user interface featuring real-time state rendering, dedicated event/chat logs, and interactive controls.

---

## Architecture
The TAP server is designed around a multi-threaded, asynchronous concurrency model leveraging Go's native goroutines and channel primitives:

- **Connection Listener & Worker Model:** The server listens on TCP port `:4242`. Each client connection spawns a dedicated read goroutine (`handleConn`) and a non-blocking write worker loop (`writeLoop`).
- **Buffered Outbound Queues:** Communication between the main thread and each client connection is isolated via buffered Go channels (`chan string`), preventing slow networks or blocked clients from impacting server processing or other players.
- **State Synchronization & Thread Safety:** World state, player locations, inventories, quest statuses, and combat states are protected by a global mutex (`sync.Mutex`), ensuring thread-safe access across concurrent command handlers and event dispatchers.
- **Graceful Disconnection & Cleanup:** Upon client disconnection (via `QUIT` or ungraceful dropped connection), the server cleanly unregisters the player, drops their inventory into their current location, removes them from active groups, and broadcasts departure events to all relevant peers.

---

## Protocol Implementation
The server strictly implements the **RFC 42TAP** protocol specification using line-delimited (`\n`) UTF-8 JSON payloads and structured text replies:

- **Command Parsing:** Commands are case-insensitive, sanitized against control characters, and parsed securely.
- **Event Dispatch System:** Asynchronous notifications (`EVT`) are categorized and broadcasted based on scope:
  - `EVT ROOM`: Broadcasted to players sharing the same room (movement, presence, room chat).
  - `EVT GLOBAL`: Server-wide events (global announcements, global chat).
  - `EVT GROUP`: Scope-restricted events sent only to group members.
  - `EVT STATS`: Global telemetry updates (e.g., active player counts).
- **Error Codes:** Complete implementation of standard error responses (`ERR 201 NAME_IN_USE`, `ERR 301 NO_EXIT`, `ERR 400 BAD_REQUEST`, `ERR 404 ITEM_NOT_FOUND`, `ERR 405 NPC_NOT_HOSTILE`, etc.).

---

## Combat System
TAP features a dynamic, turn-based combat system designed for tactical NPC encounters:

- **Health System:** Players begin with **100 Max HP**. Hostile NPCs possess variable HP defined by world configuration.
- **Player Actions:**
  - **`ATTACK <npc>`:** Deals fixed base damage (15 HP) to the target NPC. If the NPC survives, it executes an immediate counter-attack dealing 10 HP damage to the player.
  - **`DEFEND`:** Braces the player for the next incoming attack, reducing incoming NPC counter-attack damage by 50% (5 HP).
  - **`FLEE`:** Allows a player to escape active combat instantly, retreating to the world spawn location (`start`).
  - **`STATUS`:** Displays current player HP, max HP, and active combat modifiers.
- **Defeat & Respawn:** When a player's HP reaches 0, they are defeated and automatically respawn at the starting location (`loc.square`) with **50 HP**.

---

## Quest System
The quest system handles player objective tracking, validation, and reward distribution:

- **Structure:** Quests are defined in world configuration with target conditions and reward items (e.g., *The Baker's Key*, *Gathering Herbs*).
- **Commands:**
  - **`QUESTS`:** Returns a JSON list of all available world quest identifiers.
  - **`QUEST <id>`:** Returns detailed information, prerequisites, target items/NPCs, and completion rewards for a specific quest.
- **Validation:** Quests inspect player inventory and room interaction states before granting item rewards upon completion.

---

## World Design
The virtual world is loaded dynamically from static JSON (`data/world.json`) and validated on startup:

- **Topology:** Contains 8 interconnected rooms (*Village Square*, *The Prancing Pony Tavern*, *General Store*, *Dark Cellar*, *Shop Storage*, *Dark Alley*, *Old Wishing Well*, *Quiet Garden*) configured with loops and multiple non-linear branches.
- **NPC Roles:** Features 4 distinct NPC types across 3 primary roles:
  - **Dialogue / Flavor:** *Village Guard* (Protector)
  - **Quest Givers / Merchants:** *Master Baker*, *Shop Keeper*
  - **Hostile Enemies:** *Goblin* (Aggressive enemy in Dark Alley with 30 HP)
- **Item Distribution:** Features 4 unique items (*Frothy Ale*, *Loaf of Bread*, *Rusty Key*, *Healing Herbs*), all of which are obtainable and dropable in-world.

---

## Server Logging
Server monitoring and activity auditing are powered by Go's structured `log/slog` logging package outputting formatted JSON logs to `os.Stderr`:

- **Event Categories:**
  - **Connection Lifecycle:** Logged with remote IP, port, and timestamp on connect/disconnect.
  - **Command Auditing:** Every inbound client packet is logged with player identity and arguments.
  - **World Events:** Item movements, room transitions, quest lookups, and combat damage/deaths.
  - **Security & Abuse Monitoring:** Invalid input, empty flooding packets, line length exceedances (`bufio.ErrTooLong`), and socket read errors.

---

## Building and Running

### Prerequisites
- **Go** (version 1.22 or higher)
- **Make**

### Build & Run Commands

```bash
# Download dependencies
make install

# Start the TCP MUD Server
make run-server

# Run the text CLI Client (Terminal 2)
make run-client

# Run the Web GUI Client (Terminal 3)
make run-client-gui
