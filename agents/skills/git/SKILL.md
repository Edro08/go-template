---
name: git
description: Git workflow conventions, branch-prefixed commit message formatting, and mandatory explicit confirmation rule before committing.
---

# Git Workflow & Commit Conventions Skill

Use this skill when managing branches, preparing changes, drafting commit messages, and performing Git operations across any repository.

## 1. Mandatory Commit Message Format

Every commit message **MUST** strictly start with the exact name of the active branch, followed by `: ` and a concise description of the change written in the imperative or present tense.

### Structure:

```text
"<branch-name>: <change description>"
```

### Examples by branch type:

* **Task branch (`TASK-...`)**:
    - Branch: `TASK-001`
    - Commit: `"TASK-001: implement health checks for database"`

* **Feature branch (`feat/...`)**:
    - Branch: `feat/integrate-go-utils`
    - Commit: `"feat/integrate-go-utils: add git convention skill and register in agent instructions"`

* **Bugfix branch (`fix/...`)**:
    - Branch: `fix/read-timeout`
    - Commit: `"fix/read-timeout: adjust server timeout to prevent early disconnect"`

---

## 2. Explicit User Confirmation (Golden Rule)

1. **Never commit automatically**:
    - AI agents must **NEVER** run `git commit` autonomously or as part of an automated chain without prior user request.
2. **Keep changes prepared in the Working Tree**:
    - After modifying code and passing verification tests, leave changes staged or ready in the working area.
3. **Wait for explicit user approval**:
    - Only proceed with `git commit` and `git push` when the user explicitly instructs it (e.g., *"commit"*, *"commit and push"*).

---

## 3. Branching Workflow

- **Dedicated working branches**: All development and bug fixes must occur in a dedicated branch (`feat/*`, `fix/*`, `TASK-*`).
- **Protected branches**: Never develop or commit directly to `main` or `master`.
- **Active branch verification**: Always check the active branch (e.g., via `git branch --show-current`) before drafting commit messages.
