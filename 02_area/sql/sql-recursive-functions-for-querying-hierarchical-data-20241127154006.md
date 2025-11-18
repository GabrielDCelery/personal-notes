---
title: SQL recursive functions for querying hierarchical data
author: GaborZeller
date: 2024-11-27T15-40-06Z
tags: sql
draft: true
---

# SQL recursive functions for querying hierarchical data

Example of selecting all of the subordinates of a manager.

```sql
WITH recursive_cte AS (
  SELECT 
    employee_id, 
    name, 
    manager_id
  FROM employees
  WHERE manager_id = @manager_id
  
  UNION ALL
  
  SELECT 
    e.employee_id, 
    e.name, 
    e.manager_id
  FROM employees AS e
  INNER JOIN recursive_cte AS r -- The RECURSIVE CTE is utilized here within the main CTE.
    ON e.manager_id = r.employee_id
)

SELECT * 
FROM recursive_cte;
```

## How does a recursive SQL query work?

A recursive query starts from a base query that returns some value and the output of that is fed into the recursive function to start a loop of querying records until nothing is returned and the loop ends.

```
WITH R AS (
	<base query>
	UNION ALL
	<recursive query involving R>
)
<query involving R>
```
