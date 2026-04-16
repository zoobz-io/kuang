# README Create Checklist

## Phase 1: Discovery

### Application Understanding
- [ ] Read cmd/ to understand binaries (app, admin)
- [ ] Examine go.mod for module path and dependencies
- [ ] Review existing documentation and comments
- [ ] Study API surfaces (api/, admin/) for capabilities
- [ ] Check config/ for configuration requirements

### Purpose Discovery
- [ ] Identify what this application does (one sentence)
- [ ] Identify target users and their needs
- [ ] Identify what distinguishes this from alternatives

### Surface Mapping
- [ ] Document public API surface (api/) purpose and consumers
- [ ] Document admin API surface (admin/) purpose and consumers
- [ ] Note cross-surface relationships

### Capabilities Inventory
- [ ] List primary features (one line each)
- [ ] Map features to documentation locations
- [ ] Flag experimental or advanced features

## Phase 2: Synthesis

- [ ] Draft tagline (purpose in one phrase)
- [ ] Confirm tagline with user
- [ ] Draft overview section name (MUST be application-specific, NOT generic)
- [ ] Confirm overview section name with user
- [ ] Draft capabilities table
- [ ] Confirm capabilities with user

## Phase 3: Writing

### Execute
- [ ] Write README per SKILL.md specifications

### Verify Badges
- [ ] Badge 1 present: CI Status (GitHub Actions URL)
- [ ] Badge 2 present: codecov (codecov.io graph URL)
- [ ] Badge 3 present: CodeQL (GitHub Actions URL)
- [ ] Badge 4 present: Go Version (shields.io go-mod-go-version URL)
- [ ] Badge 5 present: License (shields.io github/license URL)
- [ ] Badge 6 present: Release (shields.io github/v/release URL)
- [ ] All badge URLs match specification exactly (only [app] replaced)

### Verify Structure
- [ ] Section 1: Header with title, all 6 badges, tagline + supporting sentence
- [ ] Section 2: Overview with application-specific name
- [ ] Section 3: API Surfaces (public and admin)
- [ ] Section 4: Quick Start with prerequisites and commands
- [ ] Section 5: Configuration with key options
- [ ] Section 6: Capabilities table with doc links
- [ ] Section 7: Architecture with link to full docs
- [ ] Section 8: Documentation links (Learn / Guides / Reference)
- [ ] Section 9: Development setup
- [ ] Section 10: Contributing (one sentence + link)
- [ ] Section 11: License (single line)

### Verify Prohibitions
- [ ] No alternative badge providers used
- [ ] No generic overview names used
- [ ] No full API reference tables in README
- [ ] No "Problem / Solution" framing
- [ ] README has unique voice (not template-sounding)
