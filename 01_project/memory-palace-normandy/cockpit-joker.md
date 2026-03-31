# Cockpit — Joker (Anatomy of a Request)

The Cockpit is where Joker flies the ship — obsessed with speed, routing, latency. Every millisecond of every route is accounted for. He complains loudly about unnecessary trips.

Maps to: **02-anatomy-of-a-request.md**

## Theme

Distance = latency. Every operation in a request is a journey — within your system, leaving your system, or a one-time trip you should never have to repeat. Joker knows the cost of every leg.

## Room Layout

```
[============== VIEWPORT (space) ==============]
[EDI sphere]  [Joker's seat]  [Comms panel]
                  [ENTER]
```

EDI's holographic sphere is to Joker's left. Comms panel is to his right. You enter from behind. The HUD is overhead above Joker's seat.

## Route Through the Room

1. **HUD** (overhead) — classify the cost of any operation
2. **Joker talking** — plots the full request journey as a route
3. **Comms panel** — auth costs (still to design)
4. **EDI sphere** — parallel vs sequential (still to design)

---

## The Geography

Everything in a request is a journey. Distance maps directly to latency:

| Cost | Journey | Represents |
| ---- | ------- | ---------- |
| 1 ms | Inter-planet hop — sub-light burn to a nearby station | Intra-DC service call, load balancer |
| 5 ms | Planetary landing — descent, retrieve, ascend | DB query with index (Omega, coordinates in hand) |
| 50 ms | Flying to the relay — leaving your system boundary | External API call per request |
| 20–100 ms | Citadel visit — first time only | Cold DNS lookup (walked through once, never again) |
| 500 ms | Thrusters only — FTL drive is down | Broken data access: no index, N+1, no pool |

**Within your system** (same datacenter) — you never need a relay. Station hops and planetary landings.

**Leaving your system** — you fly to the relay. That travel is the 50 ms. The relay jump itself is near-instant, but you pay the routing cost to leave your network boundary.

**The Citadel** — DNS. Deep in the network, multiple relay hops away. You go once, load the star charts into the nav computer, and never go back just to look up an address.

---

## Anchor 1 — HUD (Fuel Efficiency Gauge)

Overhead display above Joker's seat. Four bands showing the cost of each journey type. Joker glances at it instinctively before answering any question about route cost.

```
█████  1 ms   inter-planet hop     green   — routine, barely registers
█████  5 ms   planetary landing    yellow  — expected destination cost
█████  50 ms  relay travel         orange  — you left the system, was it worth it?
█████  500 ms thrusters only       red     — drive's down, something is broken
```

**Healthy request** — one inter-planet hop, one landing, back. Seven units total. Textbook.

**The needle** rests at 7 ms on a well-planned route. Joker taps the gauge before answering any route question. It's reflex.

**500 ms** is not the ship destroyed — it's the FTL drive down. You fix it by repairing the engine: add the index, fix the N+1, enable connection pooling. Fix the pattern, not the ship.

---

## Anchor 2 — Joker Talking (The Full Request Journey)

Joker is mid-conversation as you enter, already plotting a course. He talks through every request like a flight plan. His tone shifts at each leg — calm for the expected costs, irritated for the avoidable ones.

### The Citadel Visit (DNS)

*"First time we fly anywhere, we need coordinates. That means a trip to the Citadel — walk through to the right district, find the right embassy, get the charts. That's your DNS lookup. Twenty to a hundred milliseconds depending on how far the relay drops you."*

*"We went in 2183. Got everything. It's all in the nav computer. You want me to fly back to the Citadel every single request just to look up Omega's coordinates?"*

The DNS hierarchy = walking through the Citadel to the right office:
- Citadel entrance → recursive resolver
- Right district → root nameserver
- Right embassy → TLD nameserver
- Exact office → authoritative nameserver

**DNS TTL = 0** — Joker's nightmare. The nav computer is configured to forget coordinates after every jump. You're making the Citadel trip every single time. *"Someone broke the nav computer."*

### Establishing the Route (TCP / TLS)

First visit to any destination: you plot the route, establish the connection, get clearance. That costs time — but you do it once. Keep-alive means the route stays warm. You don't re-plot from scratch on every request.

*"Once I've flown a route, it's saved. I don't recalculate the whole thing from zero every jump."*

### The Journey (Steady State)

For every request after the first:

1. **Leave the dock** — request parsed, routed to the right service. One inter-planet hop. 1 ms.
2. **Fly to Omega** — travel to the DB within your system. Expected. 5 ms.
3. **Land, retrieve, ascend** — the DB does the work. With coordinates (index): fast. Without: you're wandering every level asking everyone you pass.
4. **Return** — response serialised, sent back. Same hops in reverse.

### Omega (DB Performance)

Omega is the destination. Once you're there, how long you spend depends entirely on whether you have coordinates:

- **Indexed lookup** — you land, go straight to the contact, back in the shuttle. 1–5 ms.
- **Full table scan** — no map, no contact. You're checking every level, every corridor. 100–10,000 ms.
- **N+1** — you had a list of ten names and made a separate trip to Omega for each one. You could have asked for all ten in one visit.

### Leaving the System (External API)

*"Every time you call an external service mid-request, you're flying to the relay. That's 50 ms before you've even jumped. Was it worth the trip?"*

JWT auth: coordinates already onboard, no travel needed.
OAuth per request: you're flying to the relay every single time to ask someone else if you're allowed to be here.

---

## Anchor 3 — Comms Panel (Auth Costs)

Still to design.

---

## Anchor 4 — EDI Sphere (Parallel vs Sequential)

Still to design.

---

## Still To Do

- Design Anchor 3 — Comms panel (auth costs)
- Design Anchor 4 — EDI sphere (parallel vs sequential)
- Write image prompts once all anchors are finalised
