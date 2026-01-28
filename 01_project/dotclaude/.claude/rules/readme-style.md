# README Generator

When asked to create or improve a README, follow this process.

## Analysis Phase

1. Scan project structure for: package.json, tsconfig.json, go.mod, pyproject.toml, uv.lock, poetry.lock, Dockerfile, docker-compose.yml, Makefile, mise.toml, Taskfile.yml
2. Check for existing docs in /docs, any \*.md files scattered in the project, and existing README.md (may be outdated)
3. Check CI/CD files (.github/workflows, .circleci/, buildspec.yml) for deployment info

## Information Discovery

| Info needed  | Where to find it                                                            |
| ------------ | --------------------------------------------------------------------------- |
| Project name | Directory name, package.json name field                                     |
| Description  | package.json description, existing README                                   |
| Language     | File extensions, package manager files                                      |
| Dependencies | package.json, go.mod, pyproject.toml, uv.lock, poetry.lock                  |
| Run command  | Makefile, Taskfile.yml, mise.toml, package.json scripts                     |
| Test command | Makefile, Taskfile.yml, mise.toml, package.json scripts                     |
| Env vars     | .env.example, docker-compose.yml, os.Getenv/process.env                     |
| Deployment   | .github/workflows, .circleci/, infrastructure/, \*.tf files, serverless.yml |

## Required Sections

Every README must have:

1. **Description** - A few sentences on what this does and why
2. **Quick Start** - Clone, setup, run (copy-pasteable commands)
3. **Architecture** - Language, database, external services
4. **Configuration** - Environment variables table, where secrets live
5. **Development** - How to run tests and linters
6. **Deployment** - How and where this gets deployed

## Template

```markdown
# Project Name

Describe what this service/library does and what problem it solves. Keep it to 1-3 sentences.

## Quick Start

\`\`\`sh
git clone <repo>
cd <repo>
make setup
make run
\`\`\`

## Architecture

- **Language/Framework**:
- **Database**: DynamoDB / RDS (PostgreSQL, MySQL) / None
- **Messaging**: SQS / SNS / None
- **Storage**: S3 / None
- **External APIs**:

## Configuration

| Variable | Description | Required | Default |
| -------- | ----------- | -------- | ------- |

Secrets are stored in: [SSM Parameter Store / Secrets Manager path]

## Development

\`\`\`sh
make test
make lint
\`\`\`

## Deployment

- **Production**: Deployed via [CircleCI / GitHub Actions / CodePipeline] to [ECS / Lambda]
- **Environments**: dev / staging / prod
```

## Rules

- Be concise - developers scan, they don't read
- Commands must be copy-pasteable
- Don't invent information - skip what you can't find
- Use tables for configuration variables
- Link to external docs, don't duplicate
