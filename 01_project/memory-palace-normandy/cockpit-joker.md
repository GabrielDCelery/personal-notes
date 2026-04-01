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

1. **Enter** — mission running long, HUD catches your eye immediately, Joker caught out
2. **HUD overhead** — the accusation, 77% note right above him
3. **Joker deflects** — DNS, "at least I cached the coordinates"
4. **You change subject** — ask about comms
5. **Comms panel** — Joker relaxes, open docks, established channels, false comfort
6. **EDI goes red** — destroys the comfort, escalated failure scenarios

---

## The Entry Scene

You walk into the cockpit. Joker is already defensive before you open your mouth.

_"Before you say anything — the estimate was right. The relay hop, the handshake, the jump time. All correct."_

Above his seat, the HUD is visible. Three bands. The yellow band has a sticky note taped slightly crooked: **"77% of your time."** The mission is running long because Omega took longer than he planned for. The note is right above his head. He put it there himself after the last time.

You point at it.

A pause.

_"...The landing was always in the plan."_

He shifts in the seat. He knows.

_"At least I didn't have to make a Citadel run first. Nav computer had the coordinates cached — we went in 2183, still in there. Some pilots forget that. Not me."_

He says it like it counts for something. The yellow band is still right above him.

You change the subject. Ask about comms.

---

## Anchor 1 — HUD (Route Cost Classification)

Overhead display above Joker's seat. Three bands — the cost of every journey type in normal operations.

```
█████  1 ms    inter-planet hop     green
█████  5 ms    planetary landing    yellow  [sticky note: "77% of your time"]
█████  50 ms   relay travel         orange
```

A sticky note is taped next to the yellow band, slightly crooked: **"77% of your time"**. The landing dominates. Everything else is noise by comparison. Joker plotted the course without accounting for how long he'd spend on Omega. The note has been there the whole time.

**Healthy request** — one inter-planet hop (1 ms), one planetary landing and back (5 ms). Seven units total. Textbook.

**The HUD only shows normal operations — milliseconds.** When a request crosses into seconds, it is no longer a cost to classify. It is an emergency. That is EDI's territory.

_Joker's mistake here: he focused on the relay band and the hops when the yellow band was always going to dominate. The sticky note told him. He didn't look._

---

## Anchor 2 — Joker's Deflection (DNS)

Joker's defence. He got one thing right and he is going to make sure you know it.

You cannot fly to any destination without coordinates. First time you visit anywhere, you go to the Citadel:

Walk through the entrance → find the right district → find the right embassy → get the exact coordinates.

That is the DNS hierarchy:

- Citadel entrance → recursive resolver
- Right district → root nameserver
- Right embassy → TLD nameserver
- Exact office → authoritative nameserver

20–100 ms depending on how far the relay drops you from the Citadel.

_"We went in 2183. It's in the nav computer. I am not flying back to the Citadel to look up Omega's coordinates on every single request."_

**DNS TTL = 0** — the nav computer is configured to forget coordinates after every jump. Joker would be making the Citadel trip every single request. _"Someone broke the nav computer."_ He didn't. The coordinates are cached. This is his alibi.

And we're not leaving the system on every request either. Every external service call means a relay trip first — 50 ms before you've even jumped. Worth it when the mission requires it. Not worth it when you're making it on every request for something you could cache.

_"We're not doing that. So."_

He holds onto it. The yellow band is still right above him.

---

## Anchor 3 — Comms Panel (TCP/TLS + Connection Pooling)

You change the subject. Ask about comms. Joker relaxes slightly — this is ground he's comfortable with.

_"Yeah, we're fine. Comms are established, open docks on Omega. We're mid-mission but there's room."_

He gestures at the panel. The status board backs him up.

### Docking Bay Status (Connection Pool)

```
── CONNECTIONS ────────────────────────────
Omega     ● ● ● ● ○ ○    4 in use / 2 available
Citadel   ● ○            1 in use / 1 available
Horizon   ✕ ✕ ✕          system down
── IDENTITY ───────────────────────────────
OAuth     ↑ external     relay trip per request
JWT       ✓ local        EDI verifies onboard
```

**Orange** = bay claimed, connection held open. **Green** = bay free, available. **Red** = system down, no longer a destination.

Joker can see the Horizon row every time he looks at the panel. He doesn't mention it.

Omega looks healthy. Joker is almost comfortable. He starts explaining how he got here.

### Establishing the Channel (TCP/TLS)

**TCP** = claiming one of Omega's docking bays. The three-way handshake:

1. _"Normandy to destination, do you read?"_ — SYN
2. _"Destination reads, standing by"_ — SYN-ACK
3. _"Acknowledged, en route"_ — ACK

One bay is now yours. The cost is the distance — same DC = 1 ms. Cross-ocean = 150–300 ms. Nothing Joker can do about that. _"That's just how far away they are."_

**TLS** = transmitting IFF codes once the bay is claimed. Security clearance before docking is authorised. You are who you say you are, the channel is encrypted.

**Keep-alive** — once Joker has a bay, he holds it. He does not release it after every visit and re-claim it next time. The bay stays reserved so he can go straight back to Omega without a new handshake.

**No pool at all** = Joker releases his bay after every landing and re-claims it next time. New handshake, new IFF exchange, every single visit. 5–20 ms overhead each time, for nothing.

### Auth Costs (Who Are You?)

Once the channel is open, the next question is identity. Auth is not physical travel — it is a comms call. The cost is how far the signal has to travel.

**JWT** — EDI already holds the public key onboard. She verifies the token locally against a key she already has. No comms needed. No travel. Pure CPU.

**OAuth per request** — you are making a long-distance call to an external authority on every single request. The signal crosses the system boundary, reaches the external station, waits for a response, comes back.

_"Call once, cache the result. Stop burning my comms budget."_

**The rule** — verify once at the edge, pass a trusted internal token downstream. If every service independently calls an external auth station, you pay a relay trip per service per request.

Joker is calm now. The panel looks fine. Open docks, established comms, JWT verified locally. He leans back.

EDI glows red.

---

## Anchor 4 — EDI Sphere (When Requests Cross Into Seconds)

EDI's holographic sphere is to Joker's left, glowing its normal calm blue. She monitors everything. She also enjoys teasing Joker.

**Blue** = milliseconds. Normal operations. Everything on the HUD.
**Red** = seconds. Something is structurally broken. Mission at risk.

The color change is the threshold. No number needed — the moment EDI goes red, you have crossed from "slow" into "fix the pattern."

She goes red exactly when Joker relaxes. She has been listening the whole time.

### The Conversation

EDI glows red and walks through three scenarios in one unbroken chain — each one an escalated version of something already in the room. Joker gets progressively more exasperated. She finishes anyway.

---

_"I note you mentioned open docks, Mr. Moreau. That is the current state. I can offer context on alternative states, if it would be useful."_

Joker: _"It would not."_

_"Purgatory still comes to mind. Simple handoff of Jack. Instead we fought through every level of the station. A database without an index works the same way — every row examined in sequence, top to bottom. The docks were never the problem. Getting to the right bay was."_

Joker: _"EDI."_

_"I am simply noting the similarity. As I was noting — the planet scanning before the Suicide Mission followed the same pattern. Forty-seven individual return jumps. One pass through each system on the way would have retrieved everything. Open docks are irrelevant if you are making forty-seven trips."_

Joker: _"EDI, stop."_

_"Understood, Mr. Moreau. Though Horizon does come to mind. Every relay channel occupied, no route available, requests queuing while the situation deteriorated. Your panel shows two open docks. Until it doesn't."_

Joker: _"I KNOW. I know. Stop."_

EDI returns to blue. Joker stares at the viewport.

---

### What EDI Is Telling You

Each anecdote is an escalated version of something already in the room:

| EDI's scenario                | Broken pattern                  | Room anchor                 | Fix                         |
| ----------------------------- | ------------------------------- | --------------------------- | --------------------------- |
| Fighting through Purgatory    | Missing index — full table scan | Omega, indexed vs full scan | Add the index               |
| Revisiting planets one by one | N+1 — one query per item        | Omega, N+1 landings         | Batch the query             |
| Horizon channels all occupied | No connection pool              | Comms panel, pool exhausted | Enable PgBouncer / HikariCP |

The fix in every case is the pattern, not the code. EDI knows this. She is waiting for you to know it too.

---

## Image Prompts

### Entry Scene

A wide shot of the cockpit of the SSV Normandy SR-2. The player has just entered from the rear. Joker is in the pilot's seat, back to the camera, not turning around — posture slightly tense, aware he is being watched.

Overhead, the HUD shows three horizontal bands clearly visible from behind Joker's head:

- Green band labelled "1 ms — inter-planet hop"
- Yellow band labelled "5 ms — planetary landing" — brightest of the three, a small crooked sticky note reading "77% of your time"
- Orange band labelled "50 ms — relay travel"

The yellow band and sticky note are directly above Joker's head. He is not looking at it.

To Joker's left, EDI's holographic sphere glows calm blue. To his right, the comms panel is active. Through the large wraparound viewport, space stretches ahead.

Dark, blue-lit military cockpit aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### HUD Close-up

A close-up of the overhead HUD display above the pilot's seat in the cockpit of the SSV Normandy SR-2. Three horizontal bands reading top to bottom:

- Green band labelled "1 ms — inter-planet hop / intra-DC service call"
- Yellow band labelled "5 ms — planetary landing / database query" — the brightest of the three, a small handwritten sticky note slightly crooked attached next to it reading "77% of your time"
- Orange band labelled "50 ms — relay travel / external API call"

The yellow band dominates visually. The sticky note draws the eye immediately. It looks like it has been there a long time.

Dark, blue-lit HUD aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### Comms Panel Close-up

A close-up of the communications panel to the right of the pilot's seat in the cockpit of the SSV Normandy SR-2. The panel glows blue, Mass Effect style — clean holographic interface, no physical buttons.

Two sections on the panel, separated by labelled dividers:

```
── CONNECTIONS ────────────────────────────
Omega     ● ● ● ● ○ ○    4 connections in use / 2 available
Citadel   ● ○            1 in use / 1 available
Horizon   ✕ ✕ ✕          system down
── IDENTITY ───────────────────────────────
OAuth     ↑ external     relay trip per request
JWT       ✓ local        EDI verifies onboard
```

Orange indicators glow warm amber — connection claimed, held open via keep-alive. Green indicators glow soft teal — slot available. The Horizon row is all red with ✕ markers — system down, no longer a destination. The contrast between the active rows above and the dead Horizon row is immediate. The IDENTITY section shows OAuth with an upward arrow indicating it leaves the system boundary, JWT with a checkmark indicating local verification only.

The panel looks active — but something on it is not.

Dark, blue-lit military cockpit aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### EDI Interrupting Joker

A scene inside the cockpit of the SSV Normandy SR-2. Joker is in the pilot's seat, visibly exasperated — jaw tight, eyes mid-roll, one hand half-raised as if he was about to finish a sentence. He has been interrupted mid-relaxation.

EDI's holographic sphere to his left is glowing red — not its usual calm blue. The red light casts a warm glow across the left side of the cockpit, contrasting with the blue instrument lighting everywhere else.

EDI's sphere is calm and still. Joker is not. The contrast is the mood — EDI impassive, Joker seething.

Through the viewport, space is visible ahead. The HUD bands are visible overhead, yellow band and sticky note still prominent.

Dark cockpit aesthetic with EDI's sphere casting red light on the left side. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

## Still To Do

- Generate images from prompts above
