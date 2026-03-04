# Editable Install Explained

## Question
What does `uv pip install -e .` do?

## Breaking Down the Command

**`uv pip`** - Using `uv`'s pip-compatible interface

**`install`** - Install a package

**`-e`** - "Editable" mode (also written as `--editable`)

**`.`** - Current directory (where `pyproject.toml` is)

## What "Editable Mode" Means

### Without `-e` (Normal Install)
```bash
uv pip install .
```
- Copies your code to `site-packages/`
- Creates an installed package
- **Problem**: Changes to your source code don't take effect until you reinstall

### With `-e` (Editable Install)
```bash
uv pip install -e .
```
- **Does NOT copy** your code
- Creates a **link** from `site-packages/` to your source directory
- **Benefit**: Edit your code → changes are immediately available
- Perfect for development!

## What Actually Happens

When you run `uv pip install -e .`, it creates:

1. **A `.egg-link` or `.pth` file** in site-packages pointing to your source:
   ```
   /home/gaze/.venv/lib/python3.14/site-packages/youtube-ai-notetaker.egg-link
   ```
   This file just contains the path to your source directory

2. **Entry point scripts** (if you defined them in `pyproject.toml`):
   ```
   /home/gaze/.venv/bin/youtube-notetaker
   ```

3. **Metadata** about the installed package

## Practical Example

Let's say you have:
```
youtube-ai-notetaker/
├── pyproject.toml
└── src/
    └── youtube_notetaker/
        ├── __init__.py
        └── downloader.py  # has function get_video_id()
```

**After `uv pip install -e .`:**

```python
# In ANY Python file, ANYWHERE on your system:
from youtube_notetaker.downloader import get_video_id
```

This works because Python now knows:
- "youtube_notetaker" is in `sys.path` (via the link)
- It points to your actual source code directory

**You edit `downloader.py`:**
```python
def get_video_id(url):
    print("I just added this line!")  # New line
    # ... existing code
```

**No reinstall needed** - the change is immediately available because Python is reading from your source directory!

## When You'd Use Each

### Use `-e` (editable) when:
- **Developing** your own package (your current situation)
- Running tests while coding
- Making frequent changes
- Working on the package locally

### Use regular install when:
- **Deploying** to production
- Installing dependencies you don't edit (like `ollama`, `rich`)
- Creating a distributable package
- You want a "clean" install

## Visual Comparison

**Normal install:**
```
Your code → Copied to → site-packages/youtube_notetaker/
(editing source has no effect)
```

**Editable install:**
```
Your code ← Link ← site-packages/ points here
(editing source immediately works)
```

## With uv Specifically

`uv` is faster than `pip`, so `uv pip install -e .` does the same thing as `pip install -e .` but:
- Resolves dependencies faster
- Creates `uv.lock` for reproducible installs
- Better caching

You could also use:
```bash
uv sync  # Installs everything from pyproject.toml + locks it
```

But `uv pip install -e .` is explicit about editable mode.

## Try It Out

If you restructured your project, you'd:

```bash
# 1. Restructure (move code to src/youtube_notetaker/)
# 2. Install in editable mode
uv pip install -e .

# 3. Now this works from anywhere:
python -c "from youtube_notetaker.downloader import get_video_id; print(get_video_id('https://youtube.com/watch?v=abc123'))"

# 4. Edit downloader.py, run again - sees your changes immediately!
```

The `.` means "install the package defined in the current directory's `pyproject.toml`". The `-e` means "don't copy it, just link to it so I can keep editing".
