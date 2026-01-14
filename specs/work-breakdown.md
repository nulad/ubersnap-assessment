# Work Breakdown Summary

## Project Statistics

- **Total Issues Created**: 27
- **Epics**: 5
- **Tasks**: 21 (plus 1 from AGENTS.md)
- **Open Issues**: 27
- **Ready to Start**: 8 (no dependencies)
- **Blocked Issues**: 19 (waiting on dependencies)

## Epic Structure

### 1. Project Setup & Infrastructure (ubersnap-assessment-7ty)
**4 tasks | Priority: P0**

| ID | Task | Estimate | Priority | Blockers |
|----|------|----------|----------|----------|
| 7ty.4 | Initialize Go project structure | 30min | P0 | None ⚡ |
| 7ty.5 | Design database schema with constraints | 45min | P0 | None ⚡ |
| 7ty.6 | Setup Docker and Docker Compose | 60min | P0 | None ⚡ |
| 7ty.7 | Implement configuration management | 30min | P1 | 7ty.4 |

**Total: 165 minutes (2.75 hours)**

---

### 2. Database Layer Implementation (ubersnap-assessment-b2p)
**4 tasks | Priority: P0**

| ID | Task | Estimate | Priority | Blockers |
|----|------|----------|----------|----------|
| b2p.1 | Implement database connection with pooling | 60min | P0 | 7ty.5 |
| b2p.4 | Define data models and DTOs | 30min | P1 | 7ty.4 |
| b2p.2 | Implement Coupon Repository | 90min | P0 | b2p.1 |
| b2p.3 | Implement Claim Repository | 75min | P0 | b2p.1 |

**Total: 255 minutes (4.25 hours)**

**Key Deliverables:**
- PostgreSQL connection with retry logic and pooling
- Repository pattern implementation
- SELECT FOR UPDATE query in Coupon repository
- UNIQUE constraint violation detection in Claim repository

---

### 3. Business Logic & API Layer (ubersnap-assessment-knw)
**5 tasks | Priority: P0**

| ID | Task | Estimate | Priority | Blockers |
|----|------|----------|----------|----------|
| knw.1 | Implement Coupon Service with atomic transactions ⚠️ | 120min | P0 | b2p.2, b2p.3 |
| knw.5 | Implement main.go entry point | 60min | P0 | 7ty.4, b2p.1 |
| knw.2 | Implement POST /api/coupons | 45min | P1 | knw.1 |
| knw.3 | Implement POST /api/coupons/claim ⚠️ | 60min | P0 | knw.1 |
| knw.4 | Implement GET /api/coupons/{name} | 45min | P1 | knw.1 |

**Total: 330 minutes (5.5 hours)**

**Critical Components:**
- **knw.1**: Contains the atomic transaction logic with SELECT FOR UPDATE
- **knw.3**: The claim endpoint that handles concurrency

---

### 4. Testing & Quality Assurance (ubersnap-assessment-mev)
**4 tasks | Priority: P0**

| ID | Task | Estimate | Priority | Blockers |
|----|------|----------|----------|----------|
| mev.1 | Flash Sale test (50 requests, 5 stock) ⚠️ | 120min | P0 | knw.3 |
| mev.2 | Double Dip test (10 requests, same user) ⚠️ | 90min | P0 | knw.3 |
| mev.3 | Integration tests for all endpoints | 120min | P1 | knw.2, knw.3, knw.4 |
| mev.4 | Unit tests for service layer | 90min | P2 | knw.1 |

**Total: 420 minutes (7 hours)**

**Must-Pass Tests:**
- **mev.1**: Exactly 5 successful claims from 50 concurrent requests
- **mev.2**: Exactly 1 successful claim from 10 concurrent same-user requests

---

### 5. Documentation & Deployment (ubersnap-assessment-hs0)
**5 tasks | Priority: P1**

| ID | Task | Estimate | Priority | Blockers |
|----|------|----------|----------|----------|
| hs0.1 | Write comprehensive README.md | 90min | P0 | 7ty.6, mev.1, mev.2 |
| hs0.3 | Verify Docker deployment end-to-end | 60min | P0 | 7ty.6, knw.3, mev.1, mev.2 |
| hs0.5 | Final submission checklist | 45min | P0 | hs0.1, hs0.3, mev.1, mev.2 |
| hs0.2 | Create architecture documentation | 75min | P2 | knw.1 |
| hs0.4 | Create API examples and test scripts | 30min | P2 | knw.2, knw.3, knw.4 |

**Total: 300 minutes (5 hours)**

---

## Total Project Estimate

**Sequential (Following Critical Path): ~19.5 hours**
- Development: 12.5 hours
- Testing: 7 hours
- Buffer for debugging: +3-4 hours
- **Total: 15-16 hours of focused work**

**Parallel (With Team or Multi-Session): ~8-10 hours**
- Multiple tasks can be done simultaneously
- Significant time savings in Phases 1, 2, and 4

---

## Critical Success Factors

### Must-Have (Non-Negotiable):

1. ✅ **Database schema (7ty.5)**
   - UNIQUE(user_id, coupon_name) constraint
   - Separate tables for coupons and claims
   - Proper indexes and foreign keys

2. ✅ **Atomic transaction logic (knw.1)**
   - SELECT FOR UPDATE for row locking
   - Single transaction: lock → check → insert → update
   - Proper rollback on errors

3. ✅ **Claim endpoint (knw.3)**
   - Returns 409 for already claimed
   - Returns 400 for no stock
   - Handles concurrency correctly

4. ✅ **Flash Sale test passes (mev.1)**
   - 50 concurrent requests
   - 5 stock items
   - Result: exactly 5 success, 45 failures

5. ✅ **Double Dip test passes (mev.2)**
   - 10 concurrent requests
   - Same user_id
   - Result: exactly 1 success, 9 failures (409)

6. ✅ **Docker deployment works (7ty.6, hs0.3)**
   - docker-compose up --build
   - Single command to start everything
   - Services start in correct order

7. ✅ **README complete (hs0.1)**
   - Prerequisites listed
   - How to run instructions
   - How to test instructions
   - Architecture notes

### Nice-to-Have (But Not Required):

- Comprehensive unit tests (mev.4)
- Detailed architecture documentation (hs0.2)
- API example scripts (hs0.4)
- Perfect code formatting and comments

---

## Risk Mitigation

### High-Risk Areas:

1. **Transaction logic in service layer**
   - Risk: Race conditions, overselling
   - Mitigation: Test early with concurrency scenarios, review transaction boundaries

2. **Docker configuration**
   - Risk: Services start in wrong order, connection failures
   - Mitigation: Use healthchecks, depends_on with condition, test clean startup

3. **Concurrency tests**
   - Risk: Flaky tests, non-deterministic results
   - Mitigation: Use proper synchronization (WaitGroup), test multiple times

4. **PostgreSQL constraint handling**
   - Risk: Wrong error codes, not detecting UNIQUE violations
   - Mitigation: Test constraint violations explicitly, check pq error codes

---

## Work Sessions

### Session 1: Foundation (2-3 hours)
- Complete all Phase 1 tasks (7ty.4, 7ty.5, 7ty.6, 7ty.7)
- Verify Docker setup works
- Database schema created and tested
- **Deliverable**: Can run `docker-compose up` successfully

### Session 2: Core Implementation (3-4 hours)
- Complete Phase 2 (database layer)
- Complete knw.1 (service with transactions) - CRITICAL
- Complete knw.5 (main.go)
- **Deliverable**: Can create coupons via API

### Session 3: API Endpoints (2-3 hours)
- Complete knw.2, knw.3, knw.4
- Manual testing of all endpoints
- **Deliverable**: All three endpoints working

### Session 4: Critical Testing (2-3 hours)
- Complete mev.1 (Flash Sale test) - MUST PASS
- Complete mev.2 (Double Dip test) - MUST PASS
- Debug any concurrency issues
- **Deliverable**: Both critical tests passing 100%

### Session 5: Polish & Submit (2-3 hours)
- Complete hs0.1 (README)
- Complete hs0.3 (E2E verification)
- Complete hs0.5 (submission checklist)
- Final testing
- **Deliverable**: Ready to submit

---

## Quick Start Guide

### 1. Start with these three (parallel if possible):
```bash
bd update ubersnap-assessment-7ty.4 --status in-progress
bd update ubersnap-assessment-7ty.5 --status in-progress
bd update ubersnap-assessment-7ty.6 --status in-progress
```

### 2. Check what's ready at any time:
```bash
bd ready
```

### 3. View progress:
```bash
bd epic status
```

### 4. Close issues as you complete them:
```bash
bd close <issue-id>
```

### 5. Check for blockers:
```bash
bd blocked
```

---

## Definition of Done Checklist

Before closing any issue, verify:

- [ ] Code is written and tested
- [ ] All acceptance criteria met
- [ ] No syntax errors or compilation issues
- [ ] Manual testing performed (if applicable)
- [ ] Tests written and passing (if applicable)
- [ ] Code committed (if using git)
- [ ] Dependencies unblocked for next tasks

---

## Submission Preparation

Before submitting:

1. Run full test suite: `go test ./...`
2. Run concurrency tests multiple times: `go test ./tests -run "Flash|Double" -count=5`
3. Test clean deployment: `docker-compose down -v && docker-compose up --build`
4. Verify all API endpoints with curl
5. Review README instructions by following them exactly
6. Check repository for sensitive data (passwords, keys)
7. Verify .gitignore is correct
8. Push to GitHub
9. Test clone from GitHub and run deployment
10. Send submission email with repo link and CV

---

## Contact for Submission

**Email To:** rofie@ubersnap.com
**Email CC:** boonchin@ubersnap.com
**Subject:** Backend Engineer Test - [Your Name]
**Body:** Link to GitHub repo + any notes
**Attachment:** CV/Resume (PDF format)
