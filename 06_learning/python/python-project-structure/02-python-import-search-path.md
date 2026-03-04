# Python Import Search Path

## Question
How does Python know where to do the absolute import from?

## Python's Import Search Path

When you do `from youtube_notetaker.video import download_video`, Python searches in this order:

1. **Built-in modules** (like `sys`, `os`, `json`)
2. **Directories in `sys.path`** (a list you can inspect)

You can check your search path with:
```bash
python -c "import sys; print('\n'.join(sys.path))"
```

## What's in `sys.path`

1. **Current directory** (usually first)
2. **PYTHONPATH environment variable** (if set)
3. **Standard library** (e.g., `/usr/lib/python3.14`)
4. **site-packages** (installed packages like `ollama`, `cv2`)

## How Python Finds Your Package

### Method 1: Running from Project Root (Current Approach)

When you run `python main.py` from `/home/gaze/projects/github-GabrielDCelery/youtube-ai-notetaker/`:
- Python automatically adds the **current directory** to `sys.path`
- So it finds `downloader.py` because it's right there

### Method 2: Installing Your Package (For Absolute Imports)

If you structured it as:
```
youtube-ai-notetaker/
├── pyproject.toml
└── src/
    └── youtube_notetaker/
        ├── __init__.py
        ├── video/
        │   └── downloader.py
        └── cli.py
```

You'd install it in "editable mode":
```bash
pip install -e .
```

This adds `src/` to Python's search path, so `from youtube_notetaker.video import download_video` works **from anywhere**.

### Method 3: Manual Path Manipulation (Not Recommended)

You could manually add to `sys.path`:
```python
import sys
from pathlib import Path

# Add project root to path
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))

# Now imports work
from youtube_notetaker.video import download_video
```

## Standard Package Layout

To enable absolute imports, restructure to:

```
youtube-ai-notetaker/
├── pyproject.toml
├── src/
│   └── youtube_notetaker/    # This becomes importable
│       ├── __init__.py
│       ├── cli.py
│       └── downloader.py
└── README.md
```

Update `pyproject.toml`:
```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "youtube-ai-notetaker"
# ... rest of your config

[tool.hatch.build.targets.wheel]
packages = ["src/youtube_notetaker"]
```

Then run:
```bash
uv pip install -e .
```

Now `from youtube_notetaker.downloader import get_video_id` works from anywhere!

## The Key Insight

**Python finds packages by:**
1. Checking if the top-level name (`youtube_notetaker`) is a directory with `__init__.py`
2. Searching through `sys.path` to find that directory
3. Installing your package adds its location to `sys.path` permanently

**Why this matters:**
- **Without installation**: imports only work when running from project root
- **With installation**: imports work from anywhere, tests can import your code, other projects can use it

The `src/` layout is a modern best practice because it forces you to install the package (even in editable mode), ensuring imports work consistently.
