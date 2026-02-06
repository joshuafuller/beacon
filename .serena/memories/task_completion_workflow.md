# Task Completion Workflow for BEACON

## Before Starting Any Task

1. **Read CLAUDE.md** - Project guidelines and current status
2. **Check active specs** - `specs/006-mdns-responder/`, `specs/007-interface-specific-addressing/`
3. **Verify environment**:
   ```bash
   go version  # Should be 1.23.5
   export PATH=$PATH:$HOME/go_installation/go/bin
   ```

## During Development

### TDD Cycle (MANDATORY)
1. **RED**: Write test FIRST
   ```bash
   # Test should fail
   go test ./package -run TestNewFeature -v
   ```

2. **GREEN**: Write minimal code to pass test
   ```bash
   # Test should pass
   go test ./package -run TestNewFeature -v
   ```

3. **REFACTOR**: Clean up while keeping tests green
   ```bash
   go test ./package -v
   ```

### Layer Boundary Validation
```bash
# Ensure no violations
grep -rn "internal/network" querier/  # Should return 0 matches
```

### RFC Compliance
- Cite RFC sections in code: `// RFC 6762 §8.1`
- Reference RFC files: `RFC%20Docs/rfc6762.txt`

## After Completing Code Changes

### 1. Format Code (REQUIRED)
```bash
gofmt -w .
```

### 2. Static Analysis (REQUIRED)
```bash
go vet ./...
```

### 3. Run All Tests (REQUIRED)
```bash
make test
```

**Expected**: All tests PASS (100% pass rate)

### 4. Check Coverage (If Core Code Modified)
```bash
make test-coverage-report
```

**Target**: ≥80% overall, ≥85% for modified packages

### 5. Race Detector (REQUIRED)
```bash
make test-race
```

**Expected**: No race conditions detected

### 6. Semgrep Static Analysis (REQUIRED)
```bash
make semgrep-check
```

**Target**: 0 ERROR findings, ≤5 WARNING findings

### 7. Build Verification (REQUIRED)
```bash
go build ./...
```

**Expected**: Clean build, no errors

## Before Committing

### Pre-Commit Checklist
- [ ] Code formatted (`gofmt -w .`)
- [ ] All tests pass (`make test`)
- [ ] Race detector clean (`make test-race`)
- [ ] Semgrep clean (`make semgrep-check`)
- [ ] Coverage maintained/improved (`make test-coverage-report`)
- [ ] Build succeeds (`go build ./...`)
- [ ] RFCs cited where applicable
- [ ] Godoc comments added for exports
- [ ] Layer boundaries respected

### Git Workflow
```bash
# Conventional commit messages
git commit -m "feat: add service conflict resolution"
git commit -m "fix: correct DNS name encoding per RFC 1035 §3.1"
git commit -m "test: add coverage for handleQuery edge cases"

# Reference spec/task numbers
git commit -m "feat(responder): implement known-answer suppression (T096)"
```

## Task-Specific Completion Criteria

### Adding Tests (Coverage Work)
- [ ] Test written FIRST (RED phase)
- [ ] Test passes (GREEN phase)
- [ ] Code refactored (REFACTOR phase)
- [ ] Coverage increased for target function
- [ ] All existing tests still pass
- [ ] No new race conditions

### Fixing Semgrep Issues
- [ ] Issue reproduced: `make semgrep-check`
- [ ] Fix applied OR suppression with justification
- [ ] Semgrep clean: 0 new ERROR findings
- [ ] Tests still pass
- [ ] Documentation updated if behavior changed

### Adding Documentation
- [ ] Document created/updated
- [ ] Links verified (no broken references)
- [ ] Code examples tested (if applicable)
- [ ] Spell-checked
- [ ] Consistent with existing docs style

### Implementing TODOs
- [ ] TODO understood and documented
- [ ] Tests written FIRST
- [ ] Implementation complete
- [ ] TODO comment removed or updated
- [ ] All tests pass
- [ ] Behavior matches RFC if protocol-related

## When Task is Complete

### Update Task Tracking
If using Ralph:
- [ ] Mark task `[x]` in `@fix_plan.md`
- [ ] Log completion in `logs/ralph.log`

If using specs:
- [ ] Mark task complete in `specs/*/tasks.md`
- [ ] Update milestone progress

### Documentation
- [ ] Update relevant ADRs if architectural
- [ ] Update CLAUDE.md if workflow changed
- [ ] Update examples if API changed

### Validation
```bash
# Final validation before considering task done
make test-race && make semgrep-check && make test-coverage-report
```

**Success**: All commands exit 0, coverage ≥80%

## Current Status (Ralph Active)

Ralph is autonomously working through:
1. **P0**: Coverage increase (handleQuery, collectResponses, etc.)
2. **P1**: Semgrep analysis and fixes
3. **P2**: Public API testing
4. **P3**: Documentation polish
5. **P4**: TODO cleanup

Ralph will automatically follow this workflow for each task.
