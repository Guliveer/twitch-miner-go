# Contributing

Thank you for your interest in contributing to **twitch-miner-go**! This document covers the commit conventions, git hooks setup, and automated versioning workflow used by this project.

## Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/) to drive automated versioning and changelog generation. Every commit/PR message must follow this format:

```
<type>[optional scope][!]: <description>
```

**Allowed types and their version bump effect:**

| Type       | Description             | Version Bump  |
| ---------- | ----------------------- | ------------- |
| `feat`     | New feature             | Minor (1.x.0) |
| `fix`      | Bug fix                 | Patch (1.0.x) |
| `perf`     | Performance improvement | Patch         |
| `refactor` | Code refactoring        | Patch         |
| `build`    | Build system changes    | Patch         |
| `docs`     | Documentation only      | None          |
| `style`    | Code style/formatting   | None          |
| `test`     | Adding/updating tests   | None          |
| `ci`       | CI/CD changes           | None          |
| `chore`    | Maintenance tasks       | None          |

**Breaking changes** — adding `!` after the type (e.g., `feat!:`) or including `BREAKING CHANGE:` in the commit body triggers a **major** version bump (x.0.0).

**Examples:**

```
feat: add Discord notification support
fix(auth): resolve token refresh race condition
feat!: redesign configuration file format
docs: update installation instructions
chore: update dependencies
```

## Setting Up Git Hooks

Run the hook installer to enable local commit validation:

```bash
./scripts/install-hooks.sh
```

This configures two git hooks:

- **`commit-msg`** — validates that every commit message follows the Conventional Commits format before it is recorded.
- **`pre-push`** — re-validates all outgoing commits before they are pushed to the remote.

> **Tip:** The hooks are stored in [`scripts/githooks/`](scripts/githooks/) and the installer ([`scripts/install-hooks.sh`](scripts/install-hooks.sh)) simply points `core.hooksPath` at that directory — no files are copied into `.git/`.

## Automated Versioning

Releases are fully automated through the [CI workflow](.github/workflows/ci.yml):

1. Developers write commits using the Conventional Commits format described above.
2. Git hooks enforce the format locally (see [Setting Up Git Hooks](#setting-up-git-hooks)).
3. CI validates the commit format on pull requests.
4. On merge to `main`, the CI pipeline runs in order: **build** → **version**:
   - **build** — compiles, runs tests, vet, and lint
   - **version** — analyzes commit messages and creates a git tag and GitHub Release

Docker images are published separately via the Docker workflow. No manual tags are needed — just write well-formatted commits and the pipeline handles the rest.
