---
title: Claude README Generator Agent
tags: claude, automation, documentation
---

# Claude README Generator Agent

How to set up a Claude Code agent that analyzes repositories and generates/refines READMEs following our company standard.

## Goal

Automate README creation across 200 repositories without bloating each repo's CLAUDE.md file.

## Project Structure for Claude Rules

Use the `.claude/rules/` directory to keep CLAUDE.md slim. Rules are only loaded when relevant:

```
your-repo/
├── .claude/
│   ├── CLAUDE.md              # Slim, core project context only
│   └── rules/
│       └── readme-generator.md  # Only loaded when working on READMEs
```

## Distribution Strategies

### Option A: Add to each repo (code-level)

Commit the rule to every repository:

```
each-repo/
├── .claude/
│   └── rules/
│       └── readme-generator.md
```

| Pros                             | Cons                              |
| -------------------------------- | --------------------------------- |
| Version controlled with the repo | 200 copies to maintain            |
| Everyone gets it automatically   | Updates require PRs to every repo |
| Can be enforced via CI           |                                   |

**Mitigation:** Use a script or GitHub Action to sync the rule file across all repos when it changes.

### Option B: Developer-local settings

Claude Code supports user-level settings at `~/.claude/rules/`:

```
~/.claude/
└── rules/
    └── readme-generator.md
```

| Pros                                   | Cons                              |
| -------------------------------------- | --------------------------------- |
| Single source of truth                 | Each developer needs to set it up |
| Instant updates                        | Can drift between developers      |
| Works across all repos without commits | Not version controlled            |

### Option C: Hybrid (recommended)

Create a shared repo for Claude rules that developers clone and symlink:

```
claude-rules/
├── readme-generator.md
├── code-review.md
└── install.sh
```

**install.sh:**

```sh
#!/bin/bash
mkdir -p ~/.claude/rules
for rule in *.md; do
  ln -sf "$(pwd)/$rule" ~/.claude/rules/
done
echo "Rules installed. Pull this repo to get updates."
```

**Setup for developers:**

```sh
git clone git@company.com:claude-rules.git
cd claude-rules && ./install.sh
```

| Pros                                   | Cons                             |
| -------------------------------------- | -------------------------------- |
| Rules version controlled in one place  | Requires initial developer setup |
| Developers get updates with `git pull` |                                  |
| No bloat in individual repos           |                                  |
| Works across all 200 repos immediately |                                  |

## The README Generator Rule

Create `readme-generator.md` with the following content:

```markdown
# README Generator

When asked to create or improve a README, follow this process.

## Analysis Phase

1. Scan the project structure to understand the codebase
2. Look for: package.json, go.mod, requirements.txt, Cargo.toml, Dockerfile, docker-compose.yml, Makefile
3. Check for existing documentation in /docs, wiki links, or inline comments
4. Identify the language, framework, dependencies, and entry points
5. Look for .env.example files to understand configuration
6. Check CI/CD files (.github/workflows, .gitlab-ci.yml, Jenkinsfile) for deployment info

## Information Discovery

| Info needed  | Where to find it                                                 |
| ------------ | ---------------------------------------------------------------- |
| Project name | Directory name, package.json name field                          |
| Description  | package.json description, existing README, repo description      |
| Language     | File extensions, package manager files                           |
| Dependencies | package.json, go.mod, requirements.txt                           |
| Run command  | Makefile, package.json scripts, Dockerfile CMD                   |
| Test command | Makefile, package.json scripts, CI config                        |
| Env vars     | .env.example, docker-compose.yml, grep for os.Getenv/process.env |
| Deployment   | .github/workflows, k8s/, helm/, terraform/                       |

## Required Sections

Every README must have:

- **One-line description** - What does this do?
- **Overview** - Purpose, owners, status, related repos
- **Quick Start** - Clone, setup, run (copy-pasteable commands)
- **Architecture** - Language, database, external services
- **Configuration** - Environment variables table, where secrets live
- **Development** - How to run tests and linters
- **Deployment** - How and where this gets deployed

## Rules

- Be concise - developers scan, they don't read
- Commands must be copy-pasteable
- Don't invent information - if you can't find it, mark with [TODO]
- Use tables for configuration variables
- Link to external docs, don't duplicate
```

## Usage

Run Claude Code in any target repository:

```sh
cd /path/to/some-service
claude "Analyze this repo and generate a README following our standard template"
```

Or to refine an existing README:

```sh
claude "Review and improve the README based on our company standards"
```

## Scaling to 200 Repos

1. **Batch processing** - Script that loops through repos and runs Claude on each
2. **PR workflow** - Have Claude create branches and PRs for human review
3. **Track TODOs** - Collect `[TODO]` markers to identify what needs human input
4. **Iterate** - Start with 5 repos, refine the rule based on output quality, then scale

## Related

- [how-to-construct-a-readme.md](./how-to-construct-a-readme.md) - The README template and examples
