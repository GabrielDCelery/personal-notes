# Galaxy Map — CIC (Numbers and Quick Math)

The CIC is the nerve centre of the Normandy — where Shepard plots courses and makes high-level decisions. It maps to **quick estimation before committing to a path**.

## Room Layout

```
[Left comms terminal]  [Galaxy map - holographic, center]  [Kelly Chambers - right]
                       [flat workstations around the map]
                       [Pressly's station - left of center]
```

## Route Through the Room

Enter → Galaxy Map → Left Comms Terminal → Pressly's Station → Kelly's Station → Exit

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

## Anchor 2 — Left Comms Terminal (Time Conversion — Drop 5 Zeros)

A device for bridging enormous scale differences. You "dial in" your time period, apply the scaling factor, drop 5 zeros, arrive at RPS.

**The rule:** always drop 5 zeros. Only the scaling factor changes.

| Time Period | Scaling Factor | Operation        |
|-------------|---------------|------------------|
| 1 hour      | ×30           | ×30, drop 5      |
| 1 day       | ×1            | drop 5 (anchor)  |
| 1 month     | ÷30           | ÷30, drop 5      |
| 1 year      | ÷300          | ÷300, drop 5     |

**Peak multiplier:** ×3 — always applied after.

**Example:** 10M requests/day → 10M ÷ 100K = 100 RPS → ×3 = 300 peak RPS

---

## Anchor 3 — Pressly's Station (Throughput Tiers + Scale)

Pressly is standing at his station, slightly hunched, holding a notepad protectively. You are standing behind him looking over his shoulder.

### Notepad (in his hands)

A hand-drawn bar chart in pencil — the ship's limits, what each layer can handle before it breaks:

```
writes  |█          1K    DB write tx/sec
reads   |██████████ 10K   DB read QPS
cache   |████████████████ 100K  cache ops/sec
```

Each bar is 10× longer than the one above. Each tier is 10× apart.

### Terminal (in front of him)

A mission briefing screen — the mission demands, incoming RPS mapped to Mass Effect missions:

```
~100 RPS     → Illium
~1,000 RPS   → Horizon        → index, pool
~10,000 RPS  → [TBD mission]  → cache, replicas
~100,000 RPS → Suicide Mission → distributed, CDN, sharding
~1,000,000   → THE REAPERS   → DARK SPACE
```

### The Narrative

Pressly is doing a pre-jump capacity check. The notepad tells him what the ship can handle. The terminal tells him what the mission requires. He cross-references the two — decomposing incoming RPS into DB reads, DB writes, and cache ops — before committing to a course.

### Key Decisions

- The notepad is **internal layer limits** (ops/sec per tier)
- The terminal is **external incoming RPS** (mission scale)
- The multiplier between them: reads ×3–5 queries/request, writes ×1 tx/request

---

## Anchor 4 — Kelly's Station (Common Mistakes)

Kelly is at her station, facing you, warm and engaged, mid-conversation.

### The Sticky Note

On her monitor, handwritten, slightly crooked:
**"most systems are smaller than you think"**

100K users × 100 actions/day = 10M/day ÷ 100K = **100 RPS**. One server. One DB. Fine.

### Her Expression

That quiet "I told you so" smile. Soft exhale. Eyes flicking briefly to the sticky note then back to you. Someone mapped their 100 RPS app to the wrong mission tier again.

### Her Gestures

- **Right hand** pointing at Pressly — *"did you decompose the load first?"*
- **Left hand — four fingers** — reads multiply (×4 queries per request)
- **Left hand — index and pinky (rock on sign)** — writes (1 pre-check read + 1 write tx) — weird enough to be unforgettable

### Her Glance

Brief glance at the galaxy map — she doesn't say "you're optimising the wrong layer." The map says it. She just looks at it and back at you.

### The Narrative

Kelly is calm, Kelly is always calm. But she's had this conversation before. The sticky note has been there since day one. She's not nagging — she's just waiting for you to notice it.

---

## Image Prompts

### Pressly's Station

A close-up scene inside the CIC of the SSV Normandy SR-2. Navigator Pressly is standing at his workstation, slightly hunched forward, holding a notepad protectively in both hands. The player is standing behind him, looking over his shoulder.

The notepad has a hand-drawn pencil bar chart, three rows, each bar visibly 10× longer than the one above, labelled:
```
writes  |█          1K
reads   |██████████ 10K
cache   |████████████████████ 100K
```
The notepad looks worn, like he has referenced it many times.

On the terminal screen in front of him, a mission briefing display shows a tiered list of mission threat levels with corresponding RPS values:
```
~100 RPS     → Illium
~1,000 RPS   → Horizon
~10,000 RPS  → [mission name]
~100,000 RPS → Suicide Mission
~1,000,000   → THE REAPERS
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

## Still To Do

- Fill in the [TBD mission] for the ~10,000 RPS tier on Pressly's terminal
- Design anchors for the flat workstations (operational latencies still unplaced)
- Generate images once prompts are finalised
