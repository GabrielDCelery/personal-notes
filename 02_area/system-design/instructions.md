.claude/commands/condense.md

# /condense command

Takes a verbose system design note and rewrites it as a condensed mental model reference. Reads the source file, applies the rules below, and writes a `-condensed.md` version alongside the original.

## Usage

```
/condense <filename>
```

Example: `/condense 05-queues-and-async.md` → writes `05-queues-and-async-condensed.md`

## Rules

### What to keep

- **Tables** — comparison tables, decision frameworks, trade-off matrices. These are the mental model.
- **ASCII visualizations** — latency bar charts, architecture diagrams, access pattern charts. Seeing magnitude is more useful than reading it.
- **Decision rules** — "use X when Y", "default to X", "never do X for payments"
- **Numbers you'd quote** — latency figures, throughput thresholds, sizing formulas
- **"When NOT to" sections** — negative cases are as important as positive ones
- **Scaling progressions** — ordered steps, flowcharts, decision trees

### What to cut

- **Essay introductions** — paragraphs that restate what the previous lesson covered
- **Step-by-step derivations** — if the conclusion is in the table, the derivation can go
- **Verbose examples** — keep the worked example but compress it to just the decisions, not the thinking-out-loud
- **Duplicate coverage** — if something is covered well in another lesson in the series, one line and a reference is enough
- **"Here's why this works" explanations** — keep the rule, cut the proof
- **Transition paragraphs** — "Now that we understand X, let's look at Y"

### ASCII visualization guidelines

- Use log scale for latency charts when the range spans 3+ orders of magnitude
- Use proportional bars for budget/breakdown charts (e.g. latency budget where DB = 77%)
- Don't use dots (·) as placeholder space at scale markers — they look like data points. Use spaces instead, or omit the scale grid entirely on rows with no data.
- Keep the scale markers (0 ms / 300 ms / 600 ms) on the opening and closing rule lines only

### Format guidelines

- Lead each section with the mental model or rule, not the explanation
- Bold the default choice or the most important rule in each section
- Target 150-200 lines total
- Keep the "Key Mental Models" summary section at the end — it's the cheat sheet
