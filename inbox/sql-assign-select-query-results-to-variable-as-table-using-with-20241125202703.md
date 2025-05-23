---
title: SQL assign select query results to variable as table using with
author: GaborZeller
date: 2024-11-25T20-27-03Z
tags:
  - sql
draft: true
---

# SQL assign select query results to variable as table using with

You can use the `with` statement to create temporary tables that can be referenced accross the query.

```sql
with
laptop_views as (select count(*) from viewership  where device_type='laptop'),
mobile_views as (select count(*) from viewership  where device_type in ('tablet', 'phone'))

select
	laptop_views.count as laptop_views,
	mobile_views.count as mobile_views
from
	laptop_views
cross join mobile_views
```

```sql
WITH laptop_views AS
  (SELECT count(*)
   FROM viewership
   WHERE device_type='laptop'),
     mobile_views AS
  (SELECT count(*)
   FROM viewership
   WHERE device_type IN ('tablet',
                         'phone'))
SELECT laptop_views.count AS laptop_views,
       mobile_views.count AS mobile_views
FROM laptop_views
CROSS JOIN mobile_views
```
