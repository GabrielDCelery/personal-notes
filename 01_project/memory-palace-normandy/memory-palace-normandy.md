# Memory Palace — Normandy SR-2

Map all 12 system design docs in `02_area/system-design/` to rooms on the Normandy SR-2.
Each room = one doc. The palace is walked top deck to bottom.

---

## Route

Cockpit → Galaxy Map → Miranda's Office → Mess Hall → Mordin's Lab → Med Bay → AI Core → Armoury → Garrus's Alcove → Cargo Bay → Engineering → Engine Room

---

## Room Status

| Room                | Doc                              | File                      | Status                                      |
| ------------------- | -------------------------------- | ------------------------- | ------------------------------------------- |
| Cockpit (Joker)     | 02 - Anatomy of a Request        | cockpit-joker.md          | In progress — HUD + Joker anchors written, comms + EDI still to do |
| Galaxy Map (CIC)    | 01 - Numbers and Quick Math      | galaxy-map-cic.md         | Complete — anchors + image prompts, no imgs |
| Miranda's Office    | 06 - Scaling Decisions           | —                         | Not started                                 |
| Mess Hall           | 04 - Caching                     | —                         | Not started                                 |
| Mordin's Lab        | 03 - Databases                   | —                         | Not started                                 |
| Med Bay (Chakwas)   | 09 - Logging and Observability   | —                         | Not started                                 |
| AI Core (EDI)       | 10 - Auth Patterns               | —                         | Not started                                 |
| Armoury (Jacob)     | 12 - Security                    | —                         | Not started                                 |
| Garrus's Alcove     | 11 - Communication Protocols     | —                         | Anchors designed (see below), no file yet   |
| Cargo Bay           | 05 - Queues and Async            | —                         | Not started                                 |
| Engineering (Tali)  | 07 - Large Data and Migrations   | engineering-tali.md       | Placeholder — content listed, no anchors    |
| Engine Room         | 08 - Cost and Storage Lifecycle  | —                         | Not started                                 |

---

## Garrus's Alcove — Anchor Design (Communication Protocols)

**Room layout:**

- Entrance
- Centre table/panel (right in front as you enter)
- Thanix cannon chamber (at the back beyond the table)
- Left wall: wall computer + cables running floor to ceiling + monitors on the walls
- Small room — everything visible at once

### Centre table — the four protocols + decision framework

- Pistol at the front = REST (default, pick this up first)
- Pistol duct-taped to a ticking metronome firing at the wall = short polling (wasted ammo counter on monitor above)
- Pistol aimed at the cannon chamber door, trigger held, Garrus frozen waiting = long polling
- Decision flowchart drawn on the table surface: "who initiates? how often?"
- Garrus standing at the table, glaring at the metronome pistol (emotional hook)

### Thanix cannon chamber — SSE + WebSockets

- Cannon firing one way only = SSE (unidirectional, persistent, server to client)
- Two cannons facing each other both firing = WebSockets (bidirectional, full duplex)

### Cannon chamber door frame — latency ranking

- Etched from top to bottom, closest to furthest:
  - WebSockets ~1–5 ms
  - SSE ~1–10 ms
  - Long polling ~10–100 ms
  - Short polling up to 5,000 ms

### Wall monitors — latency numbers displayed as bar chart

- Same four protocols, fastest to slowest, visual bar style

### Left wall cables + wall computer — scaling

- Cables floor to ceiling = the stateful problem (Client A plugged into Node 1, event fires from Node 2, sparks where they don't connect)
- Three junction boxes on wall computer = three solutions:
  1. Cable bolted to Node 1 (sticky sessions — works but unbalanced)
  2. Relay hub all nodes connect to (pub/sub backplane)
  3. Node 3 handles only cables, others handle only logic (dedicated gateway)
- Small plaque at the bottom: "10K–100K connections. Memory not CPU."

---

## Still To Do

- Design anchors for Cockpit (Joker) — next up
- Create garrus-alcove.md and write image prompts
- Design anchors for Engineering (Tali)
- Work through remaining 8 rooms
- Generate AI images for each room once prompts are ready
