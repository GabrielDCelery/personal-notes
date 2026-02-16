---
title: Fix go database migration
---

# The problem

When working with go migrations encountered this issue, this means the database got stuck in an inconsistent state.

```sh
error: Dirty database version 1. Fix and force version.
[migrate-up] ERROR task failed
```

# Solution

Either try to go back to the previous working version

```sh
migrate -path migrations -database "$DATABASE_URL" force 0
```

Or try to roll "forward" after fixing the query

```sh
migrate -path migrations -database "$DATABASE_URL" force 1
```

Continue/finish the migration

```sh
migrate -path migrations -database "$DATABASE_URL" up
```

Or go nucler

> [!WARNING]
> Yeah try not to do this, especially in prod, could be viable in dev

```sh
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" down
```
