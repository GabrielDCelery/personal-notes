---
title: Common URI schema examples and patterns
author: GaborZeller
date: 2025-07-02T14-47-59Z
tags:
draft: true
---

# Common URI schema examples and patterns

## The structure of an URI

The point of an URI schema is to standardize the location/identification of a resource accross different protocols:

```sh
scheme://[authority]/[path][?query][#fragment]
```

- Scheme: The scheme is the first part of the URI and indicates the protocol to be used (e.g., http, https, ftp, mailto, file)
- Authority: The authority component, preceded by `//`, contains information about the server or resource own
- Path: The path component specifies the location of the resource on the server or within the system
- Query: The query component, preceded by a ?, provides additional parameters or data to be passed to the resource
- Fragment: The fragment component, preceded by a #, identifies a specific part or section within the resource

## Environment Variables

```sh
# Examples
env://DATABASE_URL
env://API_KEY
env://AWS_ACCESS_KEY
```

These reference environment variables stored in the system.

For example, in: `env://DATABASE_URL`

| Component | Value        | Description                                   |
| --------- | ------------ | --------------------------------------------- |
| Scheme    | env          | Protocol used                                 |
| Authority | none         | No authority section after //                 |
| Path      | DATABASE_URL | Name of the environment variable to reference |
| Query     | none         | No query parameters                           |
| Fragment  | none         | No fragment section                           |

## File System

```sh
# Examples
file:///home/user/document.txt
file://localhost/c:/windows/system32
file://./config.json
```

These point to local file system locations.

For example, in `file:///home/user/document.txt`

| Component | Value                   | Description                                               |
| --------- | ----------------------- | --------------------------------------------------------- |
| Scheme    | file                    | The protocol identifier for local filesystem access       |
| Authority | (empty)                 | The triple forward slashes (///) indicate empty authority |
| Path      | /home/user/document.txt | The absolute path to the file on the local filesystem     |
| Query     | (none)                  | No query parameters present                               |
| Fragment  | (none)                  | No fragment identifier present                            |

## HTTP/HTTPS

```sh
# Examples
http://www.example.com
https://api.service.com/v1
https://localhost:8080
```

For web resources over HTTP/HTTPS protocols.

For exmple, in `https://www.example.com:8080/products/123?sort=price#details`

| Component | Value                | Description                            |
| --------- | -------------------- | -------------------------------------- |
| Scheme    | https                | Protocol used                          |
| Authority | www.example.com:8080 | Server address (including port number) |
| Path      | /products/123        | Resource location                      |
| Query     | sort=price           | URL parameters                         |
| Fragment  | details              | Section reference                      |

## Database

```sh
# Examples
postgresql://user:password@localhost:5432/dbname
mysql://user:password@hostname:3306/database
mongodb://localhost:27017/dbname
```

Database connection strings.

For example, in `postgresql://user:password@localhost:5432/dbname`

| Component | Value                        | Description                                   |
| --------- | ---------------------------- | --------------------------------------------- |
| Scheme    | postgresql                   | Database protocol used                        |
| Authority | user:password@localhost:5432 | Authentication, hostname and port information |
| Path      | /dbname                      | Database name                                 |
| Query     | none                         | No query parameters                           |
| Fragment  | none                         | No fragment identifier                        |

## Amazon S3

```sh
# Examples
s3://bucket-name/path/to/file
s3://my-bucket/images/photo.jpg
```

AWS S3 bucket locations.

For example, in `s3://bucket-name/path/to/file`

| Component | Value         | Description                          |
| --------- | ------------- | ------------------------------------ |
| Scheme    | s3            | Amazon S3 protocol identifier        |
| Authority | bucket-name   | The S3 bucket name                   |
| Path      | /path/to/file | Path to the object within the bucket |
| Query     | none          | No query parameters                  |
| Fragment  | none          | No fragment identifier               |

## Redis

```sh
redis://username:password@hostname:6379
redis://localhost:6379
```

Redis connection strings.

For example, in `redis://username:password@hostname:6379`

| Component | Value                           | Description                                   |
| --------- | ------------------------------- | --------------------------------------------- |
| Scheme    | redis                           | Redis protocol identifier                     |
| Authority | username:password@hostname:6379 | Authentication credentials, hostname and port |
| Path      | none                            | No path specified                             |
| Query     | none                            | No query parameters                           |
| Fragment  | none                            | No fragment identifier                        |

## FTP

```sh
ftp://user:password@host:port/path
ftp://ftp.example.com/public
```

File Transfer Protocol locations.

For example, in `ftp://user:password@host:port/path`

| Component | Value                   | Description                                   |
| --------- | ----------------------- | --------------------------------------------- |
| Scheme    | ftp                     | File Transfer Protocol identifier             |
| Authority | user:password@host:port | Authentication credentials, hostname and port |
| Path      | /path                   | Path to the resource on the FTP server        |
| Query     | none                    | No query parameters                           |
| Fragment  | none                    | No fragment identifier                        |

## Data URIs

```sh
data:text/plain;base64,SGVsbG8gV29ybGQ=
data:image/jpeg;base64,/9j/4AAQSkZJRg...
```

Note: The actual data content "SGVsbG8gV29ybGQ=" follows the comma and is the base64-encoded payload.

Inline data encoding.

For example, in `data:text/plain;base64,SGVsbG8gV29ybGQ=`

| Component | Value             | Description                               |
| --------- | ----------------- | ----------------------------------------- |
| Scheme    | data              | Data URI scheme identifier                |
| Authority | none              | No authority section needed for data URIs |
| Path      | text/plain;base64 | MIME type and encoding method             |
| Query     | none              | No query parameters                       |
| Fragment  | none              | No fragment identifier                    |
