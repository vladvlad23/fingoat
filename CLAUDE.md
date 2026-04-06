# FinGoat — Project Instructions

## Agent Memory Convention

All agents have a persistent memory directory at `.claude/agent-memory/<agent-name>/`.

- `MEMORY.md` = index, loaded into every session (truncates at 200 lines — keep it concise)
- Create topic files for details; link them from `MEMORY.md`
- **Save**: stable patterns, conventions, architectural decisions, user preferences
- **Don't save**: session-specific work, anything derivable from the repo, git history, or already in CLAUDE.md
- Update stale memories; never create duplicates

### Memory types
| Type | When to save | Body structure |
|------|-------------|----------------|
| **user** | Role, preferences, knowledge level | Plain prose |
| **feedback** | Corrections AND validated non-obvious approaches | Rule → **Why:** → **How to apply:** |
| **project** | Ongoing work, decisions, deadlines | Fact → **Why:** → **How to apply:**; use absolute dates |
| **reference** | Pointers to external systems (Linear, Grafana, etc.) | System + URL + purpose |
