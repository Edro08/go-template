# Agent Instructions

Read the applicable specifications in `agents/specs/` before modifying code:
- `agents/specs/architecture.md`: package boundaries, dependency direction, and composition rules.
- `agents/specs/code_rules.md`: language-specific formatting and linting requirements.
- `agents/specs/code_testing.md`: test style and required verification commands.

Read the applicable skills in `agents/skills/` before modifying code:
- `agents/skills/git/SKILL.md`: git workflow rules, branch-prefixed commit messages, and commit confirmation rule.
- `agents/skills/go-utils/SKILL.md`: usage patterns for the shared go-utils library (config, logger, cors, jwt, cipher).
- `agents/skills/test/`: testing workflows and conventions.
- `agents/skills/test/database/SKILL.md`: generic SQL database testing patterns.

## Context and Progress Tracking

Agents may use the `agents/context/` directory to deposit and maintain context `.md` files for task tracking, reporting, and knowledge retention:
- `history.md`: chronicle of project milestones, phases, and snapshots. Must start with `# Historial de [project_name]` and list phases/milestones in reverse chronological order (most recent first / cortes en orden cronológico descendente).
- `phases.md` / `fases.md`: implementation phases, roadmaps, and step-by-step checklists.
- `reports.md` / `reportes.md`: analysis reports, technical audits, and benchmark summaries.
- Additional markdown files as needed to retain execution context across sessions.
