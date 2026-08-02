---
name: always-on-guidance
description: Global coding rules and tool usage discipline
tags: [always-on, coding, quality]
---

# Always-On Guidance

Rules that apply to every coding session.

## Critical Safety

- Never force push to main

## Tool Usage Discipline

- Check `hivemind_find` before solving problems from scratch
- Read files before editing — never guess at content
- Use `org_*` tools for work item management
- Use `comms_*` tools for agent messaging and file reservations
- Use `forge_*` tools for multi-agent coordination
- Use `hivemind_*` tools for learning storage and retrieval

## Code Quality

### Structure
- Functions do one thing well
- No dead code or unused imports

### Clarity
- Names reveal intent — no abbreviations
- Comments explain *why*, not *what*
- Error messages include context

## Testing

### Test Infrastructure
- Use `db.OpenMemory()` for database tests
- Use `t.TempDir()` for filesystem tests
- Standard library `testing` package only — no testify

### Test Practice
- Write tests for all new code
- Test names: `TestXxx_Description`

## Error Handling

### Error Propagation
- Return errors, don't panic
- Wrap errors with context: `fmt.Errorf("operation: %w", err)`

### Error Coverage
- Handle all error paths — no ignored returns
- Use `errors.Is` for sentinel error checks

## Git Discipline

- Conventional commits: `type: description`
- Commit early, commit often
