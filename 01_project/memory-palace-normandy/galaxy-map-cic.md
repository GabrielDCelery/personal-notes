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
2. **Comms terminal** (front left) — abandoned, flooding with noise, tame it first, then convert traffic to RPS
3. **Walk to Pressly** — brief exchange about Horizon, silence, the weight of the circled writes
4. **Pressly's terminal** — mission tiers, identify where you are
5. **Pressly's notepad** — capacity limits, the cross-reference
6. **Kelly interjects** — reads the room, steps in before you fill the silence, redirects you to her
7. **Kelly's station** — sticky note, hand gestures, objections as her own observations

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

_Kelly's objection here: you optimised the wrong layer. You spent a week on something that lives in the inner ring while your bottleneck is in the outer ring. The map told you. You didn't look._

---

## Anchor 2 — Comms Terminal (Time Conversion — Drop 5 Zeros)

You enter the CIC and the terminal is already going. Nobody is minding it. Raw request numbers scroll fast across the display — a flood of figures, no context, no order. Comms static underneath, just loud enough to be irritating. The galaxy map is right there but you can't focus while this is running.

The time period selector is flashing. It wants you to act first.

You set the time period. The flood slows. The static cuts. The room settles.

**Three steps, left to right:**

```
hour  ×30
day   ×1  ◄  →  [ DROP 5 ]  →  [ ×3 PEAK ]
month ÷30
year  ÷300
```

You tap your time period — it pulses brighter, the others dim. The flood of numbers collapses to a single day-equivalent figure. DROP 5 wipes the zeros. ×3 PEAK pulses once. One clean RPS sits large on the screen.

The terminal goes quiet. You carry the number to Pressly.

_Kelly's objection here: you designed the system before you stopped at this terminal. You architected for a scale you hadn't calculated. The terminal was flooding the whole time and you walked straight past it._

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
~100 RPS     → Illium              → ship it.           back it up.
~1K RPS      → Horizon             → index, pool.       alerting.
~10K RPS     → Loyalty Missions    → cache.             replica. watch it.
~100K RPS    → Suicide Mission     → async, shard, queues. no single points of failure.
~1M RPS      → THE REAPERS         → DARK SPACE
```

Two threads per line, separated by a period. Left is the scale answer. Right is the resilience answer.

**How the current ship maps against each mission:**

| Mission             | Read QPS needed | Write tx/sec needed | Can we handle it?             |
| ------------------- | --------------- | ------------------- | ----------------------------- |
| Illium (~100 RPS)   | 400 QPS         | 100 tx/sec          | ✓ Pressly is relaxed          |
| Horizon (~1K RPS)   | 4K QPS          | 1K tx/sec           | ⚠ writes right at ceiling     |
| Loyalty (~10K RPS)  | 40K QPS         | 10K tx/sec          | ✗ 4× over ceiling — not ready |
| Suicide (~100K RPS) | 400K QPS        | 100K tx/sec         | ✗ needs full overhaul         |
| The Reapers         | —               | —                   | DARK SPACE                    |

**Narrative arc:**

- **Illium** — routine, nothing needed. back it up and go home.
- **Horizon** — scrappy survival, patch what you have, no new components ← we are here
- **Loyalty Missions** — the architecture doesn't change shape, but vertical scaling has stopped being the answer. introduce a helper for each pressure point: CDN for static load, Redis for hot reads, read replicas for resilience, horizontal app servers for HA. harden each component individually so it survives what's next. a component you didn't prepare is Grunt without a loyalty mission — technically present, functionally a liability.
- **Suicide Mission** — you're not building a traditional app anymore. synchronous writes that survived Loyalty Missions will fail here. placing an order means writing to the DB, emitting an event, and an orchestrator picks it up. the app accepts work, it doesn't complete it. every component needs a backup plan. the one role you didn't prepare for is the one that kills you.
- **The Reapers** — you didn't build for this, everything is custom. assume failure, test it deliberately.

_Kelly's objection here: you looked at the terminal, saw "cache" further down the list, and started building Redis. You're at Horizon. The terminal told you exactly what Horizon needs — index and pool. You skipped the line. You built a Loyalty Mission solution for a Horizon problem and your tables still don't have indexes._

### The Narrative

You fixed the terminal. You have the number. You walk over to Pressly.

He doesn't look up. He's already cross-referencing the notepad against the terminal — the writes circled in red, the mission tier staring back at him. You mention Horizon. That you made it.

He taps the circled row with one finger. Doesn't say anything for a moment.

_"Barely."_

Then he goes quiet. Back to the notepad. The silence sits there.

You open your mouth. Kelly gets there first.

_Kelly's objection here: you gave Pressly your RPS and forgot to decompose it. You handed him one number and he's been cross-referencing against write capacity — but your number was requests, not transactions. Writes cost more. You didn't account for that before arriving._

### Key Decisions

- The notepad is **current ship capacity** (what we can handle right now)
- The terminal is **mission demand** (what we're being asked to handle)
- Multipliers live in Pressly's head — Kelly makes sure they live in yours too

---

## Anchor 4 — Kelly's Station (Common Mistakes)

Pressly said "barely" and went quiet. You were about to say something. Kelly's voice comes across the room before you do — warm, easy, as if she just thought of something.

_"Actually — while you're here. I wanted to run something by you."_

It wasn't a question. She's already redirected you. You walk over. Pressly never has to deal with whatever you were about to say.

### The Sticky Note

On her monitor, handwritten, slightly crooked:
**"most systems are smaller than you think"**

100K users × 100 actions/day = 10M/day ÷ 100K = **100 RPS**. One server. One DB. Fine.

**The wordplay:** Kelly's note is visible from the galaxy map — the literal map of star systems. You're standing there looking at the entire galaxy, thousands of star systems, and her note says "most systems are smaller than you think." The galaxy map makes the joke land. All that scale, all that vastness, and Kelly's point is: yours isn't as big as you think it is.

### Her Expression

That quiet "I told you so" smile. Soft exhale. Eyes flicking briefly to the sticky note then back to you. She's not auditing you — she's thinking out loud. With you. The corrections land as observations, not verdicts. Pressly's competence is never in the room.

### Her Gestures

- **Right hand — four fingers** — reads ×4 queries per request
- **Left hand — two fingers** — writes ×2 (1 pre-check read + 1 write tx)

Both hands up simultaneously. You see the multipliers at a glance before she says a word. She's not pointing at Pressly. She's just showing you something she noticed.

### Her Glance

Brief glance at the galaxy map — she doesn't say "you're optimising the wrong layer." The map says it. She just looks at it and back at you.

### The Narrative

Kelly saw the silence coming before you did. She pulled you out of it cleanly, brought you to her, and walked you through everything Pressly would have had to correct — without ever making it his problem. The sticky note, the gestures, the glance at the galaxy map. All her own observations. All landed gently. You leave knowing exactly what you got wrong and somehow still feeling fine about it.

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

A wide shot of the CIC of the SSV Normandy SR-2, focus on the comms terminal to one side. The terminal is unattended — no crew member present. Its display is flooded with raw scrolling numbers, fast and chaotic, filling the screen with no structure. The time period selector on the left side of the terminal flashes urgently, demanding input.

The player's hand reaches toward the selector. The "day" row is about to be tapped.

A second image shows the terminal after input: the flood has stopped. The display now shows the three-step sequence cleanly, left to right:

```
hour  ×30
day   ×1  ◄ (highlighted, steady)  →  [ DROP 5 ]  →  [ ×3 PEAK ]
month ÷30
year  ÷300
```

A single large RPS figure sits prominently on the right side of the screen — the result. The terminal is calm. The rest of the CIC is visible in the background, quiet.

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

On the terminal screen in front of him, a mission briefing display shows a tiered list of mission threat levels with corresponding RPS values. Each line has two parts — a scale answer and a resilience answer, separated by a period:

```
~100 RPS     → Illium              → ship it.            back it up.
~1K RPS      → Horizon             → index, pool.        alerting.
~10K RPS     → Loyalty Missions    → cache.              replica. watch it.
~100K RPS    → Suicide Mission     → async, shard, queues. no single points of failure.
~1M RPS      → THE REAPERS         → DARK SPACE
```

The screen has a faint red glow at the bottom tier. Pressly looks tense, cross-referencing the notepad against the terminal as if calculating whether the ship can handle the mission ahead.

Dark, blue-lit military CIC aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting. No other characters visible.

---

### Kelly's Station

A scene inside the CIC of the SSV Normandy SR-2. Yeoman Kelly Chambers is at her station, turned towards the player, mid-conversation. Warm, engaged, completely at ease. She called you over and you came — that's the dynamic.

On her monitor, a handwritten sticky note slightly crooked reads: **"most systems are smaller than you think"**

Her expression is a soft knowing smile — not correcting, not judging. Thinking out loud. Her eyes flick briefly to the sticky note then back to you, as if she just remembered it was there.

Both hands are raised towards the player simultaneously. Her right hand shows four fingers clearly extended — reads ×4. Her left hand shows two fingers — writes ×2. The multipliers are visible at a glance, before she says a word. The gesture feels like sharing, not lecturing.

In the background, Pressly is hunched over his notepad at his station, undisturbed. The holographic galaxy map glows in the centre of the room behind Kelly's shoulder.

The overall mood is warm and precise — Kelly is the most socially intelligent person on the ship and she just quietly made sure nobody had an awkward moment.

Dark, blue-lit military CIC aesthetic. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

## Content Moved to Other Rooms

- **Operational latencies** (Redis, Postgres, external APIs) → Cockpit (Joker) — "how fast can we get there"
- **Size and bandwidth anchors** (500MB fits in RAM, 1TB → S3) → Engineering (Tali) — physical infrastructure, what the ship can carry
- **Bottleneck types** (DB = throughput, external API = latency) → Cockpit (Joker)

## Still To Do

- Generate images once prompts are finalised
