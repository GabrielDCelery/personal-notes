# Mordin's Lab — (Databases)

Mordin's lab is where the precise work happens. He doesn't deal in rough capacity tiers — that's Pressly's job. Mordin knows the mechanisms. Why reads are fast. Why writes are slow. Why an index turns 1,000,000 comparisons into 30. Why Postgres needs PgBouncer when nothing else does.

Maps to: **03-databases.md**

## Theme

Precision. Mordin knows the internals — not just what the limits are, but exactly why they exist and what controls them.

## Narrative Ordering Principle

Anchors ordered by severity — most catastrophic and irreversible first, most mechanical and fixable last. Mordin opens with controlled urgency and relaxes into precision as the topics become recoverable.

1. **Wrong database pick** — irreversible, months to undo, driven by asking the wrong question. MongoDB because migrations are annoying. DynamoDB because it scales. The embed/reference trap. The access pattern lock-in. This is the one that keeps him up at night.
2. **Missing index** — weeks of mystery slow queries, invisible until production, fixable in hours once found
3. **No connection pool** — Postgres forking 500 processes, 5GB gone before a query runs, fixable but only after something breaks
4. **Write amplification from indexes** — you added indexes to fix reads and made writes worse, schema-level decision with real cost
5. **RAM vs disk** — wrong instance type, you're running to the filing room on every query until you fix it

The wrong database is Mordin's genophage — a decision made early that determines everything that follows. He'd rather be the one who makes sure you understand it correctly than leave it to someone who might get it wrong.

## Room Layout

```
                    [BACK WALL DISPLAY — Which Database]
[Whiteboard]   [Central Workstation — Mordin]   [Corner — Lab Assistants]
[Scaling]      [Two Trays]  [Directory]          [PgBouncer / MVCC shelf]
               [Side bench — sample storage]
                            [ENTER]
```

Compact lab. You enter from the lower deck corridor. Mordin is immediately visible at the central workstation, back slightly turned, muttering. Everything is within a few steps of each other — the room forces you to look at everything.

## Route Through the Room

1. **Enter** — Mordin mid-monologue, running through the write path out loud
2. **Central workstation — two trays** — reads vs writes, the asymmetry
3. **Side bench** — bench, cabinet, filing room: RAM vs disk
4. **Directory on the workstation corner** — indexes, leftmost prefix rule, B-tree
5. **Back wall display** — which database decision framework
6. **Whiteboard** — scaling in order, numbered, strict
7. **Back corner** — lab assistants crowding the door, MVCC shelf, PgBouncer supervisor

---

## Entry Scene

You step into the lab. Mordin is at his central workstation, back half-turned, rapid-firing to himself:

_"WAL first. Durability — survives a crash. Then indexes — all of them, every one on the table. Five indexes: five updates. Then FK constraints — requires reads, yes, those cost time. Then locks, acquire, release. Then replicate. Five steps. Every write. Never fewer."_

He registers your presence without fully turning.

_"Yes? How can I help?"_

_"Wanted to talk about our next mission."_

He turns. Looks at you properly.

_"Ah. Next mission. Yes. Horizon — barely survived that one. Indexes helped. Connection pooling helped. Vertical scale helped. Correct order, good instincts. No more obvious fixes left now. Someone suggested MongoDB? DynamoDB perhaps? Scales well, apparently."_

A beat. Not unkind. He glances around the lab.

_"Had a clinic on Omega. No funding. Mercenaries. Constant interruptions. Solved a viral outbreak with what was on the bench. Didn't need better equipment. Needed to understand the problem."_

He looks back at you.

_"Most people skip the mechanism. Go straight to 'which database is best.' Wrong question. You have Postgres. You have indexes. You have a connection pool. You don't fully understand any of them yet. Come in."_

→ he gestures toward the workstation

---

## The Three Samples — Database Decision Framework

Three specimen containers on the bench. Mordin picks them up one by one during the entry conversation.

---

### Colonist Sample (MongoDB)

Mordin picks it up, examines it.

_"Remarkable species. Humans. Adaptable. Show up everywhere, adjust to anything. Alliance expanded faster than anyone predicted. No fixed structure — that is the strength. Genuine strength."_

He sets it down slightly, continues.

_"Also makes the records a nightmare. Collecting medical data for the investigation — had to visit seven colony administrations. Seven. Varren samples, one landing, everything I needed. Colonists — relay trip after relay trip. Different formats, different systems, incomplete cross-references."_

He picks it back up.

_"Someone embedded colony status directly into the medical records. Convenient at the time. Horizon happened. Status changed. Now half the records say active colony. Can't trust the location data. Should have referenced the colony independently — one update, consistent everywhere. Instead — hunting through records one by one."_

Brief glance at the varren sample.

_"Cross-referencing medical history with location data. Simple question, should be a landing — five milliseconds. Wrong tool for it. Fifty milliseconds minimum. Hundreds at scale. Unacceptable at Horizon. Worse beyond."_

**What the colonist sample encodes:**

- Flexible schema — humans are adaptive, no fixed structure, Alliance expanded fast. Right tool when data is genuinely document-shaped and naturally distributed
- Auto-sharding — colonists live across different planets, data naturally distributes the same way
- Embed vs reference trap — embedded colony status, Horizon happened, data stale, hunting through records one by one
- Slow JOINs — cross-referencing is a relay trip not a landing. 50ms minimum, hundreds at scale

---

### Collector Sample (DynamoDB)

Mordin picks it up. A brief pause — different energy than the colonist sample. More careful.

_"Collectors. Protheans once. Reapers engineered out everything except the function they needed. Extraordinarily efficient. Single purpose. No waste."_

He sets it under the analysis equipment.

_"Collecting samples — one mission, straight to what I needed. No relay trips, no colony administrations, no cross-referencing. O(1). Every time. Regardless of how many samples exist."_

He examines the readout.

_"Knew my access patterns going in. Always querying by the same key. Perfect tool for that. Then they adapted. New weapon variant, different mechanism. These samples answer the old questions perfectly."_

He sets it down.

_"New questions — useless. Had to go back out. Dangerous. Time consuming. Months."_

A beat.

_"Tried working with the old samples anyway. Application layer — extract what I can, combine it manually in my analysis. Works. Barely. Every researcher who needs the same combined data repeats the same process. At scale — forty-seven planet scans."_

_"Has workarounds. Access patterns are fixed. Workarounds exist. Still limited."_

Brief glance at the varren sample.

**What the Collector sample encodes:**

- O(1) lookups regardless of scale — one mission, straight to what you need, no cross-referencing
- Zero ops overhead — scales to zero, perfect for Lambda/serverless, no connection limits
- Access pattern lock-in — Collectors adapted, old samples useless, months to collect new ones. Changing access patterns means redesigning the table
- Application layer joins are N+1 — manually combining extractions, logic repeated everywhere, breaks at scale
- No cheap transactions — cross-referencing multiple samples simultaneously requires special handling, expensive, limited scope
- GSIs exist but limited — workarounds for anticipated questions, duplicates the entire dataset, still can't ask something you didn't prepare for

---

### Varren Sample (Postgres)

Mordin picks up the last container. Different energy — no pause, no caution. Familiar.

_"Varren. Originally from Tuchanka. Followed the krogan everywhere — every colony, every warzone, every station. Not because anyone brought them deliberately. Because they adapt. Pack hunters when prey is available. Scavengers when outnumbered. Omnivores. Work in almost any environment."_

He sets it under the analysis equipment.

_"STG designated them species 408. Studied them for potential weaponisation. Consensus among most xenobiologists: pest, nuisance, manage and contain. STG conclusion: underestimated. More capable than reputation suggested. Didn't need to be exotic to be effective."_

He glances at the collector sample, then back.

_"Krogan have kept varren for millennia. Fight them for territory. Then tame them. Keep them as companions, war beasts. Love-hate. Understandable. They need management — they make a mess, you clean it up. Don't clean up, the mess accumulates. But they work. In almost any environment, against almost any problem. You don't need to know exactly what you're hunting before you bring one."_

A brief pause.

_"Pack hunters. Multiple varren pursuing different prey simultaneously — they don't get in each other's way. Readers and writers, same table, same moment. Never blocking each other. That is MVCC. Cost: dead tissue accumulates after each kill. Vacuum process cleans it. Heavy hunting — vacuum struggles to keep up. Performance degrades."_

He sets the container down next to the other two.

_"One limitation. Pack gets too large — no coordination mechanism. Each varren operates independently. You cannot direct five hundred simultaneously across different territories from one position. No built-in sharding. Above fifty thousand writes per second on a single node, vacuum overhead becomes the ceiling."_

He looks at all three containers.

_"Wrong question: which database is best. Right question: what are your access patterns. Don't know yet? Use the varren. It can hunt almost anything. Optimise when the patterns emerge."_

**What the varren sample encodes:**

- Adaptability — handles CRUD, analytics, JSONB, full-text, CTEs, window functions. Works in almost any environment
- Default choice — when access patterns are unknown, Postgres first. Optimise later
- Underestimated — species 408. Most engineers reach for exotic tools before exhausting what Postgres can do
- MVCC — pack hunters don't block each other. Readers never block writers. Cost: dead tuples accumulate, vacuum cleans them
- Needs management — PgBouncer mandatory, vacuum overhead, process-per-connection. Love-hate. It rewards the care
- Hard ceiling — no built-in sharding, struggles above ~50k writes/sec on one node. Pack gets too large to coordinate

---

## Transition — From Samples to Index Research

Mordin sets the three containers to one side. Not away — they stay on the bench, reference material. He turns back to the workstation, pulls a stack of documents toward him. Migration schema. Query logs. Several things annotated in red.

_"Choice is made. Postgres. Correct choice. Now the harder part."_

He doesn't look up.

_"Knowing the varren can hunt anything is not the same as knowing where to point it."_

---

## Anchor 1 — Central Workstation: The Two Trays (Reads vs Writes)

Mordin turns from the documents. Gestures at the two trays on the workstation.

_"You saw all of that. Now you can see why."_

Two trays side by side. The asymmetry is obvious at a glance.

**Left tray — READ:** One sample container. Pick it up. Return it. Done.

**Right tray — WRITE:** An assembly line. Labeled stations, left to right:

```
[WAL recorder] → [Index × 5] → [FK validator] → [Lock/release] → [Replication tube]
```

1. WAL recorder — crash-proof copy, written first every time, always disk
2. Index updater × 5 — one station per index on the table (nine indexes = nine updates)
3. FK validator — reads inside the write, warm cache fast, cold cache hits disk again
4. Lock mechanism — acquire before work, release after
5. Replication tube — sends copy to secondary lab (if synchronous: wait for signal back)

The right tray is crowded. The left tray has one item.

_"Same table. Same data. Different direction. More indexes — wider the gap. Not a fixed ratio. Schema determines everything."_

---

## Anchor 2 — Side Bench: Sample Storage (RAM vs Disk)

Three locations in the lab. Three latencies. You can see all three from where you're standing.

| Location                      | What it is | Latency  |
| ----------------------------- | ---------- | -------- |
| Bench surface                 | RAM        | ~0.01 ms |
| Wall cabinet (across the lab) | SSD        | ~0.1 ms  |
| Filing room down the hall     | HDD        | ~5 ms    |

A label is stuck on the bench surface: `r family (memory-optimised)`.
A crossed-out label on the shelf above: `m family (compute-optimised)` — wrong choice for a database.

Mordin: _"One question. Is the working set on the bench or down the hall? Everything else follows. If it fits in RAM, reads are fast. If it doesn't, every read might leave the lab entirely. Get bench space. CPU speed is irrelevant if you're running to the filing room on every query."_

---

## Anchor 3 — Research Index: Schema, Indexes and the Write Path

The stack of documents on the workstation. Migration notes, query logs, schema diagrams. Several pages have columns crossed out and redrawn. One page has a single word underlined three times: **ORDER.**

Mordin picks up a query log, sets it down again.

_"Medical records. Moved from MongoDB. Structure is correct now — no more embedded colony status, no more stale location data. Joins are fast. Cross-referencing works. Good."_

A pause.

_"First query across the full dataset. Filtering by colony of origin. One million records. Took four seconds. Unacceptable — I need to run this hundreds of times."_

He taps the schema diagram.

_"Missing index. Added it. Query: eight milliseconds. Correct. Then I added indexes for every query pattern I could anticipate. Every species combination. Every date range. Every symptom cluster. Nine indexes total."_

He picks up the query log.

_"Then I re-seeded. New collector weapon data — bulk insert, one million records. Took three hours. Previous seed: forty minutes."_

He sets it down.

_"Knew why immediately. Every write to this table — every single insert — hits the WAL first. Disk. Always. Crash survival, non-negotiable. Then updates every index on the table. Nine indexes: nine separate write operations per record. Then validates FK constraints — colony of origin references the colonies table, species references the species table. Reads inside every write. If those tables aren't in memory: disk again. Then locks, acquire and release. One million records. Every step, every time."_

He pulls out the laminated card.

_"Made this after the third rebuild."_

```
Query by species + colony      →  index (species, colony)   ✓
Query by species alone         →  index (species, colony)   ✓  leftmost prefix
Query by colony alone          →  index (species, colony)   ✗  must rebuild
```

_"Composite index column order is fixed. Sorted left to right. Must start from the left column. Had it backwards twice. Two rebuilds before I made the card."_

He sets it down and gestures at the schema.

_"Two decisions that determine everything. What is a column, what stays JSONB. Symptom presentation — varies per species, changes as I learn more about the weapon. JSONB. Colony of origin, species, collection date — queried constantly, always the same shape. Columns. Indexable."_

_"Wrong choice either direction: JSONB on something you filter constantly — slow. Column on something with variable structure — migration every time the weapon evolves."_

He looks at the query logs.

_"Trimmed to four indexes. Re-seed: fifty-five minutes. Queries still fast. Acceptable balance. Write-heavy workload — I am always inserting. Every index is a tax on every write. The missing index was a problem. Nine indexes was a worse one."_

He sets the documents down.

_"The varren can handle it. That was never the question. The question is what you ask it to carry."_

**What this anchor encodes:**

- Missing index turns milliseconds into seconds — invisible until you run the query at scale
- Every write hits WAL first — disk, always, non-negotiable
- Every index is an extra write operation per insert — nine indexes, nine extra writes per record
- FK constraints are reads inside writes — warm cache fast, cold cache hits disk
- Composite index column order is fixed — leftmost prefix rule, wrong order means rebuild
- JSONB vs column: variable structure vs queryable shape
- Write-heavy workloads feel index cost hardest — balance is everything

Pinned to the edge of the workstation: the index types table, annotated in Mordin's handwriting:

| Index type | Good for                      | Bad for                 |
| ---------- | ----------------------------- | ----------------------- |
| B-tree     | Equality, range, sorting      | Full-text               |
| Hash       | Exact equality only           | Range, sorting          |
| GIN        | Full-text, JSONB, arrays      | Simple equality, writes |
| Partial    | Filtering rare conditions     | General queries         |
| Composite  | Queries on all/prefix of cols | Non-prefix columns      |

Mordin's annotation at the bottom: _"Don't index everything. Write-heavy tables: as few as possible."_

---

## Anchor 4 — Back Wall Display: Which Database

Large screen at the back of the lab. Mordin's decision framework — not preference, trade-offs. Three columns.

```
                 Query flexibility ──────────────────►
                 ▲
Ops overhead     DynamoDB        MongoDB          Postgres
◄──────────────  zero ops,       flexible schema  most flexible
                 limited         good scaling     most ops overhead
                 queries
```

| Workload                             | Use          | Why                                        |
| ------------------------------------ | ------------ | ------------------------------------------ |
| Generic web app (CRUD, JOINs)        | **Postgres** | Most flexible, handles everything          |
| Key-value at massive scale           | DynamoDB     | Zero ops, single-digit ms at any scale     |
| Schema varies per record             | MongoDB      | Document model, no migrations              |
| High write throughput, need to shard | Mongo/Dynamo | Both shard natively, Postgres doesn't      |
| Complex analytics                    | **Postgres** | Best query planner, CTEs, window functions |
| Serverless (Lambda)                  | DynamoDB     | No connection limits, scales to zero       |
| You don't know yet                   | **Postgres** | It can do everything                       |

A large label across the top of the screen, Mordin's annotation:

**"Wrong question: which is best. Right question: what are your access patterns."**

Under Postgres: _"You don't know yet? Use Postgres. Optimise when patterns emerge."_

---

## Anchor 5 — Whiteboard: Scaling in Order

Side wall. Mordin's diagnostic sequence. Numbered. Strict. He has underlined "IN ORDER" twice.

```
Is the database slow?

  1. Slow queries        → EXPLAIN ANALYZE, add indexes, fix N+1   (hours, free)
  2. Too many conns      → PgBouncer                               (hours, free)
  3. CPU/memory maxed    → bigger instance                         (minutes, $$)
  4. Read QPS too high   → read replicas → caching (Redis)         (days, $$$)
  5. Write QPS too high  → reduce indexes, batch → shard           (weeks, hard)
  6. Data too large      → partition tables, archive cold data     (days)
```

Written at the bottom in red: **"SKIP NOTHING. Each step is ~10× harder than the last."**

Side note in smaller text: _"Vertical scale: 10 min from 5K to 50K reads/sec. Sharding: weeks. Run the numbers first. Sharding is a last resort."_

---

## Anchor 6 — Back Corner: Lab Assistants (Connection Pooling + Postgres Internals)

The back corner of the lab. Two things to look at.

### The Doorway Crowd (Process-per-connection)

A crowd of lab assistants standing in the doorway. Each one has a badge: **10 MB**. That's their footprint just for existing — before any work begins.

- **Postgres:** one process per connection. 500 connections = 500 assistants in the doorway = 5 GB consumed before a single query runs.
- **MySQL:** one assistant running between stations. Far lighter. Not the same model.

**Pool sizing note** pinned to the corner cabinet:
`connections ≈ (2 × CPU cores) + disks`
For a 4-core instance: ~10 connections is right.

Mordin: _"Process-per-connection. Postgres. One process, ten megabytes, one connection. Not optional knowledge. PgBouncer is not an optimisation — it is mandatory at any meaningful scale."_

### The PgBouncer Supervisor

A supervisor figure standing at the door. Outside, 1,000 callers are waiting. The supervisor takes requests and hands them to the 20 real assistants behind him. The callers never know they're sharing. The real assistants stay in the lab — they are never dismissed and re-hired.

New connection cost without pool: **3–13 ms** (TCP handshake + TLS + auth + Postgres forking a process).
Connection from pool: **~0.05 ms** (grab the handle, query, return it).

### The Accumulating Shelf (MVCC)

On the back shelf: a stack of old experiment logs. Not cleaned up. Each one is a row version superseded by an update — kept alive by MVCC until the vacuum process runs. Heavy write load fills this shelf fast.

MVCC rule: readers never block writers. Old row versions accumulate as the cost. Vacuum cleans them. If vacuum can't keep up, dead tuples pile up and slow everything down.

**Read replica tube** in the wall: sends copies to the secondary lab. The secondary lab receives them **10–100 ms** later.

Mordin: _"For most reads: acceptable lag. For 'show account balance after transfer': primary lab only."_

---

## Image Prompts

### Entry Scene

A compact science lab aboard the SSV Normandy SR-2 — Mordin Solus's personal research space on the lower deck. The room is small and functional, every surface covered in equipment and samples.

Mordin is at the central workstation, back slightly turned to the camera, speaking rapidly to himself. One hand gestures toward a cluttered right-side tray with multiple labeled stations arranged in a line. A much simpler left-side tray has a single sample container.

The room is lit blue-white — cold, precise lab lighting overlaid with the warm amber of Mordin's Salarian complexion. Equipment, sample containers, a large display screen visible at the back.

Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### The Two Trays Close-up

A close-up of the central workstation in Mordin's lab. Two laboratory trays side by side.

Left tray: nearly empty. One sealed sample container. A small label reads "READ: find + return."

Right tray: crowded assembly line, five labeled stations in sequence with arrows between them:

```
[WAL recorder] → [Index ×5] → [FK validator] → [Lock/release] → [Replication tube]
```

The contrast is immediate — one tray sparse, one tray full. The right tray's stations are all active, indicators lit. The replication tube runs off the edge of the workstation toward a second lab.

Cold lab lighting. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### Sample Storage (RAM vs Disk)

A wide shot of Mordin's lab showing three storage locations in the same frame:

- **Foreground:** lab bench surface, clear and accessible. Label: `r family (memory-optimised)`. Latency marker: `~0.01 ms`.
- **Mid-ground:** wall cabinet on the far side of the lab, closed. Latency marker: `~0.1 ms SSD`.
- **Background:** a door leading to a filing room down the corridor, slightly ajar, dim light beyond. Latency marker: `~5 ms HDD`.

On the shelf above the bench: a label with a red strikethrough reading `m family (compute-optimised)`.

The visual gradient from near to far implies the latency gap — each step farther is an order of magnitude slower.

Cold lab lighting. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### The Directory (Indexes)

A close-up of the corner of Mordin's workstation. A large, thick directory labelled "Omega Residents — sorted: last name, first name." It is open to a middle section.

Beside it, a laminated card showing two rows:

```
Without index, 1 billion rows:   ████████████████████████████████████ ~1,000,000,000
B-tree index, 1 billion rows:    ·  ~30
```

The contrast between the two rows is stark — the first bar extends off the card, the second is a single dot.

Pinned to the workstation edge: the index types table, handwritten annotation at the bottom: "Write-heavy tables: as few as possible."

Cold lab lighting. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

### Back Corner: Lab Assistants

A wide shot of the back corner of Mordin's lab. Near the doorway, a crowd of lab assistants stands waiting — not working, just present. Each wears a badge reading "10 MB." There are many of them. The doorway is crowded.

Standing at the door between the crowd and the lab: a single supervisor figure (PgBouncer). Calm. Twenty real assistants work inside the lab; the crowd outside is kept at bay.

On the back shelf to the right: a growing stack of old experiment logs, labelled "dead tuples." The stack is visibly taller than it should be.

Cold lab lighting, slightly more cramped and cluttered than the rest of the lab. Mass Effect 2 visual style. Cinematic, photorealistic lighting.

---

## Still To Do

- Design room layout — DONE
- Design anchors — DONE
- Write image prompts — DONE
- Generate AI images from prompts above
