---
title: SQL using CASE for if conditions
author: GaborZeller
date: 2024-11-28T12-22-12Z
tags:
draft: true
---

# SQL using CASE for if conditions

The CASE expression goes through conditions and returns a value when the first condition is met.

```sql
CASE
    WHEN condition1 THEN result1
    WHEN condition2 THEN result2
    WHEN conditionN THEN resultN
    ELSE result
END;
```

## Using CASE statement in select

```sql
SELECT OrderID, Quantity,
CASE
    WHEN Quantity > 30 THEN 'The quantity is greater than 30'
    WHEN Quantity = 30 THEN 'The quantity is 30'
    ELSE 'The quantity is under 30'
END AS QuantityText
FROM OrderDetails;
```

## Using CASE statement in aggregator fields 

```sql
select
sum(case when device_type ='laptop' then 1 else 0 end) as laptop_views,
sum(case when device_type in ('tablet', 'phone') then 1 else 0 end) as mobile_views
from viewership
```




