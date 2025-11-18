---
title: SQL combine tables with UNION and UNION ALL
author: GaborZeller
date: 2024-11-27T16-22-35Z
tags: sql
draft: true
---

# SQL combine tables with UNION and UNION ALL

## Union vs union all

`Union` combines two result sets and creates a `de-duplicated` version while `union all` is like concatenate.

Example of union to combine ingredients:

```sql
select
	ingredient
from
	recipe_1
union
select
	ingredient
from
	recipe_2;
```

Example of union all to create a unique list of books:

```sql
select
	o.order_id,
	g.book_title,
	g.author,
	g.genre
from
	orders as o
inner join goodreads as g
  on
	o.book_id = g.book_id
union all
select
	o.order_id,
	g.book_title,
	g.author,
	g.genre
from
	goodreads as g
inner join orders as o
  on
	o.book_id = g.book_id
order by
	g.book_id;
```



