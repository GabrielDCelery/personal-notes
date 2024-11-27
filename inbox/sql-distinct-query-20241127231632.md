---
title: SQL distinct query
author: GaborZeller
date: 2024-11-27T23-16-32Z
tags: sql
draft: true
---

# SQL distinct query

Distinct is used to only return different values in an SQL query.

## Distinct values by single field

```sql
SELECT DISTINCT manufacturer
FROM pharmacy_sales;
```

## Distinct values by multiple fields

```sql
SELECT DISTINCT user_id, status
FROM trades
ORDER BY user_id;
```

## Distinct values in aggregator functions

```sql
SELECT COUNT(DISTINCT user_id) 
FROM trades;
```
