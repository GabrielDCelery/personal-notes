---
title: SQL ranking tables using row number, rank and dense rank
author: GaborZeller
date: 2024-11-26T21-40-34Z
tags:
draft: true
---

# SQL ranking tables using row number, rank and dense rank

## Row number

- assigns a unique sequential number to each row within a partition
- no gaps in numbering even for tied values

```sql
SELECT
	user_id,
    spend,
    transaction_date,
	ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY transaction_date) as row_num
FROM transactions
```

## Rank

- assigns a rank based on the ORDER BY clause
- handles ties by assigning same rank but skips next rank
- may have gaps in ranking if there are multiple ties

```sql
SELECT
	user_id,
    spend,
    transaction_date,
	RANK() OVER (PARTITION BY user_id ORDER BY transaction_date) as rank
FROM transactions
```

## Dense Rank

- similar to RANK() but does not skip ranks for ties
- continues consecutive ranking even with tied values
- no gaps in ranking for any number of ties

```sql
SELECT
	user_id,
    spend,
    transaction_date,
	DENSE_RANK() OVER (PARTITION BY user_id ORDER BY transaction_date) as dense_rank
FROM transactions
```
