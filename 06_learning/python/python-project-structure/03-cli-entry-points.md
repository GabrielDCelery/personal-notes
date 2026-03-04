# CLI Entry Points in Python

## Question
How would it work with `cli.py`? What would I run with the new structure?

## Two Ways to Run CLI Tools

### Method 1: Entry Points (Professional/Recommended)

With the new structure:
```
youtube-ai-notetaker/
├── pyproject.toml
└── src/
    └── youtube_notetaker/
        ├── __init__.py
        ├── cli.py
        └── downloader.py
```

**Update `pyproject.toml`:**
```toml
[project]
name = "youtube-ai-notetaker"
version = "0.1.0"
# ... your existing config ...

[project.scripts]
youtube-notetaker = "youtube_notetaker.cli:main"
# or shorter:
ytnotetaker = "youtube_notetaker.cli:main"
```

**In `cli.py`:**
```python
import sys
from youtube_notetaker.downloader import (
    download_video,
    get_video_id,
    get_transcript,
    analyze_video_content,
    generate_summary_with_diagrams,
)

def main():
    if len(sys.argv) < 2:
        print("Usage: youtube-notetaker <youtube_url>")
        sys.exit(1)

    youtube_url = sys.argv[1]
    # ... your existing main() logic ...

if __name__ == "__main__":
    main()
```

**Install it:**
```bash
uv pip install -e .
```

**Now you can run it from ANYWHERE:**
```bash
youtube-notetaker https://youtube.com/watch?v=...
# or
ytnotetaker https://youtube.com/watch?v=...
```

The entry point creates a small executable script in your virtual environment's `bin/` directory that calls `youtube_notetaker.cli:main()`.

### Method 2: Python Module (Alternative)

You can also run it as a module without entry points:

```bash
python -m youtube_notetaker <url>
```

For this to work, create `src/youtube_notetaker/__main__.py`:
```python
from youtube_notetaker.cli import main

if __name__ == "__main__":
    main()
```

## Comparison: Before vs After

**Your current setup:**
```bash
# Must be in project directory
cd /home/gaze/projects/github-GabrielDCelery/youtube-ai-notetaker
python main.py <url>
```

**With entry points:**
```bash
# Can be anywhere
cd ~/Documents
youtube-notetaker <url>
```

**With module approach:**
```bash
# Must have package in path, but clearer what you're running
python -m youtube_notetaker <url>
```

## Real-World Examples

Popular tools use entry points:
- `pip` → `pip install ...` (entry point to `pip._internal.cli.main:main`)
- `black` → `black .` (code formatter)
- `pytest` → `pytest tests/`
- `django-admin` → `django-admin startproject`

## Advanced: Arguments with argparse

Most CLI tools use `argparse` or `click` instead of raw `sys.argv`:

```python
import argparse

def main():
    parser = argparse.ArgumentParser(
        description="Generate AI-powered notes from YouTube videos"
    )
    parser.add_argument("url", help="YouTube video URL")
    parser.add_argument(
        "--frames",
        type=int,
        default=3,
        help="Number of frames to analyze (default: 3)"
    )
    parser.add_argument(
        "--output",
        help="Output file path (default: <video_id>_analysis.md)"
    )

    args = parser.parse_args()

    # Now use args.url, args.frames, args.output
    video_id = get_video_id(args.url)
    # ...
```

Then users can do:
```bash
youtube-notetaker https://... --frames 5 --output my_notes.md
youtube-notetaker --help  # Automatically generated help
```

## The Magic Behind Entry Points

When you install with entry points, Python creates a small wrapper script:

```bash
# In .venv/bin/youtube-notetaker:
#!/path/to/.venv/bin/python
from youtube_notetaker.cli import main
main()
```

This is why you can just type `youtube-notetaker` instead of `python -m youtube_notetaker` - it's a real executable in your PATH!
