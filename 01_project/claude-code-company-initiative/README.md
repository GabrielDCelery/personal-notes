# What is this repo for?

Shared Claude Code configuration for all company repositories. Provides default project context and rules for common tasks like README generation. Designed to be used by developers locally or agents running on remote machines.

## What is Included

```
.claude/
├── CLAUDE.md              # Default project context (tech stack, conventions)
└── rules/
    └── readme-style.md    # README generation guidelines
```

## Setup

Clone this repo:

```sh
git clone git@github.com:your-org/dotclaude.git
cd dotclaude
```

Then choose one of the following options:

### Option A: Per-project setup

Symlink the `.claude` directory into a specific project:

```sh
# The below is assuming you are in the dotclaude repo dir
ln -s $(pwd)/.claude /path/to/your-project/.claude
```

### Option B: User-wide setup

Symlink the individual files into your existing `~/.claude` directory:

```sh
# The below is assuming you are in the dotclaude repo dir
ln -s $(pwd)/.claude/CLAUDE.md ~/.claude/CLAUDE.md
ln -s $(pwd)/.claude/rules ~/.claude/rules
```

> [!WARNING]
> Don't symlink the entire `.claude` directory to home - it contains Claude Code's data files (cache, history, settings, etc.).

> [!NOTE]
> To update, run `git pull` in this directory. Symlinks pick up changes automatically.

## Adding Project-Specific Context

If a project needs additional context, create a `CLAUDE.md` at the project root. Claude Code reads both files and combines them.

```
your-project/
├── CLAUDE.md          # Project-specific context
├── .claude -> ...     # Symlinked company defaults
└── src/
```
