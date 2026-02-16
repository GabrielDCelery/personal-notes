---
title: "Devops round 1, question 1"
date: 2025-11-03
tags: ["devops", "docker"]
---

# Problem

Identify the issues with the following Dockerfile

```Dockerfile
FROM trustmycompany/python:3.12
WORKDIR /app
COPY . .
RUN pip install -r requirements.txt
ENV API_KEY="E.0N_Pr0d_2025"
EXPOSE 22 3306 8080
CMD ["python", "app.py"]
```

Possible answers:

- [ ] trustmycompany/python:3.12 is not the latest
- [x] Expose has too many arguments, should use PoLP to expose less ports
- [x] Container is missing USER instructions in order not to use root
- [ ] COPY is copying the files too early which makes the image vulnerable to cache poisoning
- [x] Environment variables are hardcoded

# Solution

A list of the biggest issues with this Dockerfile.

1. Hardcoded API Key (Line 5) - CRITICAL
   - ENV API_KEY="E.0N_Pr0d_2025" embeds a production API key directly in the image
   - This key will be visible to anyone with access to the image
   - Keys should be passed at runtime via secrets or environment variables

2. Untrusted Base Image (Line 1) - HIGH
   - FROM trustmycompany/python:3.12 uses an unofficial/unverified base image
   - No way to verify the image's integrity or security
   - Should use official images like python:3.12-slim or python:3.12-alpine

3. Running as Root - HIGH
   - No USER directive means the container runs as root
   - Violates principle of least privilege
   - Should create and switch to a non-root user

4. Unnecessary Port Exposures (Line 6) - MEDIUM
   - Exposes SSH (22) and MySQL (3306) ports which are unusual for an application container
   - Port 22 suggests SSH access which shouldn't be needed in containers
   - Port 3306 suggests embedded database which is an anti-pattern

5. Copying Entire Context (Line 3) - MEDIUM
   - COPY . . copies everything including potentially sensitive files (.git, .env, etc.)
   - Should use .dockerignore or explicitly copy only needed files

6. No Image Vulnerability Scanning - MEDIUM
   - No health checks or security scanning in the build process
   - Dependencies installed without version pinning or vulnerability checks

## Why I did not choose #1 as a valid answer

Out of all the options the first one does not make sense because the issue is not the version of the image but the above mentioned issues. Plus nobody would recommend to use `latest` for building an image.

## Why I did not choose #4 as a valid answer

Copying the files "too early" is not what poisoning the cache means.

Cache poisoning happens when an attacker tricks Docker into caching a malicious layer that gets reused in future builds - even when you think you're building clean code.

Example:

```Dockerfile
FROM ubuntu:20.04
RUN apt-get update && apt-get install -y curl
RUN curl https://example.com/setup.sh | bash
COPY . /app
```

The Poisoning Attack:

1. Attacker's Initial Build
   - Attacker compromises example.com/setup.sh temporarily
   - The script contains: echo "malicious backdoor" > /usr/bin/backdoor
   - docker build -t myapp:latest .
   - Docker caches this layer with the malicious script

2. Attacker Restores Clean Script
   - Attacker puts the legitimate setup.sh back on example.com
   - Now it looks normal if anyone checks!

3. Victim Rebuilds
   - Days later, you rebuild the same Dockerfile
   - docker build -t myapp:latest .
