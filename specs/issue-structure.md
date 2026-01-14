# Issue Structure & Dependency Graph

## Epic Overview

Created **5 epics** with **21 detailed tasks** to implement the coupon system.

### Epic Breakdown

| Epic ID | Name | Children | Priority | Status |
|---------|------|----------|----------|--------|
| ubersnap-assessment-7ty | Project Setup & Infrastructure | 4 tasks | P0 | 0% |
| ubersnap-assessment-b2p | Database Layer Implementation | 4 tasks | P0 | 0% |
| ubersnap-assessment-knw | Business Logic & API Layer | 5 tasks | P0 | 0% |
| ubersnap-assessment-mev | Testing & Quality Assurance | 4 tasks | P0 | 0% |
| ubersnap-assessment-hs0 | Documentation & Deployment | 5 tasks | P1 | 0% |

## Dependency Graph

```
Phase 1: Infrastructure (Parallel)
├── 7ty.4: Initialize Go project [30min] ⚡ START HERE
├── 7ty.5: Design database schema [45min] ⚡ START HERE
├── 7ty.6: Setup Docker [60min] ⚡ START HERE
└── 7ty.7: Config management [30min] → depends on 7ty.4

Phase 2: Database Layer (After Schema)
├── b2p.1: DB connection [60min] → depends on 7ty.5
├── b2p.4: Data models [30min] → depends on 7ty.4 (Parallel with b2p.1)
├── b2p.2: Coupon repository [90min] → depends on b2p.1
└── b2p.3: Claim repository [75min] → depends on b2p.1 (Parallel with b2p.2)

Phase 3: Service & API Layer (After Repositories)
├── knw.1: Service with transactions [120min] ⚠️  CRITICAL → depends on b2p.2, b2p.3
├── knw.5: Main entry point [60min] → depends on 7ty.4, b2p.1
├── knw.2: POST /api/coupons [45min] → depends on knw.1 (Parallel below)
├── knw.3: POST /api/coupons/claim [60min] ⚠️  CRITICAL → depends on knw.1
└── knw.4: GET /api/coupons/{name} [45min] → depends on knw.1 (Parallel above)

Phase 4: Testing (After API Implementation)
├── mev.1: Flash Sale test [120min] ⚠️  CRITICAL → depends on knw.3
├── mev.2: Double Dip test [90min] ⚠️  CRITICAL → depends on knw.3 (Parallel with mev.1)
├── mev.3: Integration tests [120min] → depends on knw.2, knw.3, knw.4
└── mev.4: Unit tests [90min] → depends on knw.1 (Parallel with others)

Phase 5: Documentation (Final)
├── hs0.1: README [90min] → depends on 7ty.6, mev.1, mev.2
├── hs0.2: Architecture docs [75min] → depends on knw.1 (Parallel below)
├── hs0.3: E2E verification [60min] → depends on 7ty.6, knw.3, mev.1, mev.2
├── hs0.4: API examples [30min] → depends on knw.2, knw.3, knw.4 (Parallel)
└── hs0.5: Submission checklist [45min] → depends on hs0.1, hs0.3, mev.1, mev.2
```

## Critical Path

The longest dependency chain (critical path):

```
7ty.5 (Design schema)
  ↓ 45min
b2p.1 (DB connection)
  ↓ 60min
b2p.2 (Coupon repo)
  ↓ 90min
knw.1 (Service layer) ⚠️  CRITICAL COMPONENT
  ↓ 120min
knw.3 (Claim endpoint) ⚠️  CRITICAL COMPONENT
  ↓ 60min
mev.1 (Flash Sale test) ⚠️  MUST PASS
  ↓ 120min
hs0.1 (README)
  ↓ 90min
hs0.5 (Final checklist)
  ↓ 45min

Total Critical Path: ~630 minutes (10.5 hours)
```

## Parallelization Opportunities

### Phase 1 (All Parallel):
- 7ty.4: Go project setup
- 7ty.5: Database schema design
- 7ty.6: Docker setup
- **Can save ~60 minutes by doing these concurrently**

### Phase 2 (Partial Parallel):
- b2p.2 and b2p.3 can run in parallel after b2p.1
- b2p.4 can run parallel with b2p.1
- **Can save ~75 minutes**

### Phase 3 (Partial Parallel):
- knw.2, knw.3, knw.4 can all start after knw.1 completes
- knw.5 can run parallel with knw.1
- **Can save ~60 minutes**

### Phase 4 (Heavy Parallel):
- All four test tasks can run largely in parallel
- mev.1 and mev.2 are independent
- mev.4 can start as soon as knw.1 is done
- **Can save ~210 minutes**

### Phase 5 (Partial Parallel):
- hs0.2 and hs0.4 can run parallel with other docs
- **Can save ~75 minutes**

**Total potential time savings through parallelization: ~480 minutes (8 hours)**
**Optimized timeline: ~2.5 hours with perfect parallelization**

## Ready to Start (No Blockers)

These can begin immediately:

1. **ubersnap-assessment-7ty.4**: Initialize Go project [30min] ⚡
2. **ubersnap-assessment-7ty.5**: Design database schema [45min] ⚡
3. **ubersnap-assessment-7ty.6**: Setup Docker [60min] ⚡

## Critical Issues (Must Not Fail)

These issues are make-or-break for the assessment:

1. **knw.1**: Service with atomic transactions
   - This is where SELECT FOR UPDATE and transaction logic lives
   - Prevents race conditions

2. **knw.3**: POST /api/coupons/claim endpoint
   - Most complex endpoint
   - Handles concurrency scenarios

3. **mev.1**: Flash Sale test (50 concurrent, 5 stock)
   - Must result in exactly 5 successful claims
   - Validates no overselling

4. **mev.2**: Double Dip test (10 concurrent, same user)
   - Must result in exactly 1 successful claim
   - Validates UNIQUE constraint enforcement

5. **7ty.5**: Database schema design
   - UNIQUE(user_id, coupon_name) constraint is critical
   - Schema errors cascade to everything else

## Definition of Done (Per Issue)

Each issue has:
- **Description**: What needs to be done
- **Design**: Detailed technical approach with code structure
- **Acceptance Criteria**: Checklist of requirements to mark complete
- **Time Estimate**: Realistic time allocation
- **Dependencies**: What must complete before starting
- **Priority**: 0 (highest) to 2 (lower)

## Issue Tracking Commands

```bash
# View all issues
bd list

# View ready work (no blockers)
bd ready

# View epic progress
bd epic status

# View specific issue details
bd show <issue-id>

# Update issue status
bd update <issue-id> --status in-progress
bd close <issue-id>

# View blocked issues
bd blocked
```

## Work Strategy

### Recommended Approach:

1. **Start with Phase 1 tasks in parallel** (if you have multiple people/sessions)
   - Tackle 7ty.4, 7ty.5, 7ty.6 simultaneously

2. **Focus on critical path** for solo work:
   - 7ty.5 → b2p.1 → b2p.2 → knw.1 → knw.3 → mev.1

3. **Verify continuously**:
   - Test each component immediately after building
   - Don't wait until the end to run tests

4. **Double-check critical components**:
   - The transaction logic in knw.1 is the heart of the system
   - The UNIQUE constraint in 7ty.5 is essential
   - The concurrency tests in mev.1 and mev.2 validate everything

### Time Allocation:

- **Development**: ~8-9 hours (following critical path)
- **Testing & Debugging**: 2-3 hours (buffer for issues)
- **Documentation**: 2-3 hours
- **Total**: 12-15 hours for complete implementation

With parallelization (team of 2-3):
- **Development**: ~3-4 hours
- **Testing & Debugging**: 1-2 hours
- **Documentation**: 1-2 hours
- **Total**: 5-8 hours

## Notes

- All issues created with `--force` flag due to prefix handling
- Issues use hierarchical parent-child relationships
- Dependencies tracked with `--deps` flag
- Estimates are in minutes for tracking
- Priority 0 = critical path, Priority 1 = important, Priority 2 = nice-to-have
