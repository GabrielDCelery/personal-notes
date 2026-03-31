# Galaxy Map — CIC (Numbers and Quick Math)

The CIC is the nerve centre of the Normandy — where Shepard plots courses and makes high-level decisions. It maps to **quick estimation before committing to a path**.

## Room Layout

```
             [ENTER (elevator)]
[Kelly Chambers][Galaxy Map][Comms Terminal]
                            [Pressly's Station]
```

You enter from the top. Galaxy map is centre. Kelly is immediately to your right, comms terminal is further right, Pressly is beyond the terminal. Flat workstations around the galaxy map — empty, negative space.

## Route Through the Room

1. **Galaxy map** — classify the operation (which speed world?)
2. **Comms terminal** (front left) — convert traffic numbers to RPS
3. **Pressly's terminal** — identify the mission tier
4. **Pressly's notepad** — check the ship's capacity limits
5. **Kelly intercepts** — "did you decompose the load first?" four fingers, rock on
6. **Walk to Kelly** — sticky note, "most systems are smaller than you think"

It's a story, not a list. You're pulled through the room naturally.

---

## Anchor 1 — Galaxy Map (Three Speed Worlds)

The holographic galaxy map shows three concentric glowing rings radiating outward:

- **Inner ring** (bright white) — `ns` nanoseconds — CPU, RAM
- **Middle ring** (softer blue) — `μs` microseconds — SSD, local I/O
- **Outer ring** (dim, distant) — `ms` milliseconds — network, disk

The 1,000× gap between each ring is implied by the increasing distance between them.

**Visceral anchor:** if RAM = 1 second, cross-region network = 5.7 days.

**Key insight:** any operation touching the network is in a different league. Rewriting in Go saves nothing if the DB is the bottleneck.

---

## Anchor 2 — Comms Terminal (Time Conversion — Drop 5 Zeros)

A passing stop — you glance at it on the way to Pressly, convert your raw traffic figure to RPS, arrive with the right number.

**Three steps, left to right:**

```
hour  ×30
day   ×1  ◄  →  [ DROP 5 ]  →  [ ×3 PEAK ]
month ÷30
year  ÷300
```

Holographic panel, Mass Effect style — glowing blue. You tap your time period, it pulses brighter, the others dim. The time period selector is always step one on the left, ×3 PEAK is always step three on the right. Day is highlighted by default — the anchor, the most common case.

The sequence reads left to right: convert to requests per day, drop 5 zeros to get RPS (because ~100K seconds in a day), multiply by 3 for peak.

**Examples:**

- 10M requests/day → drop 5 → 100 → ×1 → 100 RPS → ×3 = 300 peak
- 300M requests/month → drop 5 → 3,000 → ÷30 → 100 RPS → ×3 = 300 peak
- 360K requests/hour → drop 5 → 3.6 → ×30 → 108 RPS → ×3 = 324 peak
- 1B requests/year → drop 5 → 10,000 → ÷300 → 33 RPS → ×3 = 100 peak (still Illium)

---

## Anchor 3 — Pressly's Station (Throughput Tiers + Scale)

Pressly is standing at his station, slightly hunched, holding a notepad protectively. You are standing behind him looking over his shoulder.

### Current Ship State

This is post-Horizon. The ship has been through its first real fight and Pressly has already done the basic hardening:

- Single app server
- r6g.xlarge DB
- Indexes in place
- Connection pooling (PgBouncer) running
- No cache yet. No read replicas. No horizontal scaling.

### Notepad (in his hands)

A hand-drawn bar chart in pencil — the ship's current capacity limits:

```
writes  |█          1K    ← circled (we are right at the limit)
reads   |██████████ 10K
cache   |████████████████ 100K  ?
         * r6g.xlarge
```

Each bar is 10× longer than the one above. Each tier is 10× apart.

- **Writes are circled** — we survived Horizon but one traffic spike and we're over
- **Reads are comfortable** — underlined as the next thing to watch
- **Cache has a question mark** — Pressly looked at Loyalty Missions, did the math in his head, and added it before anyone asked. Not in place yet, but he knows exactly what it would give us. When we hit Loyalty Missions he'll cross out the question mark without saying a word.

No multipliers on the notepad — Pressly does that in his head. Kelly makes sure you've done it too before he starts his check.

### Terminal (in front of him)

A mission briefing screen — incoming RPS mapped to Mass Effect missions, cross-referenced against the notepad:

```
~100 RPS     → Illium              →
~1,000 RPS   → Horizon             → bigger server, index, pool
~10,000 RPS  → Loyalty Missions    → cache, replicas, monitoring
~100,000 RPS → Suicide Mission     → distributed, CDN, sharding
~1,000,000   → THE REAPERS         → DARK SPACE
```

**How the current ship maps against each mission:**

| Mission             | Read QPS needed | Write tx/sec needed | Can we handle it?             |
| ------------------- | --------------- | ------------------- | ----------------------------- |
| Illium (~100 RPS)   | 400 QPS         | 100 tx/sec          | ✓ Pressly is relaxed          |
| Horizon (~1K RPS)   | 4K QPS          | 1K tx/sec           | ⚠ writes right at ceiling     |
| Loyalty (~10K RPS)  | 40K QPS         | 10K tx/sec          | ✗ 4× over ceiling — not ready |
| Suicide (~100K RPS) | 400K QPS        | 100K tx/sec         | ✗ needs full overhaul         |
| The Reapers         | —               | —                   | DARK SPACE                    |

**Narrative arc:**

- **Illium** — routine, nothing needed, Kelly has the answer
- **Horizon** — scrappy survival, patch what you have, no new components ← we are here
- **Loyalty Missions** — invest in each component to unlock full potential, one by one
- **Suicide Mission** — everyone has a role, single points of failure get people killed, full coordination
- **The Reapers** — you didn't build for this, everything is custom, good luck

### The Narrative

Pressly is doing a pre-jump capacity check. The notepad is the ship's current limits. The terminal is the mission demand. He cross-references the two. The writes are circled — he's worried. We survived Horizon but only just. The next mission requires upgrades before we jump.

You ask "can we handle the next mission?" Kelly jumps in before he answers.

### Key Decisions

- The notepad is **current ship capacity** (what we can handle right now)
- The terminal is **mission demand** (what we're being asked to handle)
- Multipliers live in Pressly's head — Kelly makes sure they live in yours too

---

## Anchor 4 — Kelly's Station (Common Mistakes)

Kelly is at her station, facing you, warm and engaged, mid-conversation.

### The Sticky Note

On her monitor, handwritten, slightly crooked:
**"most systems are smaller than you think"**

100K users × 100 actions/day = 10M/day ÷ 100K = **100 RPS**. One server. One DB. Fine.

**The wordplay:** Kelly's note is visible from the galaxy map — the literal map of star systems. You're standing there looking at the entire galaxy, thousands of star systems, and her note says "most systems are smaller than you think." The galaxy map makes the joke land. All that scale, all that vastness, and Kelly's point is: yours isn't as big as you think it is.

### Her Expression

That quiet "I told you so" smile. Soft exhale. Eyes flicking briefly to the sticky note then back to you. Someone mapped their 100 RPS app to the wrong mission tier again.

### Her Gestures

- **Right hand** pointing at Pressly — _"he's got the ceilings, but did you give him the right numbers?"_
- **Left hand — four fingers** — reads ×4 queries per request
- **Left hand — rock on sign (index and pinky)** — writes ×2 (1 pre-check read + 1 write tx)

One flowing motion: point at Pressly → turn to you → four fingers → rock on. Kelly's correction loop.

### Her Glance

Brief glance at the galaxy map — she doesn't say "you're optimising the wrong layer." The map says it. She just looks at it and back at you.

### The Narrative

Kelly is calm, Kelly is always calm. But she's had this conversation before. The sticky note has been there since day one. She's not nagging — she's just waiting for you to notice it.

---

## Image Prompts

### Galaxy Map

A wide cinematic shot of the holographic galaxy map at the centre of the CIC aboard the SSV Normandy SR-2. The map dominates the room — a large glowing blue holographic projection floating above a circular pedestal.

The map shows three concentric rings radiating outward from a bright central core:

- **Inner ring** — bright white, tight around the core, labelled `ns`
- **Middle ring** — softer blue, further out, labelled `μs`
- **Outer ring** — dim and distant, barely visible at the edges, labelled `ms`

The distance between each ring is visibly much larger than the last — implying the 1,000× gap between each world. The outer ring feels unreachably far from the centre.

Flat workstations around the map are empty and dark — negative space. No crew present. The room is quiet, the map is the only thing that matters.

Dark, blue-lit military CIC aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### Comms Terminal

A close-up of a holographic communications terminal inside the CIC of the SSV Normandy SR-2. The terminal glows blue, Mass Effect style — no physical buttons, pure light.

The display shows a three-step sequence reading left to right:

```
hour  ×30
day   ×1  ◄ (highlighted, pulsing)  →  [ DROP 5 ]  →  [ ×3 PEAK ]
month ÷30
year  ÷300
```

The time period selector is on the left — four rows, day highlighted brighter than the others. DROP 5 is in the centre in a bright panel. ×3 PEAK is on the right in a bright panel. The arrow flow left to right is clear.

The terminal looks like a passing stop — functional, minimal, designed to be read in seconds.

Dark, blue-lit military CIC aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### Pressly's Station

A close-up scene inside the CIC of the SSV Normandy SR-2. Navigator Pressly is standing at his workstation, slightly hunched forward, holding a notepad protectively in both hands. The player is standing behind him, looking over his shoulder.

The notepad has a hand-drawn pencil bar chart, three rows, each bar visibly 10× longer than the one above, labelled:

```
writes  |█          1K  ← circled in red
reads   |██████████ 10K
cache   |████████████████████ 100K
         * r6g.xlarge
```

The 1K writes row is circled in red pencil — emphasis that the ship is right at the limit. Below the chart, a small annotation reads "r6g.xlarge". The notepad looks worn, like he has referenced it many times.

On the terminal screen in front of him, a mission briefing display shows a tiered list of mission threat levels with corresponding RPS values:

```
~100 RPS     → Illium              →
~1,000 RPS   → Horizon             → index, pool, bigger server
~10,000 RPS  → Loyalty Missions    → cache, replicas, monitoring
~100,000 RPS → Suicide Mission     → distributed, CDN, sharding
~1,000,000   → THE REAPERS        → DARK SPACE
```

The screen has a faint red glow at the bottom tier. Pressly looks tense, cross-referencing the notepad against the terminal as if calculating whether the ship can handle the mission ahead.

Dark, blue-lit military CIC aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting. No other characters visible.

---

### Kelly's Station

A scene inside the CIC of the SSV Normandy SR-2. Yeoman Kelly Chambers is at her station on the right side of the CIC, turned towards the player, mid-conversation. Warm, engaged, slightly exasperated in the gentlest possible way.

On her monitor, a handwritten sticky note slightly crooked reads: **"most systems are smaller than you think"**

Her expression is a soft, knowing smile — not happy, not angry. The "I told you so" exhale of someone who has had this conversation many times before. Her eyes flick briefly to the sticky note then back to you.

Her right hand is pointing across the CIC towards Pressly's station. Her left hand is raised with two gestures visible — four fingers extended on one beat, then index finger and pinky only (the rock on sign) on the next, as if counting off two different things.

In the background, Pressly is hunched over his notepad at his station, oblivious to the conversation. The holographic galaxy map glows in the centre of the room.

The overall mood is warm but pointed — Kelly is the most approachable person on the ship but she is absolutely right and she knows it.

Dark, blue-lit military CIC aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

## Content Moved to Other Rooms

- **Operational latencies** (Redis, Postgres, external APIs) → Cockpit (Joker) — "how fast can we get there"
- **Size and bandwidth anchors** (500MB fits in RAM, 1TB → S3) → Engineering (Tali) — physical infrastructure, what the ship can carry
- **Bottleneck types** (DB = throughput, external API = latency) → Cockpit (Joker)

## Still To Do

- Generate images once prompts are finalised
