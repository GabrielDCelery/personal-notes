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
5. **EDI sphere** — failure scenarios, when requests cross into seconds

---

## The Entry Scene

You walk into the cockpit. Joker is on the comms panel, mid-handshake with a destination:

_"Normandy to Omega docking authority, do you read? ... Confirmed. Transmitting IFF codes now."_

He holds up a hand without turning — _"with me in a second."_

He closes the channel, glances up at the HUD fuel gauge, plots the route. Once the ship starts moving he relaxes into the seat and starts venting:

_"You know what I hate? When someone sets the nav computer to forget coordinates after every jump. We went to the Citadel in 2183, got the charts, they're in the nav computer. So why am I flying back to the Citadel every single—"_

EDI glows red.

_"Oh, there is worse, Mr. Moreau."_

→ continues in Anchor 4

---

## Anchor 1 — HUD (Route Cost Classification)

Overhead display above Joker's seat. Three bands — the cost of every journey type in normal operations. Joker glances at it instinctively before answering any route question. It's reflex.

```
█████  1 ms    inter-planet hop     green
█████  5 ms    planetary landing    yellow  [sticky note: "77% of your time"]
█████  50 ms   relay travel         orange
```

A sticky note is taped next to the yellow band, slightly crooked: **"77% of your time"**. The landing dominates. Everything else is noise by comparison.

**Healthy request** — one inter-planet hop (1 ms), one planetary landing and back (5 ms). Seven units total. Textbook.

**The HUD only shows normal operations — milliseconds.** When a request crosses into seconds, it is no longer a cost to classify. It is an emergency. That is EDI's territory.

---

## Anchor 2 — Comms Panel (TCP/TLS + Connection Pooling)

### Establishing the Channel (TCP/TLS)

**TCP** = claiming one of Omega's docking bays. The three-way handshake:

1. _"Normandy to destination, do you read?"_ — SYN
2. _"Destination reads, standing by"_ — SYN-ACK
3. _"Acknowledged, en route"_ — ACK

One bay is now yours. The cost is the distance — radio waves travel at the speed of light, same as physical travel. Same DC = 1 ms. Cross-ocean = 150–300 ms. Nothing Joker can do about that. _"That's just how far away they are."_

**TLS** = transmitting IFF codes once the bay is claimed. Security clearance before docking is authorised. You are who you say you are, the channel is encrypted.

**Keep-alive** — once Joker has a bay, he holds it. He does not release it after every visit and re-claim it next time. The bay stays reserved so he can go straight back to Omega without a new handshake.

### Docking Bay Status (Connection Pool)

On the comms panel: Omega's docking bay status board. Shows how many bays are claimed vs still available.

```
── CONNECTIONS ────────────────────────────
Omega     ● ● ● ● ○ ○    4 in use / 2 available
Citadel   ● ○            1 in use / 1 available
Horizon   ● ● ●          all 3 in use — pool exhausted
── IDENTITY ───────────────────────────────
OAuth     ↑ external     relay trip per request
JWT       ✓ local        EDI verifies onboard
```

**Orange** = bay claimed, connection held open. **Green** = bay free, available. **All orange** = pool exhausted — new requests queue until a bay is released.

**No pool at all** = Joker releases his bay after every landing and re-claims it next time. New handshake, new IFF exchange, every single visit. 5–20 ms overhead each time, for nothing.

### Auth Costs (Who Are You?)

Once the channel is open, the next question is identity. Auth is not physical travel — it is a comms call. The cost is how far the signal has to travel.

**JWT** — EDI already holds the public key onboard. She verifies the token locally against a key she already has. No comms needed. No travel. Pure CPU.

**OAuth per request** — you are making a long-distance call to an external authority on every single request. The signal crosses the system boundary, reaches the external station, waits for a response, comes back.

_"Call once, cache the result. Stop burning my comms budget."_

**The rule** — verify once at the edge, pass a trusted internal token downstream. If every service independently calls an external auth station, you pay 50–100 ms per service per request.

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

_"We went in 2183. It's in the nav computer. I am not flying back to the Citadel to look up Omega's coordinates on every single request."_

**DNS TTL = 0** — the nav computer is configured to forget coordinates after every jump. Joker is making the Citadel trip every single request. _"Someone broke the nav computer."_

### Omega (DB Performance)

Omega is the destination. Once you land, how long you spend depends entirely on whether you have coordinates:

- **Indexed lookup** — straight to the contact, back in the shuttle. 1–5 ms.
- **Full table scan** — no map, checking every level, every corridor, asking everyone. 100–10,000 ms.
- **N+1** — you had ten names on the list and made a separate landing for each one. You could have asked for all ten in one visit.

### Leaving the System (External API)

Every time you call an external service mid-request, you fly to the relay first. That travel is 50 ms before you have even jumped.

_"Was it worth the trip? Because you're making it every single request."_

Not every external trip is avoidable — sometimes the mission requires it. The question is whether you are making the trip when you do not need to, or making it on every request when once would do.

---

## Anchor 4 — EDI Sphere (When Requests Cross Into Seconds)

EDI's holographic sphere is to Joker's left, glowing its normal calm blue. She monitors everything. She also enjoys teasing Joker.

**Blue** = milliseconds. Normal operations. Everything on the HUD.
**Red** = seconds. Something is structurally broken. Mission at risk.

The color change is the threshold. No number needed — the moment EDI goes red, you have crossed from "slow" into "fix the pattern."

### The Conversation

EDI interrupts Joker mid-complaint, glows red, and walks through three worse examples in one unbroken chain. Joker gets progressively more exasperated. She finishes anyway.

---

_"If we are listing frustrations, Mr. Moreau — Purgatory still comes to mind. Simple handoff of Jack. Instead we fought through every level of the station. A database without an index works the same way — every row examined in sequence, top to bottom."_

Joker: _"EDI."_

_"I am simply noting the similarity. As I was noting — the planet scanning before the Suicide Mission followed the same pattern. Forty-seven individual return jumps. One pass through each system on the way would have retrieved everything."_

Joker: _"EDI, stop."_

_"Understood, Mr. Moreau. Though Horizon does come to mind. Every relay channel occupied, no route available, requests queuing while the situation deteriorated—"_

Joker: _"I KNOW. I know. Stop."_

EDI returns to blue. Joker stares at the viewport.

---

### What EDI Is Telling You

Each anecdote maps to a broken pattern:

| EDI's scenario                | Broken pattern                  | Fix                         |
| ----------------------------- | ------------------------------- | --------------------------- |
| Fighting through Purgatory    | Missing index — full table scan | Add the index               |
| Revisiting planets one by one | N+1 — one query per item        | Batch the query             |
| Horizon channels all occupied | No connection pool              | Enable PgBouncer / HikariCP |

The fix in every case is the pattern, not the code. EDI knows this. She is waiting for you to know it too.

---

## Image Prompts

### Entry Scene

A wide shot of the cockpit of the SSV Normandy SR-2. The player has just entered from the rear — Joker is in the pilot's seat at the front, back to the camera, slightly turned as he works the comms panel to his right. One hand is raised without turning — "with me in a second."

Through the large wraparound viewport, space stretches ahead — stars and a distant nebula.

Overhead, an HUD display shows three horizontal bands:

- Green band labelled "1 ms — inter-planet hop / intra-DC service call"
- Yellow band labelled "5 ms — planetary landing / database query" — a small crooked sticky note attached reading "77% of your time"
- Orange band labelled "50 ms — relay travel / external API call"

To Joker's left, EDI's holographic sphere glows calm blue. To his right, the comms panel is active. The room is small and functional — every surface has a purpose.

Dark, blue-lit military cockpit aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### HUD Close-up

A close-up of the overhead HUD display above the pilot's seat in the cockpit of the SSV Normandy SR-2. Three horizontal fuel efficiency bands reading top to bottom:

- Green band labelled "1 ms — inter-planet hop / intra-DC service call"
- Yellow band labelled "5 ms — planetary landing / database query" — the brightest of the three, a small handwritten sticky note slightly crooked attached next to it reading "77% of your time"
- Orange band labelled "50 ms — relay travel / external API call"

The yellow band dominates visually. The sticky note draws the eye immediately.

Dark, blue-lit HUD aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### Comms Panel Close-up

A close-up of the communications panel to the right of the pilot's seat in the cockpit of the SSV Normandy SR-2. The panel glows blue, Mass Effect style — clean holographic interface, no physical buttons.

Two sections on the panel, separated by labelled dividers:

```
── CONNECTIONS ────────────────────────────
Omega     ● ● ● ● ○ ○    4 connections in use / 2 available
Citadel   ● ○            1 in use / 1 available
Horizon   ● ● ●          all 3 in use — pool exhausted
── IDENTITY ───────────────────────────────
OAuth     ↑ external     relay trip per request
JWT       ✓ local        EDI verifies onboard
```

Orange indicators glow warm amber — connection claimed, held open via keep-alive. Green indicators glow soft teal — slot available. The IDENTITY section has no dots — OAuth shows an upward arrow indicating it leaves the system boundary, JWT shows a checkmark indicating local verification only. Labels in clean military font. The panel looks active — a working comms station mid-operation.

Dark, blue-lit military cockpit aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### EDI Interrupting Joker

A scene inside the cockpit of the SSV Normandy SR-2. Joker is in the pilot's seat, visibly exasperated — jaw tight, eyes mid-roll, one hand half-raised as if he was about to finish a sentence. He has been interrupted.

EDI's holographic sphere to his left is glowing red — not its usual calm blue. The red light casts a warm glow across the left side of the cockpit, contrasting with the blue instrument lighting everywhere else.

EDI's sphere is calm and still. Joker is not. The contrast is the mood — EDI impassive, Joker seething.

Through the viewport, space is visible ahead. The HUD bands are visible overhead.

Dark cockpit aesthetic with EDI's sphere casting red light on the left side. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

## Still To Do

- Generate images from prompts above
