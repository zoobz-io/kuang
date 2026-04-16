# Create Issue

Create a well-formed GitHub issue that triggers the development workflow.

## Purpose

Transform a validated plan into a GitHub issue that:
- Clearly states the objective
- Provides acceptance criteria
- Gives Fidgel enough context to architect
- Can be tracked through completion

## Execution

1. Ensure plan has been validated (validate-plan skill)
2. Structure the issue with required sections
3. Create issue via gh CLI
4. Apply appropriate labels
5. Confirm issue created

## Issue Structure

```markdown
## Objective

[What needs to be accomplished. One clear statement.]

## Context

[Why this matters. Background information Fidgel needs for architecture.]

## Acceptance Criteria

- [ ] [Specific, verifiable criterion]
- [ ] [Specific, verifiable criterion]
- [ ] [Specific, verifiable criterion]

## Scope

### In Scope
- [What this issue covers]

### Out of Scope
- [What this issue explicitly does NOT cover]

## Notes

[Any additional context, constraints, or considerations]
```

## Labels

Apply appropriate labels:
- `feature` / `bug` / `docs` / `infra` — type
- `needs-architecture` — triggers Fidgel

## Command

```bash
gh issue create \
  --title "[type]: [brief description]" \
  --body "[structured body]" \
  --label "needs-architecture"
```

## Output

```
## Issue Created

**Issue:** #[number]
**Title:** [title]
**URL:** [url]

Assigned label: needs-architecture
Fidgel will architect the solution.
```

## What This Skill Does NOT Do

- Validate the plan (use validate-plan first)
- Architect the solution (Fidgel handles this)
- Assign to implementers (workflow handles this)

This creates the issue. The workflow takes over from here.
