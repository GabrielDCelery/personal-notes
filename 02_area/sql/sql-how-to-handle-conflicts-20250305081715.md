---
title: SQL how to handle conflicts
author: GaborZeller
date: 2025-03-05T08-17-15Z
tags:
draft: true
---

# SQL how to handle conflicts

## Do nothing

```sql
INSERT INTO sometable.speed_events_osm (somefield,) VALUES ($1) ON CONFLICT DO NOTHING`;
```
