---
title: SQL date functions
author: GaborZeller
date: 2024-11-26T14-03-29Z
tags:
draft: true
---

# SQL date functions

## Getting current date time in SQL

- `CURRENT_DATE` - returns today's date
- `CURRENT_TIME` - returns current time without the date
- `CURRENT_TIMESTAMP` - returns current date and time

example:

```sql
SELECT
  CURRENT_DATE AS current_date,
  CURRENT_TIME AS current_time,
  CURRENT_TIMESTAMP AS current_timestamp
FROM messages
LIMIT 1;
```

| current_date        | current_time       | current_timestamp   |
| ------------------- | ------------------ | ------------------- |
| 08/27/2023 00:00:00 | 07:35:15.989933+00 | 08/27/2023 07:35:15 |

## Extracting parts from dates

```sql
SELECT
  message_id,
  sent_date,
  EXTRACT(YEAR FROM sent_date) AS extracted_year,
  DATE_PART('year', sent_date) AS part_year,

  EXTRACT(MONTH FROM sent_date) AS extracted_month,
  DATE_PART('month', sent_date) AS part_month,

  EXTRACT(DAY FROM sent_date) AS extracted_day,
  DATE_PART('day', sent_date) AS part_day,

  EXTRACT(HOUR FROM sent_date) AS extracted_hour,
  DATE_PART('hour', sent_date) AS part_hour,

  EXTRACT(MINUTE FROM sent_date) AS extracted_minute,
  DATE_PART('minute', sent_date) AS part_minute
FROM messages
LIMIT 3;
```

## Truncating dates

```sql
SELECT
  message_id,
  sent_date,
  DATE_TRUNC('month', sent_date) AS truncated_to_month,
  DATE_TRUNC('day', sent_date) AS truncated_to_day,
  DATE_TRUNC('hour', sent_date) AS truncated_to_hour
FROM messages
LIMIT 3;
```

- `trucated_to_month` - rounds down the the date of the beginning of the month
- `truncated_to_day` - rounds down the the date of the beginning of the day
- `truncated_to_hour` - rounds down the the date to the beginning of the hour

## Adding and subracting intervals

```sql
SELECT
  message_id,
  sent_date,
  sent_date + INTERVAL '2 days' AS add_2days,
  sent_date - INTERVAL '3 days' AS minus_3days,
  sent_date + INTERVAL '2 hours' AS add_2hours,
  sent_date - INTERVAL '10 minutes' AS minus_10mins
FROM messages
LIMIT 3;
```

## Formatting dates in SQL

```sql
SELECT
  message_id,
  sent_date,
  TO_CHAR(sent_date, 'YYYY-MM-DD"T"HH24:MI:SS.FF3') AS formatted_iso8601utc,
  TO_CHAR(sent_date, 'YYYY-MM-DD HH:MI:SS') AS formatted_iso8601,
  TO_CHAR(sent_date, 'YYYY-MM-DD HH:MI:SS AM') AS formatted_12hr,
  TO_CHAR(sent_date, 'Month DDth, YYYY') AS formatted_longmonth,
  TO_CHAR(sent_date, 'Mon DD, YYYY') AS formatted_shortmonth,
  TO_CHAR(sent_date, 'DD Month YYYY') AS formatted_daymonthyear,
  TO_CHAR(sent_date, 'Month') AS formatted_dayofmonth,
  TO_CHAR(sent_date, 'Day') AS formatted_dayofweek
FROM messages
LIMIT 3;
```

| format                 | output                   |
| ---------------------- | ------------------------ |
| formatted_iso8601utc   | 2022-08-03T04:43:00.000Z |
| formatted_iso8601      | 2022-08-03 04:43:00      |
| formatted_12hr         | 2022-08-03 04:43:00 PM   |
| formatted_longmonth    | August 03rd, 2022        |
| formatted_shortmonth   | Aug 03, 2022             |
| formatted_daymonthyear | 03 August 2022           |
| formatted_dayofmonth   | August                   |
| formatted_dayofweek    | Wednesday                |

## Casting strings as dates and timestamps

- `::DATE` or `TO_DATE()` - casts date as the date part only
- `::TIMESTAMP` or `TO_TIMESTAMP()` - casts date as date and time

```sql
SELECT
  sent_date,
  sent_date::DATE AS casted_date,
  TO_DATE('2023-08-27', 'YYYY-MM-DD') AS converted_to_date,
  sent_date::TIMESTAMP AS casted_timestamp,
  TO_TIMESTAMP('2023-08-27 10:30:00', 'YYYY-MM-DD HH:MI:SS') AS converted_to_timestamp
FROM messages
LIMIT 3;
```

## Generate date from components using MAKE_DATE

```
MAKE_DATE( year int, month int, day int ) → date
```

```sql
SELECT MAKE_DATE(2023,3, 25);
```
