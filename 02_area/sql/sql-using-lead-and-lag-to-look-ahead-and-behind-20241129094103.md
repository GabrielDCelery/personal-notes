---
title: SQL using LEAD and LAG to look ahead and behind
author: GaborZeller
date: 2024-11-29T09-41-03Z
tags: sql
draft: true
---

# SQL using LEAD and LAG to look ahead and behind

## How does LEAD and LAG work?

These are window functions that are desinged to analyze time series data. The goal is to access data that either came before or after the current row.

`LEAD` - looks into the future
`LAG` - looks into the past

```sql
LEAD(column_name, offset) OVER (  -- Compulsory expression
  PARTITION BY partition_column -- Optional expression
  ORDER BY order_column) -- Compulsory expression

LAG(column_name, offset) OVER ( -- Compulsory expression
  PARTITION BY partition_column -- Optional expression
  ORDER BY order_column) -- Compulsory expression
```

Example of how you would use it to be able to calculate closing prices between consecutive months.

```sql
SELECT
  date,
  close,
  LEAD(close) OVER (ORDER BY date) AS next_month_close,
  LAG(close) OVER (ORDER BY date) AS prev_month_close
FROM stock_prices
WHERE EXTRACT(YEAR FROM date) = 2023
  AND ticker = 'GOOG';
```
