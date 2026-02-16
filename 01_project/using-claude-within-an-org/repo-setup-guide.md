# Setting Up a Repo for Agentic Claude Code Workflows

## 1. CLAUDE.md - The Core Control File

The single most impactful file. It constrains the agent's behavior per repo. It should encode:

- **Workflow rules**: always work on a branch, never push to main, raise PRs
- **Definition of done**: tests must pass, lint must pass, coverage thresholds met before raising a PR
- **Commit discipline**: atomic commits, conventional commit messages
- **Documentation requirements**: update relevant docs with any change
- **What's off-limits**: don't modify CI config, don't change infra, don't touch secrets

## 2. TASKS.md - Definition of Done

A file in the repo that describes:

- What the current goals are
- What "done" looks like for each task
- Acceptance criteria the agent can check against

This gives the agent (and reviewers) a shared contract.

## 3. Documentation Templates

Standardized docs across repos so agents produce consistent output:

- `docs/SECURITY.md` - security posture, auth flows, known risks
- `docs/TESTING.md` - test strategy, how to run tests, coverage expectations
- `docs/ARCHITECTURE.md` - system design, dependencies, data flows

The consistent format allows later aggregation across services to build a systems-level architectural view.

## 4. Branch Protection & PR Gates

On the repo itself (GitHub settings or via `gh` CLI):

- Require PRs to merge into main
- Require status checks to pass (tests, lint, coverage)
- Require at least one human review

This is blast radius control - the agent can do whatever it wants on a branch, but nothing lands without passing gates and human review.

## 5. CI Pipeline With Quality Gates

A GitHub Actions workflow (or equivalent) that runs on every PR:

- Unit tests
- Integration tests (if applicable)
- Linting
- Coverage check with a minimum threshold
- Security scanning (e.g., `npm audit`, `trivy`, `bandit`)

Automated testing is the foundation for trusting agent-produced code. You cannot deploy unread code without it.

## 6. Agentic Personas via Prompt Files

For specialized lenses (security, observability), create separate prompt files:

- `.claude/prompts/security-audit.md` - review for OWASP top 10, dependency vulnerabilities, secrets exposure
- `.claude/prompts/observability-audit.md` - check logging coverage, metrics instrumentation, trace propagation
- `.claude/prompts/test-coverage.md` - identify untested code paths, generate tests

These are reusable and can be run against any repo.

## 7. Stacked PR Workflow

For longer tasks, the CLAUDE.md should instruct the agent to:

- Break work into logical chunks
- Create a chain of branches (`feature/step-1` -> `feature/step-2` -> ...)
- Raise separate PRs for each step
- Each PR should be independently reviewable

## Priority Order for Setup

1. **CLAUDE.md** with strict workflow rules and definition-of-done criteria
2. **CI pipeline** with test + lint + coverage gates
3. **Branch protection** on main
4. **TASKS.md** with concrete tasks to throw at the agent
5. **Doc templates** so the agent has a structure to fill in
6. **Prompt files** for specialized personas
