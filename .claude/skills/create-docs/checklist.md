# Documentation Create Checklist

## Phase 1: Discovery

### Application Understanding
- [ ] Read the application's API surfaces (api/, admin/)
- [ ] Examine existing documentation, comments, or doc.go files
- [ ] Study test files for usage patterns and edge cases
- [ ] Review config/ for configuration options
- [ ] Review any existing README for context

### Scope Determination
- [ ] Ask: "Which documentation files are needed?" (overview, quickstart, concepts, architecture, guides, reference)
- [ ] Ask: "What deployment targets need documentation?" (Docker, Kubernetes, bare metal)
- [ ] Ask: "What features need dedicated guides beyond required ones?"

### Audience Understanding
- [ ] Ask: "Who is the intended reader—what do they already know?"
- [ ] Ask: "What questions do users commonly ask?"
- [ ] Ask: "What concepts are most confusing to newcomers?"

### Content Inventory
- [ ] List core abstractions that need explanation in Concepts
- [ ] List internal mechanisms that need explanation in Architecture
- [ ] List features that need dedicated guides
- [ ] Map API surfaces for Reference documentation

## Phase 2: Overview (`1.learn/1.overview.md`)

**Must answer:**
- What is this application? (one sentence)
- What problem does it solve?
- Who is it for?
- What can you do with it?
- Where to go next?

**Structure:**
- [ ] Opening paragraph: what it is + core value proposition
- [ ] "The Purpose": the problem this application solves
- [ ] "API Surfaces": public (api/) vs admin (admin/) overview
- [ ] "Key Features": scannable list of capabilities
- [ ] "Next Steps": links to quickstart, concepts, reference

## Phase 3: Quickstart (`1.learn/2.quickstart.md`)

**Must contain:**
- Prerequisites (Go version, dependencies)
- Installation/clone instructions
- Configuration for first run
- Running the application
- Verifying it works

**Structure:**
- [ ] Prerequisites section with specific versions
- [ ] Installation section with commands
- [ ] Configuration section with minimal required config
- [ ] Running section with startup commands
- [ ] Verification section with expected output
- [ ] Next Steps linking to concepts, configuration guide

## Phase 4: Concepts (`1.learn/3.concepts.md`)

**Must contain:**
- API surface model (public vs admin)
- Layer architecture (models, stores, handlers, etc.)
- Key domain concepts
- Cross-links to Reference sections

**Structure:**
- [ ] Brief intro establishing what mental models the page covers
- [ ] "API Surfaces": public (api/) and admin (admin/) purposes
- [ ] "Layers": models, stores, contracts, handlers, wire, transformers
- [ ] "Domain Concepts": application-specific entities
- [ ] Next Steps linking to architecture, reference

## Phase 5: Architecture (`1.learn/4.architecture.md`)

**Must contain:**
- System diagram (layers and flow)
- Layer responsibilities
- Data flow through the system
- Design rationale as Q&A

**Structure:**
- [ ] Brief intro (who this is for, what it covers)
- [ ] System Overview with diagram
- [ ] Layer sections (what each layer does, how they interact)
- [ ] Data Flow section (request → response path)
- [ ] Design Q&A (common questions about design decisions)
- [ ] Next Steps linking to guides, reference

## Phase 6: Required Guides (`2.guides/`)

All guide files use numeric prefixes.

### Configuration Guide (`1.configuration.md`)
- [ ] Environment variables table
- [ ] Config file format
- [ ] Per-surface configuration
- [ ] Secret management

### Deployment Guide (`2.deployment.md`)
- [ ] Production prerequisites
- [ ] Deployment options (Docker, Kubernetes, etc.)
- [ ] Database setup
- [ ] Reverse proxy configuration

### Testing Guide (`3.testing.md`)
- [ ] Running tests locally
- [ ] Test structure
- [ ] Integration test setup
- [ ] Writing new tests

### Troubleshooting Guide (`4.troubleshooting.md`)
- [ ] Common errors and solutions
- [ ] Debugging techniques
- [ ] Log analysis
- [ ] Getting help

## Phase 7: API Reference (`3.api-reference/`)

### Public API Reference (`1.public-api.md`)
- [ ] Authentication requirements
- [ ] Endpoint tables with: method, path, description
- [ ] Request/response examples
- [ ] Error codes

### Admin API Reference (`2.admin-api.md`)
- [ ] Authentication requirements (if different)
- [ ] Endpoint tables with: method, path, description
- [ ] Request/response examples
- [ ] Admin-specific operations

## Phase 8: Operations (`4.operations/`)

### Monitoring (`1.monitoring.md`)
- [ ] Health check endpoints
- [ ] Metrics endpoints
- [ ] Logging configuration
- [ ] Alerting recommendations

### Backup & Recovery (`2.backup-recovery.md`)
- [ ] What to back up
- [ ] Backup procedures
- [ ] Recovery procedures
- [ ] Disaster recovery

### Scaling (`3.scaling.md`)
- [ ] Horizontal scaling approach
- [ ] Database scaling considerations
- [ ] Caching strategies
- [ ] Load balancing

## Phase 9: Verification

- [ ] All files have frontmatter
- [ ] All cross-references use numbered paths
- [ ] All required guides present
- [ ] Tone is consistent across documents
- [ ] Next Steps in each document link correctly
