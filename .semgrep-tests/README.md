# Semgrep Rule Testing

This directory contains test files to validate that Semgrep rules correctly detect violations.

## Purpose

The test files intentionally contain violations of Constitution principles, F-Spec requirements, and RFC compliance rules to ensure our Semgrep configuration catches them.

## Test Files

### `test_violations.go`
Tests for general concurrency, error handling, and resource management rules:
- Timer/Ticker leaks (F-4 REQ-F4-7)
- WaitGroup missing Done() (F-4 REQ-F4-1)
- Unbuffered result channels (F-4)
- Mutex without defer unlock (F-4)
- Error message punctuation (F-3)
- Error log and return (F-3 RULE-3)
- File missing defer close (F-7 REQ-F7-5)
- Hardcoded mDNS constants (RFC 6762)
- Context not first param (F-4 RULE-1)
- Context not checked in loop (F-4)
- Error swallowing (Constitution IV)
- Error capitalization (Constitution VIII)

### `internal/message/security_test.go`
Tests for F-11 Security Architecture rules:
- Unsafe package in parser (F-11 REQ-F11-3)
- Panic on network input (F-11 REQ-F11-3)

### `internal/test_package/bad_import.go`
Tests for F-2 Package Structure rules:
- Internal packages importing public packages (F-2 RULE-2)

### `examples/bad_example.go`
Tests for F-2 Package Structure rules:
- Examples importing internal packages (F-2 PRIN-5)

## Running Tests

### Validate Configuration Syntax

```bash
semgrep --config=.semgrep.yml --validate
```

Expected output: `Configuration is valid - found 0 configuration error(s), and 21 rule(s).`

### Run Tests on Test Files

```bash
semgrep --config=.semgrep.yml .semgrep-tests/ --no-git-ignore
```

Expected: **16 findings** across the test files

### Run Tests on Actual Codebase

```bash
# Scan entire codebase
semgrep --config=.semgrep.yml .

# Scan specific directory
semgrep --config=.semgrep.yml ./internal/

# Only show ERROR severity
semgrep --config=.semgrep.yml --severity ERROR .

# JSON output for CI
semgrep --config=.semgrep.yml --json . > semgrep-results.json
```

## Expected Findings in Test Files

The test files should trigger these rules:

| Rule ID | Severity | Count | File |
|---------|----------|-------|------|
| `beacon-timer-leak` | ERROR | 1 | `test_violations.go:22` |
| `beacon-ticker-leak` | ERROR | 1 | `test_violations.go:39` |
| `beacon-waitgroup-missing-done` | ERROR | 1 | `test_violations.go:62-66` |
| `beacon-mutex-defer-unlock` | ERROR | 1 | `test_violations.go:114` |
| `beacon-error-message-punctuation` | INFO | 2 | `test_violations.go:133,143` |
| `beacon-error-log-and-return` | WARNING | 1 | `test_violations.go:152-157` |
| `beacon-file-missing-defer-close` | ERROR | 1 | `test_violations.go:179-182` |
| `beacon-hardcoded-mdns-port` | WARNING | 1 | `test_violations.go:204` |
| `beacon-hardcoded-multicast-address` | WARNING | 1 | `test_violations.go:209` |
| `beacon-context-not-first-param` | WARNING | 1 | `test_violations.go:221` |
| `beacon-error-swallowing` | WARNING | 1 | `test_violations.go:266` |
| `beacon-error-capitalization` | INFO | 1 | `test_violations.go:279` |
| `beacon-unsafe-in-parser` | ERROR | 1 | `internal/message/security_test.go:11` |
| `beacon-panic-on-network-input` | ERROR | 1 | `internal/message/security_test.go:23` |
| `beacon-internal-imports-public` | ERROR | 1 | `internal/test_package/bad_import.go:11` |
| `beacon-test-imports-internal` | ERROR | 1 | `examples/bad_example.go:11` |

**Total: 16 findings**

## CI Integration

### GitHub Actions

Add to `.github/workflows/semgrep.yml`:

```yaml
name: Semgrep

on:
  pull_request:
  push:
    branches: [main]

jobs:
  semgrep:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Semgrep
        run: pip install semgrep

      - name: Run Semgrep
        run: |
          semgrep --config=.semgrep.yml --error --json . > semgrep-results.json

      - name: Upload Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: semgrep-results
          path: semgrep-results.json
```

### Pre-commit Hook

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Running Semgrep checks..."
semgrep --config=.semgrep.yml --error .
if [ $? -ne 0 ]; then
    echo "❌ Semgrep found errors. Please fix before committing."
    exit 1
fi
```

## Rule Categories

### High Priority (ERROR)
Rules that catch critical bugs or security issues:
- Timer/Ticker leaks
- WaitGroup missing Done()
- Mutex without defer unlock
- File missing defer close
- Unsafe in parser
- Panic on network input
- Layer boundary violations

### Medium Priority (WARNING)
Rules that enforce best practices:
- Context parameter order
- Context not checked in loop
- Error log and return
- Hardcoded constants

### Low Priority (INFO)
Rules for code style:
- Error message punctuation
- Error capitalization

## Troubleshooting

### False Positives

If a rule triggers incorrectly, you can:

1. **Suppress inline** (use sparingly):
   ```go
   // nosemgrep: beacon-timer-leak
   timer := time.NewTimer(duration)
   ```

2. **Exclude files** in `.semgrepignore`:
   ```
   # Ignore generated code
   **/generated/**

   # Ignore vendor
   vendor/
   ```

3. **Improve the rule** in `.semgrep.yml`:
   - Add `pattern-not` to exclude valid cases
   - Refine metavariable regex
   - Adjust path filters

### No Findings

If rules aren't triggering:

1. **Check file paths**: Ensure test files are in expected locations
2. **Validate syntax**: Run `semgrep --validate`
3. **Check language**: Ensure `languages: [go]` is set
4. **Test patterns**: Use `semgrep --test` to debug individual patterns

## Maintenance

When adding new rules:

1. Add rule to `.semgrep.yml`
2. Validate syntax: `semgrep --config=.semgrep.yml --validate`
3. Add test case to this directory
4. Run tests and verify detection
5. Document in this README

When modifying existing rules:

1. Check test files still trigger correctly
2. Run on codebase to check for regressions
3. Update documentation if behavior changes

## Reference

- **Semgrep Documentation**: https://semgrep.dev/docs/
- **Semgrep Rule Syntax**: https://semgrep.dev/docs/writing-rules/rule-syntax/
- **Constitution**: `.specify/memory/constitution.md`
- **F-Specs**: `.specify/specs/F-*.md`
- **RFC 6762**: `RFC%20Docs/RFC-6762-Multicast-DNS.txt`
