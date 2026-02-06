# 🎯 RALPH TASK: P1 - Hugo Documentation Site & Intermediate Examples

## COMPLETION PROMISE

When ALL P1 tasks (T056-T110) are complete, output:

```
<promise>DONE</promise>
```

## P0 STATUS

✅ **P0 IS COMPLETE** (T001-T055)
- 49 tasks complete
- 6 tasks skipped (valid reasons)
- See P0_TASK_AUDIT_REPORT.md for details

---

## P1 SCOPE

**P1 = Hugo Documentation Site (US5) + Intermediate Examples (US6)**

**Tasks**: T056-T110 (55 tasks total)
- US5: Hugo site (T056-T073) - 18 tasks
- US6: Intermediate examples (T074-T110) - 37 tasks

---

## SUCCESS CRITERIA

### Hugo Site (US5)
- [ ] Hugo + Docsy theme configured (config.toml, submodule)
- [ ] 5 content sections created (getting-started, guides, examples, reference, architecture)
- [ ] Existing docs migrated to Hugo format with front matter
- [ ] GitHub Actions workflow deploys to joshuafuller.github.io/beacon
- [ ] Site builds and serves locally: `cd docs && hugo serve`

### Intermediate Examples (US6)
- [ ] Example 6: web-server (HTTP + mDNS integration)
- [ ] Example 7: service-updates (dynamic TXT records)
- [ ] Example 8: multi-interface-bridge (IoT WiFi ↔ Ethernet)
- [ ] Example 9: query-filter (advanced browsing)
- [ ] Example 10: conflict-handling (name conflict resolution)
- [ ] All examples compile: `cd examples/intermediate/* && go build`

---

## PROCESS

1. **BEFORE starting work**: Check if P1 is already complete:
   ```bash
   pending=$(grep -c "^- \[ \] T0\(5[6-9]\|[6-9][0-9]\|1[0-1][0-9]\)" specs/008-documentation-production-polish/tasks.md)
   if [ $pending -eq 0 ]; then
       echo "<promise>DONE</promise>"
       exit 0
   fi
   ```

2. **Hugo Dependency Check** (T056-T073 require Hugo):
   ```bash
   if ! command -v hugo >/dev/null 2>&1; then
       echo "⚠️  Hugo not installed - SKIPPING US5 tasks (T056-T073)"
       # Mark all US5 tasks as SKIP
       sed -i 's/^- \[ \] T0\(5[6-9]\|6[0-9]\|7[0-3]\) .*\[US5\]/- [SKIP] &/' tasks.md
       # Continue to US6 (T074+)
   fi
   ```

3. Work through incomplete tasks **in order** (T056 → T110)

4. After completing EACH task, mark it complete:
   ```bash
   sed -i 's/^- \[ \] TXXX /- [x] TXXX /' specs/008-documentation-production-polish/tasks.md
   ```

5. If a task CANNOT be completed:
   - Mark it [SKIP] with reason: `sed -i 's/^- \[ \] TXXX /- [SKIP] TXXX - NOTE: reason/' tasks.md`
   - Move to next task immediately

6. **AFTER each task**: Check if P1 is done:
   ```bash
   if [ $(grep -c "^- \[ \] T0\(5[6-9]\|[6-9][0-9]\|1[0-1][0-9]\)" tasks.md) -eq 0 ]; then
       echo "<promise>DONE</promise>"
       exit 0
   fi
   ```

7. **CRITICAL**: Do NOT proceed to T111+ (P2 tasks) without user approval

---

## TASK REFERENCE

See: `specs/008-documentation-production-polish/tasks.md`

**Phase 7: User Story 5 - Hugo Site (T056-T073)** - 18 tasks
**Phase 8: User Story 6 - Intermediate Examples (T074-T110)** - 37 tasks

---

## DEPENDENCIES

- **Hugo Extended 0.120+**: Required for US5 (install via `sudo apt install hugo` or download from https://github.com/gohugoio/hugo/releases)
- **Go 1.21+**: Required for US6 (examples compilation)
- **Docker**: Optional for testing multi-interface bridge example

---

## SKIP POLICY

**US5 (Hugo Site)**: If Hugo not installed, SKIP ALL T056-T073 and continue to US6
**US6 (Examples)**: If Go not installed, SKIP test tasks (T078, T083, T088, T093, T098, T103, T108) but CREATE all files

---

## NOTES

- P0 deliverables are in place (examples/basic/, docs/deployment/)
- Godoc examples already created (responder/example_test.go, querier/example_test.go)
- README already updated with RFC compliance badges
- Focus P1 on: Hugo site structure + intermediate examples

---

**Start working on T056 (or first pending P1 task)**

