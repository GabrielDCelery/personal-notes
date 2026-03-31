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

1. **Enter** — Joker is mid-comms on the panel, establishing a channel
2. **Comms panel** — TCP/TLS handshake in progress, connection pool indicators visible
3. **HUD overhead** — route cost classification, Joker glances at it as he plots the course
4. **Joker talking** — full request journey as a flight plan, DNS complaint during travel
5. **EDI sphere** — still to design

---

## The Geography

Distance = latency. The cost of any operation maps directly to how far you have to travel:

| Cost | Journey | Represents |
| ---- | ------- | ---------- |
| 1 ms | Inter-planet hop — sub-light burn to a nearby station | Intra-DC service call, load balancer |
| 5 ms | Planetary landing — descent, retrieve, ascend | DB query with index |
| 50 ms | Flying to the relay — leaving your system boundary | External API call per request |
| 20–100 ms | Citadel visit — first time only | Cold DNS lookup |
| 500 ms | Thrusters only — FTL drive is down | Broken data access: no index, N+1, no pool |

**Within your system** (same datacenter) — you never need a relay. Station hops and planetary landings only.

**Leaving your system** — you fly to the relay. The travel to the relay is the 50 ms. The jump itself is near-instant, but you pay the routing cost to cross the network boundary.

**The Citadel** — DNS. Deep in the network, multiple relay hops away. You go once, load the star charts into the nav computer, and never go back just to look up an address. Radio waves travel at the speed of light — a comms signal to a distant destination takes the same time as physically travelling there. You cannot beat physics.

---

## The Entry Scene

You walk into the cockpit. Joker is on the comms panel, mid-handshake with a destination:

*"Normandy to Omega docking authority, do you read? ... Confirmed. Transmitting IFF codes now."*

He holds up a hand without turning — *"with me in a second."*

He closes the channel, glances up at the HUD fuel gauge, plots the route. Once the ship starts moving he relaxes into the seat and starts venting:

*"You know what I hate? When someone sets the nav computer to forget coordinates after every jump. We went to the Citadel in 2183, got the charts, they're in the nav computer. So why am I flying back to the Citadel every single request just to look up Omega's coordinates? Someone broke the nav computer."*

---

## Anchor 1 — HUD (Route Cost Classification)

Overhead display above Joker's seat. Four bands — the cost of every journey type. Joker glances at it instinctively before answering any route question. It's reflex.

```
█████  1 ms    inter-planet hop     green   — routine, barely registers
█████  5 ms    planetary landing    yellow  — expected destination cost
█████  50 ms   relay travel         orange  — you left the system, was it worth it?
█████  500 ms  thrusters only       red     — FTL is down, something is broken
```

**Healthy request** — one inter-planet hop (1 ms), one planetary landing and back (5 ms). Seven units total. Textbook.

**500 ms is not the ship destroyed** — it's the FTL drive down. You fix it by repairing the engine: add the index, fix the N+1, enable the connection pool. Fix the pattern, not the ship.

---

## Anchor 2 — Comms Panel (TCP/TLS + Connection Pooling)

The comms panel is to Joker's right. You walk in on him using it.

### Establishing the Channel (TCP/TLS)

**TCP** = opening the comms channel. The three-way handshake:
1. *"Normandy to destination, do you read?"* — SYN
2. *"Destination reads, standing by"* — SYN-ACK
3. *"Acknowledged, en route"* — ACK

**TLS** = transmitting IFF codes. Security clearance before docking is authorised. You are who you say you are, the channel is encrypted.

The cost is the distance — radio waves travel at the speed of light, same as physical travel. Same DC = 1 ms. Cross-ocean = 150–300 ms. Nothing Joker can do about that. *"That's just how far away they are."*

**Keep-alive** — once the channel is open, Joker leaves it open. He does not re-hail a destination he is already talking to. He does not re-exchange IFF codes on every message.

### Connection Pool (Channel Indicators)

On the comms panel: a row of channel indicators for each destination Joker talks to frequently.

```
Omega     ● ● ● ● ● ●    six channels — all green, lines open
Citadel   ● ●            two channels — both green
Horizon   ● ● ●          three channels — one orange (busy)
```

**Green** = channel open, ready. **Orange** = busy, carrying a request. **All orange** = pool exhausted — new requests queue until a channel frees up.

*"I keep six channels open to Omega at all times. You need something there, the line's already open — no handshake, no IFF exchange, straight through. But if all six are busy? Now you're waiting. And if someone set the idle timeout too short and my channels keep dropping? I'm re-establishing every time — five to twenty milliseconds, every single request."*

**No pool at all** = re-establishing from scratch on every request. Joker re-hailing, re-exchanging IFF codes, every single time. The destination hasn't changed. The route hasn't changed. It's pure waste.

---

## Anchor 3 — Joker Talking (The Full Request Journey)

Joker plots the course and talks through it. Every request is a flight plan.

### The Citadel Visit (DNS)

You cannot fly to any destination without coordinates. First time you visit anywhere, you go to the Citadel:

Walk through the entrance → find the right district → find the right embassy → get the exact coordinates.

That is the DNS hierarchy:
- Citadel entrance → recursive resolver
- Right district → root nameserver
- Right embassy → TLD nameserver
- Exact office → authoritative nameserver

20–100 ms depending on how far the relay drops you from the Citadel.

*"We went in 2183. It's in the nav computer. I am not flying back to the Citadel to look up Omega's coordinates on every single request."*

**DNS TTL = 0** — the nav computer is configured to forget coordinates after every jump. Joker is making the Citadel trip every single request. *"Someone broke the nav computer."*

### The Journey (Every Request After the First)

1. **Leave the dock** — request parsed, routed to the right place. One inter-planet hop. 1 ms.
2. **Fly to Omega** — travel to the DB within your system. In-system, no relay needed. 1 ms.
3. **Land, retrieve, ascend** — the DB does the work. With coordinates: fast. Without: you're wandering.
4. **Return** — serialised response, same hops in reverse.

### Omega (DB Performance)

Omega is the destination. Once you land, how long you spend depends entirely on whether you have coordinates:

- **Indexed lookup** — straight to the contact, back in the shuttle. 1–5 ms.
- **Full table scan** — no map, checking every level, every corridor, asking everyone. 100–10,000 ms.
- **N+1** — you had ten names on the list and made a separate landing for each one. You could have asked for all ten in one visit.

### Leaving the System (External API)

Every time you call an external service mid-request, you fly to the relay first. That travel is 50 ms before you have even jumped.

*"Was it worth the trip? Because you're making it every single request."*

JWT = coordinates already onboard, no travel needed. Verified locally by EDI against a key she already holds.
OAuth per request = flying to the relay every single request to ask someone else if you are allowed to be here.

---

## Anchor 4 — EDI Sphere

Still to design.

---

## Still To Do

- Decide what EDI sphere anchors (auth costs, or parallel vs sequential)
- Design Anchor 4 — EDI sphere
- Write image prompts once all anchors are finalised
