---
title: Structure python project
author: GaborZeller
date: 2025-04-23T06-30-44Z
tags:
draft: true
---

# Structure python project

```
my_project/
├── Dockerfile.script1
├── Dockerfile.script2
├── pyproject.toml
├── poetry.lock
├── README.md
├── src/
│   ├── my_project/
│   │   ├── __init__.py
│   │   ├── common/
│   │   │   ├── __init__.py
│   │   │   ├── utils.py
│   │   │   └── config.py
│   │   └── scripts/
│   │       ├── __init__.py
│   │       ├── script1.py
│   │       └── script2.py
└── tests/
    ├── __init__.py
    ├── test_script1.py
    └── test_script2.py
```

Here's how to implement each component:

1. First, the `pyproject.toml`:

```toml
[tool.poetry]
name = "my-project"
version = "0.1.0"
description = "Project with multiple entry points"
authors = ["Your Name <your.email@example.com>"]
packages = [
    { include = "my_project", from = "src" }
]

[tool.poetry.dependencies]
python = "^3.9"
# Add your dependencies here

[tool.poetry.dev-dependencies]
pytest = "^7.0"

[tool.poetry.scripts]
script1 = "my_project.scripts.script1:main"
script2 = "my_project.scripts.script2:main"

[build-system]
requires = ["poetry-core>=1.0.0"]
build-backend = "poetry.core.masonry.api"
```

2. Example shared utility code (`src/my_project/common/utils.py`):

```python
def shared_function():
    return "I'm shared between scripts!"
```

3. Example script files:

`src/my_project/scripts/script1.py`:

```python
from my_project.common.utils import shared_function

def main():
    print("Running Script 1")
    print(shared_function())

if __name__ == "__main__":
    main()
```

`src/my_project/scripts/script2.py`:

```python
from my_project.common.utils import shared_function

def main():
    print("Running Script 2")
    print(shared_function())

if __name__ == "__main__":
    main()
```

4. Dockerfile for script1 (`Dockerfile.script1`):

```dockerfile
# Use multi-stage build for smaller final image
FROM python:3.9-slim as builder

# Install poetry
RUN pip install poetry

# Copy project files
WORKDIR /app
COPY pyproject.toml poetry.lock ./
COPY src ./src

# Install dependencies
RUN poetry config virtualenvs.create false \
    && poetry install --no-dev --no-interaction --no-ansi

# Final stage
FROM python:3.9-slim

COPY --from=builder /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages
COPY --from=builder /app/src /app/src

# Set Python path
ENV PYTHONPATH=/app/src

# Run script1
CMD ["python", "-m", "my_project.scripts.script1"]
```

5. Dockerfile for script2 (`Dockerfile.script2`):

```dockerfile
# Similar to Dockerfile.script1 but with different CMD
FROM python:3.9-slim as builder

RUN pip install poetry

WORKDIR /app
COPY pyproject.toml poetry.lock ./
COPY src ./src

RUN poetry config virtualenvs.create false \
    && poetry install --no-dev --no-interaction --no-ansi

FROM python:3.9-slim

COPY --from=builder /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages
COPY --from=builder /app/src /app/src

ENV PYTHONPATH=/app/src

CMD ["python", "-m", "my_project.scripts.script2"]
```

6. To build and run the Docker containers:

```bash
# Build containers
docker build -f Dockerfile.script1 -t my-project-script1 .
docker build -f Dockerfile.script2 -t my-project-script2 .

# Run containers
docker run my-project-script1
docker run my-project-script2
```

7. To run scripts locally using Poetry:

```bash
# Install dependencies
poetry install

# Run scripts
poetry run script1
poetry run script2
```

This structure provides several benefits:

1. Shared code is in the `common` directory and can be imported by any script
2. Each script has its own entry point and can be run independently
3. Scripts can be run either locally using Poetry or in Docker containers
4. The Docker containers are isolated but share the same codebase
5. Using multi-stage builds keeps the Docker images small
6. The project structure follows Python best practices
7. Testing can be done for each script independently

You can also add a `docker-compose.yml` file if you need to run multiple containers together or need to define environment variables and other container configurations:

```yaml
version: "3"

services:
  script1:
    build:
      context: .
      dockerfile: Dockerfile.script1
    environment:
      - SOME_ENV_VAR=value

  script2:
    build:
      context: .
      dockerfile: Dockerfile.script2
    environment:
      - SOME_ENV_VAR=value
```

This structure allows for easy development, testing, and deployment while maintaining code reusability and container isolation.

## How to install dependencies separately for scripts

Yes, in Poetry (which your project is using based on the `pyproject.toml`), you can define optional dependency groups using the `[tool.poetry.group]` section. This allows you to organize dependencies that are only needed for specific scripts or purposes.

Here's how you can modify your `pyproject.toml` to separate dependencies for your different scripts:

```toml:pyproject.toml
[project]
name = "flesh_and_blood"
version = "0.1.0"
description = ""
authors = [
    {name = "GaborZeller",email = "gabriel.d.celery@gmail.com"}
]
packages = [
    { include = "flesh_and_blood", from = "src" }
]
readme = "README.md"
requires-python = ">=3.13"
dependencies = [
    # Core dependencies that are always needed go here
]

[tool.poetry.group.scraping]
optional = true
dependencies = [
    # Dependencies specific to image scraping go here
    # For example:
    # "requests",
    # "beautifulsoup4",
]

[tool.poetry.group.training]
optional = true
dependencies = [
    # Dependencies specific to model training go here
    # For example:
    # "torch",
    # "numpy",
    # "pandas",
]

[tool.poetry.scripts]
scrape-images = "flesh_and_blood.scripts.scrape-images:main"
train-model = "flesh_and_blood.scripts.train-model:main"

[build-system]
requires = ["poetry-core>=2.0.0,<3.0.0"]
build-backend = "poetry.core.masonry.api"
```

With this setup, you can:

1. Install only core dependencies:

```bash
poetry install --without scraping --without training
```

2. Install with scraping dependencies:

```bash
poetry install --with scraping
```

3. Install with training dependencies:

```bash
poetry install --with training
```

4. Install everything:

```bash
poetry install --with scraping --with training
```

This approach has several benefits:

- Keeps your environment lighter when you only need certain functionality
- Makes it clearer which dependencies are needed for which features
- Allows other developers to quickly understand the project structure
- Reduces potential conflicts between dependencies that are only needed for specific tasks

You can add more groups as needed, and each group can have its own set of dependencies that are only installed when explicitly requested.

**Generation complete!** Please review the code suggestions above.
