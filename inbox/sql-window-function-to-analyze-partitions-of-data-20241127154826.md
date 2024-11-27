---
title: SQL window function to analyze partitions of data
author: GaborZeller
date: 2024-11-27T15-48-26Z
tags: sql
draft: true
---

# SQL window function to analyze partitions of data

## Examples of using SQL window

Example of getting the running total of products partitioned by product.

```sql
SELECT
	spend,
	SUM(spend) over (PARTITION BY product ORDER BY transaction_date) as running_total
FROM
	product_spend;
```
Example of geting product count partitioned by category AND product.

```sql
SELECT
  category,
  product,
  COUNT(*) OVER (
    PARTITION BY category, product) AS product_count
FROM product_spend;
```

## What are window functions used for

Window functions allow to crete `virtual windows` within a dataset. This allows us to create `aggregated analysis` over sub-sections of our data set.

## How to limit the number of records within the window

When using window function each row is aggregated in a cumulative manner, but what if you do not want to aggregate all the rows that came before the row you are currently evaluating?

In this case use the `rows between` syntax to determine how many records are used in the aggregate.

| Parameters | Description |
| ---------- | ----------- |
| rows between Unbounded Preceding and Current Row | includes the current row and all rows before current row |
| rows between N Preceding and Current Row | includes the current row and a specified number of rows before current row |
| rows between current and current row | includes only the current row |
| rows between N Preceding and M Following | includes specified rows before and after the current row |
| rows between Current row and M Following | includes the current row and a specified number of rows after current row |
| rows between Current Row and Unbounded Following | includes the current row and all rows after current row |








