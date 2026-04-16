# Review Criteria

Secret, repo-specific review criteria. Only Armitage reads this file.

## Mission Criteria

### What This Repo MUST Achieve

- Template must produce a repo that builds and passes tests on first clone
- Both API surfaces (api/ and admin/) must be functional, not decorative
- Store implementations must work with their providers
- Skills must produce correct output when invoked
- All registration points must be wired correctly

### What This Repo MUST NOT Contain

- No hardcoded zoobzio-specific values that template users must find-and-replace
- No application-specific business logic — infrastructure only
- No placeholder content masquerading as implementation
- No dependencies beyond stdlib and zoobzio packages

## Review Priorities

Ordered by importance. When findings conflict, higher-priority items take precedence.

1. Security: no secrets, no unsafe defaults, no vulnerable dependencies
2. Correctness: both API surfaces build and serve, store abstractions function, boundaries process correctly
3. Completeness: all template layers present (models, stores, contracts, wire, transformers, handlers per surface)
4. Consistency: naming conventions match across surfaces, patterns uniform across entity artifacts
5. Quality: documentation accurate, skills produce valid output

## Severity Calibration

Guidance for how Armitage classifies finding severity for this specific repo.

| Condition | Severity |
|-----------|----------|
| Boundary processing fails (cereal encryption/masking) | Critical |
| Store implementation doesn't satisfy its contract | Critical |
| Missing registration in stores.go or handlers.go | High |
| Surface-specific artifact in wrong directory (api/ vs admin/) | High |
| Naming convention violation (singular model, plural store) | Medium |
| Missing godoc on exported type | Medium |
| Inconsistent wire type naming between surfaces | Low |
| Minor style inconsistency in template comments | Low |

## Standing Concerns

Persistent issues or areas of known weakness that should always be checked.

- Boundary registration must happen before sum.Freeze() — ordering matters
- Wire type masking differs between surfaces (api/ masks, admin/ exposes) — verify both
- Store implementations are shared across surfaces — contract compatibility must be verified
- Migration ordering must match entity dependency order

## Out of Scope

Things the red team should NOT flag for this repo, even if they look wrong.

- cmd/ binaries are intentionally minimal — template users flesh them out
- MISSION.md references "sumatra" specifically — this is correct for the template itself
- Empty handler registrations are expected — template starts with no entities
- The testing/ directory may have minimal fixtures — template users add their own
