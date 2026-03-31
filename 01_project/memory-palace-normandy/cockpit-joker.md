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

## The Geography

Distance = latency. The cost of any operation maps directly to how far you have to travel:

| Cost      | Journey                                               | Represents                           |
| --------- | ----------------------------------------------------- | ------------------------------------ |
| 1 ms      | Inter-planet hop — sub-light burn to a nearby station | Intra-DC service call, load balancer |
| 5 ms      | Planetary landing — descent, retrieve, ascend         | DB query with index                  |
| 50 ms     | Flying to the relay — leaving your system boundary    | External API call per request        |
| 20–100 ms | Citadel visit — first time only                       | Cold DNS lookup                      |

**Within your system** (same datacenter) — you never need a relay. Station hops and planetary landings only.

**Leaving your system** — you fly to the relay. The travel to the relay is the 50 ms. The jump itself is near-instant, but you pay the routing cost to cross the network boundary.

**The Citadel** — DNS. Deep in the network, multiple relay hops away. You go once, load the star charts into the nav computer, and never go back just to look up an address. Radio waves travel at the speed of light — a comms signal to a distant destination takes the same time as physically travelling there. You cannot beat physics.

---

## The Entry Scene

You walk into the cockpit. Joker is on the comms panel, mid-handshake with a destination:

_"Normandy to Omega docking authority, do you read? ... Confirmed. Transmitting IFF codes now."_

He holds up a hand without turning — _"with me in a second."_

He closes the channel, glances up at the HUD fuel gauge, plots the route. Once the ship starts moving he relaxes into the seat and starts venting:

_"You know what I hate? When someone sets the nav computer to forget coordinates after every jump. We went to the Citadel in 2183, got the charts, they're in the nav computer. So why am I flying back to the Citadel every single request just to look up Omega's coordinates? Someone broke the nav computer."_

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

The comms panel is to Joker's right. You walk in on him using it.

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
Omega     ● ● ● ● ○ ○    four claimed, two available — comfortable
Citadel   ● ○            one claimed, one available
Horizon   ● ● ●          all three claimed — tight
```

**Orange** = bay claimed, connection held open. **Green** = bay free, available. **All orange** = pool exhausted — new requests queue until a bay is released.

**No pool at all** = Joker releases his bay after every landing and re-claims it next time. New handshake, new IFF exchange, every single visit. 5–20 ms overhead each time, for nothing.

### Auth Costs (Who Are You?)

Once the channel is open, the next question is identity. Auth is not physical travel — it is a comms call. The cost is how far the signal has to travel.

```
JWT                 EDI verifies internally       no call made        ~0.1 ms
Session → Redis     ping to a nearby station      one round trip      ~1 ms
Session → DB        ping requiring a landing      network + query     ~5 ms
OAuth               long-distance call to         crosses system      10–100 ms
                    an external authority         boundary
```

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

_"Was it worth the trip? Because you're making it every single request."_

Not every external trip is avoidable — sometimes the mission requires it. The question is whether you are making the trip when you do not need to, or making it on every request when once would do.

---

## Anchor 4 — EDI Sphere (When Requests Cross Into Seconds)

EDI's holographic sphere is to Joker's left, glowing its normal calm blue. She monitors everything. She also enjoys teasing Joker.

**Blue** = milliseconds. Normal operations. Everything on the HUD.
**Red** = seconds. Something is structurally broken. Mission at risk.

The color change is the threshold. No number needed — the moment EDI goes red, you have crossed from "slow" into "fix the pattern."

### The Three Anecdotes

EDI switches to red and narrates each failure scenario with complete calm. Joker gets progressively more exasperated.

---

**Scenario 1 — Stranded on Omega (missing index / full table scan)**

EDI glows red.

_"If the database has no index on that column, I calculate we will be stranded on Omega for approximately 8 seconds while the query scans every record. Mission success probability drops to—"_

Joker: _"EDI."_

_"I was simply providing context."_

---

**Scenario 2 — Back and forth between systems (N+1)**

EDI glows red.

_"If each item is retrieved with a separate query, we will make 47 individual relay jumps instead of one. Total accumulated travel time: 2.3 seconds. I also note the fuel cost will—"_

Joker: _"EDI, stop."_

_"Understood."_ She does not stop immediately.

---

**Scenario 3 — No comms (no connection pool / pool exhausted)**

EDI glows red.

_"If connection pooling is disabled, we will re-establish a new comms channel on every single request. At current traffic volume I estimate we will spend more time on handshakes than on actual—"_

Joker: _"I KNOW. I know. Stop."_

EDI returns to blue. Joker stares at the viewport.

---

### What EDI Is Telling You

Each anecdote maps to a broken pattern:

| EDI's scenario        | Broken pattern                  | Fix                         |
| --------------------- | ------------------------------- | --------------------------- |
| Stranded on Omega     | Missing index — full table scan | Add the index               |
| 47 relay jumps        | N+1 — one query per item        | Batch the query             |
| Re-establishing comms | No connection pool              | Enable PgBouncer / HikariCP |

The fix in every case is the pattern, not the code. EDI knows this. She is waiting for you to know it too.

---

## Still To Do

- Write image prompts once all anchors are finalised
