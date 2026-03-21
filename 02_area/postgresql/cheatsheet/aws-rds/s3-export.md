# Exporting Data to S3

RDS does not give you access to the server filesystem, so `COPY TO '/tmp/file'` does not work. You have two options: the `aws_s3` extension (server-side, data goes direct to S3) or `\copy` (client-side, data flows through your machine).

## Option 1: aws_s3 Extension (Recommended)

Data goes directly from RDS to S3 without touching your machine. Requires an IAM role attached to the RDS instance.

### IAM setup

Create an IAM role with S3 write access and attach it to the RDS instance:

```sh
# Attach an IAM role to the instance (role must trust rds.amazonaws.com)
aws rds add-role-to-db-instance \
  --db-instance-identifier mydb \
  --role-arn arn:aws:iam::111111111111:role/rds-s3-export-role \
  --feature-name s3Export
```

The role needs this policy:

```json
{
  "Effect": "Allow",
  "Action": ["s3:PutObject", "s3:AbortMultipartUpload"],
  "Resource": "arn:aws:s3:::my-bucket/archive/*"
}
```

### Enable the extension

```sql
CREATE EXTENSION aws_s3 CASCADE;
```

### Export a table or query result

```sql
-- Export a full partition
SELECT aws_s3.query_export_to_s3(
    'SELECT * FROM orders_2023',
    aws_commons.create_s3_uri(
        'my-bucket',
        'archive/orders/2023/data.csv',
        'eu-west-1'
    ),
    options := 'FORMAT csv, HEADER true'
);

-- Export with a WHERE clause (no partition required)
SELECT aws_s3.query_export_to_s3(
    'SELECT * FROM orders WHERE created_at < ''2023-01-01''',
    aws_commons.create_s3_uri(
        'my-bucket',
        'archive/orders/pre-2023/data.csv',
        'eu-west-1'
    ),
    options := 'FORMAT csv, HEADER true'
);
```

The function returns a row with `rows_uploaded`, `files_uploaded`, and `bytes_uploaded` — use this to verify the export before dropping data.

```sql
-- Capture and verify
SELECT rows_uploaded, files_uploaded, bytes_uploaded
FROM aws_s3.query_export_to_s3(
    'SELECT * FROM orders_2023',
    aws_commons.create_s3_uri('my-bucket', 'archive/orders/2023/data.csv', 'eu-west-1'),
    options := 'FORMAT csv, HEADER true'
);

-- Cross-check row count
SELECT COUNT(*) FROM orders_2023;
-- rows_uploaded should match
```

## Option 2: \copy via psql (No IAM Setup Required)

`\copy` is a psql client command — it runs on your machine and streams data over the existing database connection. No server filesystem access needed.

```sh
# Export to a local file
psql -h mydb.xxxxx.eu-west-1.rds.amazonaws.com -U myuser mydb \
  -c "\copy orders_2023 TO '/tmp/orders_2023.csv' (FORMAT csv, HEADER true)"

# Pipe directly to S3 (no local file needed)
psql -h mydb.xxxxx.eu-west-1.rds.amazonaws.com -U myuser mydb \
  -c "\copy orders_2023 TO STDOUT (FORMAT csv, HEADER true)" \
  | aws s3 cp - s3://my-bucket/archive/orders/2023/data.csv
```

Run this from an EC2 instance in the same region and AZ as the RDS instance — if you run it from your laptop the data has to travel over the internet, which is slow for large tables.

## Converting CSV to Parquet

RDS cannot write Parquet directly. Convert after the export. The easiest options:

### DuckDB (fastest, no dependencies)

```sh
# Install: https://duckdb.org/docs/installation
duckdb -c "
COPY (SELECT * FROM read_csv_auto('s3://my-bucket/archive/orders/2023/data.csv'))
TO 's3://my-bucket/archive/orders/2023/data.parquet'
(FORMAT parquet, COMPRESSION snappy);
"
```

DuckDB can read and write S3 directly. No intermediate local file needed.

### Python + pandas (simpler for one-off scripts)

```python
import pandas as pd

df = pd.read_csv('/tmp/orders_2023.csv')
df.to_parquet('/tmp/orders_2023.parquet', compression='snappy', index=False)

# Then upload
import boto3
boto3.client('s3').upload_file(
    '/tmp/orders_2023.parquet',
    'my-bucket',
    'archive/orders/2023/data.parquet'
)
```

## Querying Archived Data with Athena

Once the Parquet file is in S3, register it as an Athena table (one-time setup):

```sql
-- Run in the Athena query editor
CREATE EXTERNAL TABLE orders_archive (
    id          BIGINT,
    created_at  TIMESTAMP,
    customer_id BIGINT,
    total       DOUBLE
)
PARTITIONED BY (year STRING)
STORED AS PARQUET
LOCATION 's3://my-bucket/archive/orders/'
TBLPROPERTIES ('parquet.compress'='SNAPPY');

-- Register the partition
ALTER TABLE orders_archive ADD PARTITION (year='2023')
LOCATION 's3://my-bucket/archive/orders/2023/';
```

Then query it like any table — Athena charges $5 per TB scanned:

```sql
SELECT customer_id, SUM(total)
FROM orders_archive
WHERE year = '2023'
GROUP BY customer_id;
```

Organising files by year (or year/month) in S3 and using Athena partitions means queries only scan the relevant files, keeping costs low.

## Full Archival Workflow

```
1. Export partition via aws_s3.query_export_to_s3() → S3 (CSV)
2. Verify rows_uploaded matches COUNT(*) on the partition
3. Convert CSV → Parquet (DuckDB or Lambda)
4. Register Parquet in Athena
5. Detach partition: ALTER TABLE orders DETACH PARTITION orders_2023 CONCURRENTLY
6. Keep detached table for a few days as a safety net
7. DROP TABLE orders_2023
```

See `scaling/partitioning.md` for the detach-drop sequence.
