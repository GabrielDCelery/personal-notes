# Organizing Larger Python Projects

## Question
How would someone organize a larger Python project? How would people import code from nested directories or modularize? Also, what do these `__name__` things represent in Python, or what does `if __name__ == "__main__"` do? What other examples are there?

## Typical Project Structure

For larger Python projects, the standard structure looks like:

```
your-project/
├── pyproject.toml          # Project config & dependencies
├── README.md
├── src/
│   └── youtube_notetaker/  # Your package (importable)
│       ├── __init__.py
│       ├── cli.py          # CLI entry point
│       ├── video/
│       │   ├── __init__.py
│       │   ├── downloader.py
│       │   └── frames.py
│       ├── transcript/
│       │   ├── __init__.py
│       │   └── fetcher.py
│       ├── analysis/
│       │   ├── __init__.py
│       │   ├── visual.py
│       │   └── summary.py
│       └── utils/
│           ├── __init__.py
│           └── formatting.py
├── tests/
│   ├── __init__.py
│   ├── test_downloader.py
│   └── test_transcript.py
└── instructions/
```

## How Imports Work in Nested Directories

### Absolute Imports (Preferred)
```python
# From youtube_notetaker/analysis/visual.py
from youtube_notetaker.video.downloader import download_video
from youtube_notetaker.transcript.fetcher import get_transcript
from youtube_notetaker.utils.formatting import format_timestamp
```

### Relative Imports (Within Same Package)
```python
# From youtube_notetaker/analysis/visual.py
from ..video.downloader import download_video  # up one level, into video
from ..transcript import fetcher               # up one, into transcript
from .summary import generate_summary          # same directory
```

## The `__init__.py` Files

These make a directory a Python package. They can be empty or expose a clean API:

```python
# youtube_notetaker/video/__init__.py
from .downloader import download_video, get_video_id
from .frames import extract_frames

__all__ = ['download_video', 'get_video_id', 'extract_frames']
```

Now others can do:
```python
from youtube_notetaker.video import download_video  # clean!
# instead of:
from youtube_notetaker.video.downloader import download_video
```

## About `if __name__ == "__main__"`

### What `__name__` Represents

`__name__` is a special variable Python sets automatically:
- When you **run** a file directly: `__name__ = "__main__"`
- When you **import** a file: `__name__ = "module_name"`

### Example Usage

```python
def main(youtube_url):
    # Your main logic here
    pass

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python main.py <youtube_url>")
        sys.exit(1)
    main(sys.argv[1])
```

This means:
- **Run directly**: `python main.py <url>` → executes the code
- **Imported elsewhere**: `from main import main` → doesn't execute, just imports the function

### Other Special Variables

**`__file__`** - Path to the current file:
```python
INSTRUCTIONS_DIR = Path(__file__).parent / "instructions"
```

**`__package__`** - Package name (for relative imports)

**`__doc__`** - Module's docstring

**`__all__`** - What `from module import *` imports

**`__version__`** - Common convention for package version

## Modularization Strategies

### By Feature (Domain-Driven)
```
youtube_notetaker/
├── video/          # Everything about video handling
├── transcript/     # Transcript fetching/parsing
├── analysis/       # AI analysis logic
└── cli.py          # Command-line interface
```

### By Layer (for web apps)
```
myapp/
├── models/         # Data models
├── services/       # Business logic
├── api/            # HTTP endpoints
└── db/             # Database code
```

### When to Split Files

- When a file exceeds 300-500 lines
- When functions fall into distinct responsibilities
- When you'd want to test them separately
- When you'd want to swap implementations

## Recommended Split for This Project

The current `downloader.py` could be split into:
- `video/downloader.py` - YouTube download logic
- `video/frames.py` - Frame extraction
- `transcript/fetcher.py` - Transcript fetching
- `analysis/visual.py` - Frame analysis with LLaVA
- `analysis/summary.py` - Text summarization
- `utils/formatting.py` - `format_timestamp`
- `config.py` - Ollama client setup

This makes testing easier, allows team members to work on different parts, and makes the code more navigable.
