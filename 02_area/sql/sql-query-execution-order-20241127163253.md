---
title: SQL query execution order
author: GaborZeller
date: 2024-11-27T16-32-53Z
tags:
draft: true
---

# SQL query execution order

| Clause       | Order | Description                                                              |
| ------------ | ----- | ------------------------------------------------------------------------ |
| FROM         | 1     | First the database tables involved in the query get identified           |
| WHERE        | 2     | The data that is involved in the query gets filtered by the WHERE clause |
| GROUP BY     | 3     | Data gets grouped by the specified columns                               |
| HAVING       | 4     | Data gets further filtered by aggregated columns                         |
| SELECT       | 5     | Columns are selected for the final data set                              |
| ORDER BY     | 6     | Records get ordered by specified order                                   |
| LIMIT/OFFSET | 7     | Limits and offsets get applied                                           |
