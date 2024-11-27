---
title: SQL window function to analyze partitions of data
author: GaborZeller
date: 2024-11-27T15-48-26Z
tags:
draft: true
---

# SQL window function to analyze partitions of data

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


