# How Hatchling and project.scripts Work Together

## Question
I am confused - how does hatchling play together with project.scripts?

## The Relationship

**`project.scripts`** declares **WHAT** command-line tools your package provides:
```toml
[project.scripts]
youtube-notetaker = "youtube_notetaker.cli:main"
```
This says: "Create a command called `youtube-notetaker` that runs the `main()` function from `youtube_notetaker.cli`"

**`hatchling`** (the build backend) **IMPLEMENTS** how to create those tools:
```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```
This says: "Use hatchling to build the package and create the entry point scripts"

## How They Work Together

Let me trace through what happens when you install:

### 1. You Run Install
```bash
uv pip install -e .
```

### 2. `uv` Reads `pyproject.toml`
```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "youtube-ai-notetaker"

[project.scripts]
youtube-notetaker = "youtube_notetaker.cli:main"
```

### 3. `uv` Calls Hatchling
"Hey hatchling, build this package and install it"

### 4. Hatchling Reads `[project.scripts]`
"Oh, they want a CLI command called `youtube-notetaker`"

### 5. Hatchling Creates the Script
It generates a file at `.venv/bin/youtube-notetaker`:

```python
#!/path/to/.venv/bin/python
# -*- coding: utf-8 -*-
import re
import sys
from youtube_notetaker.cli import main

if __name__ == '__main__':
    sys.argv[0] = re.sub(r'(-script\.pyw|\.exe)?$', '', sys.argv[0])
    sys.exit(main())
```

### 6. Makes It Executable
```bash
chmod +x .venv/bin/youtube-notetaker
```

### 7. Now You Can Run It
```bash
youtube-notetaker https://youtube.com/watch?v=...
```

## The Standard: `[project.scripts]` is Universal

Here's the key insight: **`[project.scripts]` is a standard** (defined in PEP 621).

**ALL build backends understand it:**

```toml
# With Hatchling
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project.scripts]
youtube-notetaker = "youtube_notetaker.cli:main"
# ✅ Hatchling knows what to do
```

```toml
# With Setuptools
[build-system]
requires = ["setuptools>=61.0"]
build-backend = "setuptools.build_meta"

[project.scripts]
youtube-notetaker = "youtube_notetaker.cli:main"
# ✅ Setuptools knows what to do
```

```toml
# With Flit
[build-system]
requires = ["flit_core"]
build-backend = "flit_core.buildapi"

[project.scripts]
youtube-notetaker = "youtube_notetaker.cli:main"
# ✅ Flit knows what to do
```

They all create the same type of executable script!

## What Gets Created in the Wheel

When hatchling builds a wheel file, it includes metadata about entry points:

```bash
# Build creates:
youtube_ai_notetaker-0.1.0-py3-none-any.whl

# Inside the wheel, there's a file:
youtube_ai_notetaker-0.1.0.dist-info/entry_points.txt

# Contents:
[console_scripts]
youtube-notetaker = youtube_notetaker.cli:main
```

When `pip`/`uv` installs the wheel, they read this file and create the executable scripts.

## Multiple Entry Points Example

You can define multiple commands:

```toml
[project.scripts]
youtube-notetaker = "youtube_notetaker.cli:main"
ytnotetaker = "youtube_notetaker.cli:main"           # Short alias
yt-download = "youtube_notetaker.cli:download_only"  # Different command
```

Hatchling creates three executables:
- `.venv/bin/youtube-notetaker`
- `.venv/bin/ytnotetaker`
- `.venv/bin/yt-download`

## Division of Responsibilities

**You (the developer) declare in `[project.scripts]`:**
- What commands should exist
- What Python functions they should call

**Hatchling (the build backend) implements:**
- How to create the executable wrapper scripts
- How to package everything into a wheel
- How to generate the metadata

**pip/uv (the installer) uses:**
- The wheel file hatchling created
- The entry_points.txt metadata
- Creates the actual executables in your environment

## Summary

- **`[project.scripts]`** = Declaration of what CLI commands you want
- **Hatchling** = Implementation that creates those commands
- **Standard** = All build backends understand `[project.scripts]`
- **Result** = Executable scripts in `.venv/bin/` that call your Python functions

It's like:
- **`[project.scripts]`** = The blueprint/recipe
- **Hatchling** = The builder/chef
- **Executable script** = The finished product
