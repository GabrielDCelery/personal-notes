# Mordin's Lab — (Databases)

Mordin's lab is where the precise work happens. He doesn't deal in rough capacity tiers — that's Pressly's job. Mordin knows the mechanisms. Why reads are fast. Why writes are slow. Why an index turns 1,000,000 comparisons into 30. Why Postgres needs PgBouncer when nothing else does.

Maps to: **03-databases.md**

## Theme

Precision. Mordin knows the internals — not just what the limits are, but exactly why they exist and what controls them.

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

_"You're here about the database. Good. Most people skip the mechanism, go straight to 'which database is best.' Wrong question. Come in."_

→ he gestures toward the workstation

---

## Anchor 1 — Central Workstation: The Two Trays (Reads vs Writes)

Two trays side by side on the workstation. Left and right. The asymmetry is obvious at a glance.

**Left tray — READ:**
One sample container. Pick it up. Return it. Done.

**Right tray — WRITE:** An assembly line. Labeled stations, left to right:

```
[WAL recorder] → [Index × 5] → [FK validator] → [Lock/release] → [Replication tube]
```

1. WAL recorder — crash-proof copy, written first every time
2. Index updater × 5 — one tray per index on the table (five indexes = five updates)
3. FK validator — must do reads to check foreign key integrity
4. Lock mechanism — acquire before work, release after
5. Replication tube — sends copy to secondary lab (if synchronous: wait for signal back)

The right tray is crowded. The left tray has one item.

Mordin: _"Table with ten indexes. Read: one step. Write: twelve steps. Same table. Same data. Different direction. More indexes means wider gap — not fixed ratio. Schema determines everything."_

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

## Anchor 3 — Directory on the Workstation Corner (Indexes)

A large thick directory on the corner of the workstation. Omega residents. Sorted by last name, then first name.

**The phonebook test:**

- You have last name → open to that section, scan. Fast. ~30 pages.
- You have last name + first name → one entry. Done.
- You have only first name → check every entry in the entire directory.

That is the leftmost prefix rule. A composite index on `(user_id, created_at)` supports `user_id` alone or `user_id + created_at`, but not `created_at` alone. The directory is sorted left to right. You must start from the left column.

Below the directory: a laminated B-tree card:

```
Without index, 1 billion rows:   scan all 1,000,000,000
B-tree index, 1 billion rows:    30 comparisons
```

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
