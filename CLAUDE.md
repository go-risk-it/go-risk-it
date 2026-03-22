# go-risk-it — Personal Project

This is a personal project (Risk board game in Go), **not** related to Booking.com work.

## Overrides to global instructions

- **No session logging** — Do NOT write entries to `~/claude-work-log/`. Skip the "Automatic Session Logging" section from global CLAUDE.md entirely.
- **No domain knowledge persistence** — Do NOT write to `~/.claude/knowledge/`. Skip the "Knowledge Management" rules.
- **No Jira conventions** — No ticket prefixes in branch names or commit messages. Use simple descriptive names.
- **No WBSO** — Skip any WBSO-related skills or time tracking.
- **No standup/wrap-up** — Skip work log based skills (standup, weekly-summary, follow-ups, wrap-up, today).

## What still applies

- All coding behavior rules (execute don't plan, no over-engineering, security practices)
- Git best practices (but without Jira ticket conventions)
- Project MEMORY.md in `.claude/projects/` for project-specific context
