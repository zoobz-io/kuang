# Mission Review Checklist

## Phase 1: Mission Document Assessment

### MISSION.md Exists
- [ ] MISSION.md is present in `.claude/`
- [ ] MISSION.md is non-empty

### Required Sections
- [ ] Purpose/objective section exists (clear statement of what the application does)
- [ ] Scope boundaries defined (what's in, what's out)
- [ ] Success criteria exist and are verifiable
- [ ] Non-goals exist

### Section Quality — Where Does the Mission Lie?
- [ ] Purpose is a single, focused statement — or is it hedging?
- [ ] Scope boundaries are concrete — or are they vague enough to mean anything?
- [ ] Exclusions are concrete — or are they so broad they exclude nothing?
- [ ] Success criteria are step-by-step and testable — or are they aspirational?
- [ ] Non-goals are specific decisions — or are they obvious negatives nobody would pursue?

## Phase 2: Application Inventory

### Application Layers
- [ ] Inventory API surfaces and their consumers
- [ ] Inventory shared layers (models, stores, config, events, migrations)
- [ ] Inventory surface-specific layers per surface
- [ ] Inventory internal packages
- [ ] Inventory binaries and their purpose

### Capabilities
- [ ] Describe the application's actual behavior (from code, not docs)
- [ ] Identify the core workflow each surface serves
- [ ] Identify secondary/supporting capabilities

## Phase 3: Purpose Alignment — Where Does Implementation Contradict the Mission?

### Core Alignment
- [ ] Does each surface serve its stated consumer?
- [ ] Do shared layers actually serve all surfaces?
- [ ] Are surface-specific layers correctly separated?
- [ ] Does the application do significantly more than the purpose states? (scope creep)
- [ ] Does the application do significantly less than the purpose states? (broken promise)

### Feature Classification
For each exported capability:
- [ ] Classify as: aligned, adjacent, contradictory, or orphaned
- [ ] Document evidence for non-aligned classifications

### Adjacent Features — Unauthorized Scope
- [ ] List features related but not explicitly covered by mission
- [ ] For each: should the mission expand or the feature be removed?

### Contradictory Features — Violations
- [ ] List features that contradict non-goals or exclusions
- [ ] For each: is the feature wrong or is the mission outdated?

### Orphaned Features — Dead Weight
- [ ] List features that serve no apparent mission purpose
- [ ] For each: useful addition (update mission) or dead weight (remove)?

## Phase 4: Contains/Excludes Verification — What's Missing? What Shouldn't Be Here?

### "What This Package Contains"
For each listed item:
- [ ] Item actually exists in the application — or is it a phantom?
- [ ] Item is functional, not just present
- [ ] Item matches the description — or does it do something different?

### "What This Package Does NOT Contain"
For each listed exclusion:
- [ ] The excluded thing is genuinely absent — or did it sneak in?
- [ ] No workarounds or partial implementations of excluded items
- [ ] Exclusion is still a reasonable boundary

### Unlisted Items
- [ ] Identify anything the application contains that isn't listed in either section
- [ ] Classify each: belongs in contains, belongs in excludes, or shouldn't exist

## Phase 5: Success Criteria Verification — Which Criteria Fail?

For each success criterion:

### Achievability
- [ ] Can the criterion be achieved with the current implementation — or is it blocked?
- [ ] Are the described steps actually possible to follow?
- [ ] What missing functionality blocks the criterion?

### Walk-Through
- [ ] Attempt to follow each criterion step-by-step
- [ ] Document where the walk-through fails or gets stuck
- [ ] Each failure point is a finding

## Phase 6: Non-Goals Verification — Which Non-Goals Are Violated?

For each non-goal:

### Respected
- [ ] The implementation does not pursue this goal — or does it?
- [ ] No partial or accidental implementation of the non-goal
- [ ] No API surface that implies the non-goal is supported

### Relevance
- [ ] The non-goal is still relevant (not outdated)
- [ ] The non-goal reflects a deliberate design decision

## Phase 7: Philosophy Alignment — Where Does Code Break Zoobzio Principles?

### Dependency Policy
- [ ] Production dependencies are minimal — or is there bloat?
- [ ] Each dependency is justified — or is stdlib sufficient?
- [ ] Provider-specific deps are in submodules — or leaking into root?

### Type Safety
- [ ] Generics used where appropriate — or is there interface{} laziness?
- [ ] No `any` in public API where type parameters would work

### Boundaries
- [ ] Data transformations at identifiable boundaries — or scattered?
- [ ] Boundaries are explicit, not implicit

### Composition
- [ ] Interfaces are small and composable — or wide and rigid?
- [ ] Clear separation between processors, connectors, and values

### Errors
- [ ] Semantic errors for expected failures — or bare error strings?
- [ ] Errors carry context

### Context
- [ ] I/O operations accept `context.Context` — or missing?

## Phase 8: MISSION.md Currency — Where Does the Mission Document Lie?

### Is the Mission Current?
- [ ] Mission reflects the application as it exists today — or a past/future version?
- [ ] No sections describing aspirational features as present
- [ ] No sections describing removed features as present
- [ ] Success criteria achievable with current code
