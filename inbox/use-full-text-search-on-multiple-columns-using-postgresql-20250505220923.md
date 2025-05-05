---
title: Use full text search on multiple columns using Postgresql
author: GaborZeller
date: 2025-05-05T22-09-23Z
tags:
draft: true
---

# Use full text search on multiple columns using Postgresql

> how to combine two columns of postgres table for weighted full text search where I want to prioritise the results from one column over an othrer?
> To combine two columns for weighted full text search in PostgreSQL with different priorities, you can use the `setweight()` function and combine the columns using `||`. Here are a few approaches:

1. **Using a Generated Column or View:**

```sql
-- Create a generated tsvector column
ALTER TABLE your_table ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(title,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(description,'')), 'B')
) STORED;

-- Create an index
CREATE INDEX search_vector_idx ON your_table USING gin(search_vector);

-- Search query
SELECT *
FROM your_table
WHERE search_vector @@ to_tsquery('english', 'your_search_term')
ORDER BY ts_rank(search_vector, to_tsquery('english', 'your_search_term')) DESC;
```

2. **Direct in Query:**

```sql
SELECT *,
    ts_rank(
        setweight(to_tsvector('english', title), 'A') ||
        setweight(to_tsvector('english', description), 'B'),
        to_tsquery('english', 'your_search_term')
    ) as rank
FROM your_table
WHERE
    setweight(to_tsvector('english', title), 'A') ||
    setweight(to_tsvector('english', description), 'B') @@
    to_tsquery('english', 'your_search_term')
ORDER BY rank DESC;
```

3. **With Custom Weights:**

```sql
SELECT *,
    ts_rank(
        setweight(to_tsvector('english', title), 'A') ||
        setweight(to_tsvector('english', description), 'B'),
        to_tsquery('english', 'your_search_term'),
        '{0.1, 0.2, 0.4, 1.0}'  -- weights for D,C,B,A
    ) as rank
FROM your_table
WHERE
    setweight(to_tsvector('english', title), 'A') ||
    setweight(to_tsvector('english', description), 'B') @@
    to_tsquery('english', 'your_search_term')
ORDER BY rank DESC;
```

The weights go from 'A' (highest) to 'D' (lowest). In these examples:

- 'A' weight is given to the title (higher priority)
- 'B' weight is given to the description (lower priority)

Key points:

- `setweight()` assigns weights to terms
- Weights are: A > B > C > D
- Default weights are: A=1, B=0.4, C=0.2, D=0.1
- You can customize weights in `ts_rank()`
- Use `coalesce()` to handle NULL values
- Creating an index will improve performance

If you're doing this frequently, the first approach (generated column) will be more efficient as it pre-calculates the tsvector and allows for indexing.

Example with a specific search:

```sql
-- Create table
CREATE TABLE articles (
    id serial primary key,
    title text,
    description text
);

-- Add search vector
ALTER TABLE articles ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(title,'')), 'A') ||
    setweight(to_tsvector('english', coalesce(description,'')), 'B')
) STORED;

-- Create index
CREATE INDEX articles_search_idx ON articles USING gin(search_vector);

-- Search
SELECT title, description,
    ts_rank(search_vector, to_tsquery('english', 'database')) as rank
FROM articles
WHERE search_vector @@ to_tsquery('english', 'database')
ORDER BY rank DESC;
```
