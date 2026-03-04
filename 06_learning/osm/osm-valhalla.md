# OSM, Valhalla, and Spatial Indexing

## R-Tree

An R-tree is a tree structure for indexing **spatial data** (rectangles, polygons, points in 2D/3D space).

### Core Idea

Group nearby objects into **bounding boxes**, then nest those boxes hierarchically.

```
Level 0 (root):
┌─────────────────────────────────┐
│              R0                 │
│  ┌─────────────┐ ┌───────────┐  │
│  │     R1      │ │    R2     │  │
│  └─────────────┘ └───────────┘  │
└─────────────────────────────────┘

Level 1:
┌─────────────┐     ┌───────────┐
│ R1          │     │ R2        │
│ ┌───┐ ┌───┐ │     │ ┌───┐┌──┐ │
│ │ A │ │ B │ │     │ │ C ││D │ │
│ └───┘ └───┘ │     │ └───┘└──┘ │
└─────────────┘     └───────────┘

Level 2 (leaf):
A, B, C, D = actual objects (points, rectangles, polygons)
```

### Structure

| Level | Contains |
|-------|----------|
| Root | Bounding boxes of children |
| Internal nodes | Bounding boxes of children |
| Leaf nodes | Actual objects + their bounding boxes |

### Search Example

Find all objects intersecting query box Q:

```
       Q (query)
       ┌───┐
       │   │
       └───┘

1. Start at root
2. Does Q intersect R1? Yes → descend into R1
3. Does Q intersect R2? No  → skip R2
4. In R1: Does Q intersect A? Yes → return A
5. In R1: Does Q intersect B? No  → skip B

Result: [A]
```

### Bounding Boxes Are Rectangles

Bounding boxes are **axis-aligned rectangles**, not matching the polygon shape.

```
Actual road (polyline):          Bounding box:
                                 ┌─────────────┐
      ╱╲                         │     ╱╲      │
     ╱  ╲                        │    ╱  ╲     │
    ╱    ╲____                   │   ╱    ╲____│
                                 └─────────────┘
                                 (min_x, min_y) to (max_x, max_y)
```

### Bounding Boxes Can Overlap

```
France bbox:          Spain bbox:
┌─────────────┐
│             │
│   France    │
│         ┌───┼───────────┐
│         │   │           │
└─────────┼───┘           │
          │     Spain     │
          │               │
          └───────────────┘

Overlap zone: contains neither country's actual geometry
```

Overlap increases false positives, not wrong results. The R-tree first checks bounding boxes (fast), then exact geometry (slow) only on candidates.

### Common Uses

- PostGIS (spatial queries in PostgreSQL)
- Game engines (collision detection)
- Maps (find restaurants near me)
- CAD systems

---

## OSM Data Model

### Node

A point with lat/lon coordinates only.

### Way

An ordered list of nodes defining a road, with tags for attributes.

```
OSM Way #12345:
  nodes: [1, 2, 3, 4, 5]
  tags: { highway: "primary", name: "Main St", maxspeed: "50" }

     1       2       3       4       5
     ○───────○───────○───────○───────○
```

### Speed Limits in OSM

Speed limits are stored as **tags on the way**, not on individual nodes.

When speed changes mid-road, the way is split:

```
Way #12345:                    Way #12346:
  nodes: [1, 2, 3]              nodes: [3, 4, 5]
  tags:                         tags:
    highway: primary              highway: primary
    maxspeed: 50                  maxspeed: 80

○───────○───────○───────○───────○
1       2       3       4       5
└───────────────┘└──────────────┘
    Way 12345        Way 12346
     50 km/h          80 km/h

        Node 3 shared by both ways
```

---

## Valhalla Data Model

### Node

A point where roads connect (intersection, endpoint, etc.)

```
        │
        │
────────○────────    ○ = node (intersection)
        │
        │
```

### Edge

A road segment between two nodes.

```
○─────────────────○
│                 │
node              node
(start)           (end)

Edge contains:
- Road geometry (shape between nodes)
- Speed limit
- Road class (highway, residential, etc.)
- One-way?
- Turn restrictions
- Name
```

### OSM Ways → Valhalla Edges

One OSM way can become multiple Valhalla edges:

```
OSM Way #12345 "Main Street":

○───────○───────○───────○───────○───────○
1       2       3       4       5       6
                │               │
                │               │
                ○               ○
            Other roads intersecting


Valhalla splits at intersections:

Edge A        Edge B        Edge C
○─────────────○─────────────○─────────────○
1             3             5             6
              │             │
              │             │
              ○             ○
```

### Edges Split At

| Reason | Example |
|--------|---------|
| Intersection | Road crossing |
| Speed limit change | 50 → 80 zone |
| Road class change | residential → primary |
| Name change | "Main St" → "Highway 1" |
| Lane count change | 2 lanes → 4 lanes |
| Other attribute changes | Surface, access restrictions |

This ensures each edge has uniform attributes (single speed, single road class, etc.)

---

## Valhalla Tile System

Valhalla uses a **fixed grid**, not R-tree bounding boxes.

### Tile Hierarchy

```
Zoom level 2 (highways):
┌───────────────────────────────────────┐
│                                       │
│              Large tile               │
│                                       │
└───────────────────────────────────────┘

Zoom level 1 (arterial):
┌───────────────────┬───────────────────┐
│                   │                   │
│    Medium tile    │    Medium tile    │
│                   │                   │
└───────────────────┴───────────────────┘

Zoom level 0 (local):
┌─────────┬─────────┬─────────┬─────────┐
│  Small  │  Small  │  Small  │  Small  │
│  tile   │  tile   │  tile   │  tile   │
├─────────┼─────────┼─────────┼─────────┤
│  Small  │  Small  │  Small  │  Small  │
│  tile   │  tile   │  tile   │  tile   │
└─────────┴─────────┴─────────┴─────────┘
```

### What's Inside Each Tile

```
┌─────────────────────────────────┐
│  Tile (2,1) Level 0             │
│                                 │
│     ○───────○                   │
│     │       │                   │
│  ○──┼───○───┼──○    ○ = nodes   │
│     │       │       ─ = edges   │
│     ○───────○                   │
│         │                       │
│         ○──────────○            │
│                                 │
│  + edge costs, restrictions,    │
│    turn data, names, etc.       │
└─────────────────────────────────┘
```

### Hierarchical Routing

```
A ════════════════════════════════════════► B

Level 0        Level 2 (highways)       Level 0
┌────┐    ┌─────────────────────────┐    ┌────┐
│ A──┼────┼─════════════════════════┼────┼──B │
│    │    │        Highway          │    │    │
└────┘    └─────────────────────────┘    └────┘
 Find        Fast traversal across       Find
 on-ramp     long distance               off-ramp
```

### Edges Crossing Tile Boundaries

Edges can cross tile boundaries. Valhalla handles this with edge references:

```
┌─────────────────────┬─────────────────────┐
│                     │                     │
│  Tile A             │  Tile B             │
│                     │                     │
│         ○──────────────────────○          │
│         │           │          │          │
│       Node 1        │        Node 2       │
│                     │                     │
│  Edge: Node1 → ?    │  Edge: ? → Node2    │
│  (points to Tile B) │  (points to Tile A) │
│                     │                     │
└─────────────────────┴─────────────────────┘
```

### Why Tiles?

| Use | Benefit |
|-----|---------|
| On-demand loading | Low memory footprint |
| Caching | Fast repeated queries |
| Distribution | Scalable serving |
| Incremental updates | Only rebuild affected tiles |
| Hierarchy | Fast long-distance routing |

### R-tree vs Valhalla Tiles

| R-tree | Valhalla tiles |
|--------|----------------|
| Boxes fit data | Fixed grid, data fits into boxes |
| Dynamic, adapts to density | Static, same size everywhere |
| For spatial queries | For routing graph partitioning |
| Hierarchical tree | Flat grid with hierarchy levels |
