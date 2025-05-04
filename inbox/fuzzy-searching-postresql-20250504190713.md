---
title: Fuzzy searching Postresql
author: GaborZeller
date: 2025-05-04T19-07-13Z
tags:
draft: true
---

# Fuzzy searching Postresql

For fuzzy text searching in PostgreSQL, you have several good options:

1. pg_trgm (trigram) extension - Most Popular Solution:

```sql
-- Enable the extension
CREATE EXTENSION pg_trgm;

-- Create an index
CREATE INDEX idx_column_trgm ON your_table USING gin (your_column gin_trgm_ops);

-- Query example
SELECT * FROM your_table
WHERE your_column % 'search term'  -- similarity matching
   OR your_column <-> 'search term' < 0.7;  -- distance matching
```

2. fuzzystrmatch extension:

```sql
CREATE EXTENSION fuzzystrmatch;

-- Use levenshtein distance
SELECT * FROM your_table
WHERE levenshtein(your_column, 'search term') < 3;

-- Or use metaphone for phonetic matching
SELECT * FROM your_table
WHERE dmetaphone(your_column) = dmetaphone('search term');
```

3. Full-Text Search (for more complex text searching):

```sql
-- Create a tsvector column
ALTER TABLE your_table
ADD COLUMN ts_column tsvector
GENERATED ALWAYS AS (to_tsvector('english', your_column)) STORED;

-- Create an index
CREATE INDEX idx_ts ON your_table USING GIN (ts_column);

-- Query
SELECT * FROM your_table
WHERE ts_column @@ to_tsquery('english', 'search:*');
```

The pg_trgm extension is usually the best choice because:

- It's easy to use
- Performs well with proper indexing
- Handles typos and similar strings well
- Supports both similarity and distance matching
- Works with any language

Common usage pattern with pg_trgm:

```sql
-- Creating a function for convenient searching
CREATE FUNCTION fuzzy_search(search_term text) RETURNS TABLE (
    id int,
    content text,
    similarity real
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        t.id,
        t.content,
        similarity(t.content, search_term) as sim
    FROM your_table t
    WHERE t.content % search_term
    ORDER BY sim DESC
    LIMIT 10;
END;
$$ LANGUAGE plpgsql;

-- Usage
SELECT * FROM fuzzy_search('search term');
```

For performance optimization:

```sql
-- For large tables, use GiST index instead of GIN if updates are frequent
CREATE INDEX idx_column_trgm ON your_table USING gist (your_column gist_trgm_ops);

-- Set similarity threshold
SET pg_trgm.similarity_threshold = 0.3;
```

Additional considerations:

1. GIN indexes are better for static data
2. GiST indexes are better for frequently updated data
3. Consider column size and data volume when choosing index type
4. Test with your actual data to tune similarity thresholds

All these methods can be used from any programming language through standard PostgreSQL drivers.
