# Engineering — Tali (Large Data and Migrations)

Engineering is where Tali works — the physical infrastructure of the ship. What can the ship carry? What gets stored on board vs offloaded to a depot? How do you migrate without blowing up the drive core?

Maps to: **07-large-data-and-migrations.md**

## Theme

Size, storage, bandwidth, migrations — the physical limits of what the system can hold and how you move data safely.

## Content to Anchor

- **Size anchors** — 1M rows × 500B = 500MB (fits in RAM), 1M items × 1MB = 1TB (→ S3)
- **Bandwidth anchors** — 1 Gbps ≈ 100K RPS of 1KB responses, or 1K RPS of 1MB responses
- **Rule** — blobs go to S3, everything else fits in RAM on a single machine
- **Large data and migrations content** — from 07-large-data-and-migrations.md

## Still To Do

- Read 07-large-data-and-migrations.md fully
- Design specific anchors for each piece of content
- Write image prompts
