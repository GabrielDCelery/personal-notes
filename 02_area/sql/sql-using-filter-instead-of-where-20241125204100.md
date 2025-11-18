---
title: SQL using filter instead of where
author: GaborZeller
date: 2024-11-25T20-41-00Z
tags:
draft: true
---

# SQL using filter instead of where

```sql
select
count(*) filter(where device_type='laptop') as laptop_views,
count(*) filter(where device_type in ('tablet', 'phone')) as mobie_views
from viewership;
```

```sql
select
sum(case when device_type ='laptop' then 1 else 0 end) as laptop_views,
sum(case when device_type in ('tablet', 'phone') then 1 else 0 end) as mobile_views
from viewership
```
