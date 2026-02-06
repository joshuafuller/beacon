# Quick Reference: Implementation Guide

**Date**: 2026-01-06
**Feature**: 008-documentation-production-polish
**Phase**: Phase 1 - Design & Templates

## Purpose

Fast reference for Ralph or contributors implementing tasks. Each section provides step-by-step checklists that can be executed directly without consulting other documentation.

---

## README Update Checklist (T001-T006)

### Task: Update README.md with Accurate Stats

**File**: `/home/user/development/beacon/README.md`

**Changes Required**:
1. Line ~30: RFC compliance stat
2. Line ~35: Test coverage stat
3. Add RFC compliance badges (after title, before description)
4. Add link to compliance certificate

**Step-by-Step**:

```bash
# 1. Open README.md
code README.md  # or your editor

# 2. Find and replace compliance stat (search for "72.2%")
# OLD: 72.2% RFC compliance
# NEW: 97.9% RFC 6762 compliance, 96.9% RFC 6763 compliance

# 3. Find and replace coverage stat (search for "81.3%")
# OLD: 81.3% test coverage
# NEW: 68.6% test coverage

# 4. Add badges after title (after "# Beacon\n\n" line):
[![RFC 6762 P0](https://img.shields.io/badge/RFC%206762%20P0-100%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC_COMPLIANCE_CERTIFICATE_v1.0.md)
[![RFC 6762 Overall](https://img.shields.io/badge/RFC%206762-97.9%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC_COMPLIANCE_CERTIFICATE_v1.0.md)
[![RFC 6763 P0](https://img.shields.io/badge/RFC%206763%20P0-100%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC6763_KEY_REQUIREMENTS.md)
[![RFC 6763 Overall](https://img.shields.io/badge/RFC%206763-96.9%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC6763_KEY_REQUIREMENTS.md)

# 5. Add compliance note in Features section:
Beacon achieves **100% compliance with all mandatory (MUST) requirements** for both RFC 6762 (Multicast DNS) and RFC 6763 (DNS-Based Service Discovery). The overall percentages (97.9%/96.9%) include optional (SHOULD/MAY) features. See [RFC_COMPLIANCE_CERTIFICATE_v1.0.md](./RFC_COMPLIANCE_CERTIFICATE_v1.0.md) for details.

# 6. Test README rendering
# Option A: GitHub preview
git add README.md
git commit -m "docs: Update README with accurate RFC compliance stats"
git push origin 008-documentation-production-polish
# View on GitHub to verify badges render

# Option B: Local markdown preview
grip README.md  # if grip installed
```

**Verification**:
- [ ] Badges display correctly (green, 100% and 97.9%/96.9%)
- [ ] Badge links navigate to compliance certificates
- [ ] Stats updated to 97.9%, 96.9%, 68.6%
- [ ] Compliance note added explaining 100% P0 vs overall

---

## Example Creation Steps (T007-T031)

### General Pattern (All Examples)

**Template**: See `data-model.md` for full templates

**Step-by-Step**:

```bash
# 1. Create directory structure
cd /home/user/development/beacon
mkdir -p examples/[category]/[example-name]
cd examples/[category]/[example-name]

# 2. Create go.mod
cat > go.mod <<'EOF'
module github.com/joshuafuller/beacon/examples/[category]/[example-name]

go 1.21

require github.com/joshuafuller/beacon v0.0.0

// Use local Beacon code instead of remote
replace github.com/joshuafuller/beacon => ../../..
EOF

# 3. Create Makefile
cat > Makefile <<'EOF'
# [Example Name] - Makefile
# Category: [Basic/Intermediate/Advanced]

.PHONY: run build test clean help

BINARY := $(notdir $(CURDIR))
BUILD_DIR := bin

run:
	@echo "Running [example-name]..."
	@go run main.go

build:
	@echo "Building [example-name]..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY) main.go
	@echo "Binary created at $(BUILD_DIR)/$(BINARY)"

test:
	@echo "Running tests..."
	@go test -v .

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)/
	@echo "Clean complete"

help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
EOF

# 4. Create main.go (see example-specific templates below)
code main.go

# 5. Create README.md (see README template in data-model.md)
code README.md

# 6. Download dependencies
go mod download

# 7. Compile and test
make build
make run

# 8. Verify example works
# (example-specific verification steps)

# 9. Commit
git add .
git commit -m "examples: Add [category]/[example-name]"
```

---

### Example-Specific: Hello Responder (T007-T011)

**Path**: `examples/basic/hello-responder/`
**Contract**: `contracts/examples-basic.md` (Example 1)

**main.go Template**:
```go
// Package main demonstrates minimal service registration with Beacon.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshuafuller/beacon/responder"
)

func main() {
	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create responder
	r, err := responder.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create responder: %v", err)
	}
	defer r.Close()

	// Define service
	svc := &responder.Service{
		Instance: "Hello World",
		Service:  "_http._tcp",
		Domain:   "local",
		Port:     8080,
	}

	// Register service
	if err := r.Register(ctx, svc); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	fmt.Printf("Service registered: %s.%s.%s\n", svc.Instance, svc.Service, svc.Domain)
	fmt.Println("Press Ctrl+C to exit")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}
```

**Verification Steps**:
```bash
# 1. Run example
cd examples/basic/hello-responder
make run

# 2. In another terminal, verify service is visible:
# macOS:
dns-sd -B _http._tcp

# Linux:
avahi-browse -t _http._tcp

# Expected: "Hello World._http._tcp.local" appears in list

# 3. Press Ctrl+C in first terminal
# Expected: Graceful shutdown message, service disappears from list
```

---

### Example-Specific: Error Handling (T012-T016)

**Path**: `examples/basic/error-handling/`
**Contract**: `contracts/examples-basic.md` (Example 2)

**main.go Structure** (~80 lines):
1. Scenario 1: ValidationError (empty instance name)
2. Scenario 2: NetworkError (port already in use)
3. Scenario 3: Context cancellation
4. Scenario 4: Proper production error handling

**Key Code Patterns**:
```go
// Scenario 1: ValidationError
svc := &responder.Service{
	Instance: "", // Invalid: empty
	Service:  "_http._tcp",
	Domain:   "local",
	Port:     8080,
}
err := r.Register(ctx, svc)
if err != nil {
	var valErr *errors.ValidationError
	if errors.As(err, &valErr) {
		fmt.Printf("ValidationError: %v\n", valErr)
	}
}

// Scenario 2: NetworkError (simulate port conflict)
r1, _ := responder.New(ctx, responder.WithPort(5353))
r2, err := responder.New(ctx, responder.WithPort(5353)) // Conflict!
if err != nil {
	var netErr *errors.NetworkError
	if errors.As(err, &netErr) {
		fmt.Printf("NetworkError: %v\n", netErr)
	}
}
```

---

### Example-Specific: Graceful Shutdown (T017-T021)

**Path**: `examples/basic/graceful-shutdown/`
**Contract**: `contracts/examples-basic.md` (Example 3)

**Key Addition to main.go**:
```go
// In main(), before <-sigChan:
defer func() {
	fmt.Println("Shutting down gracefully...")

	// Unregister service (sends goodbye packet)
	if err := r.Unregister(ctx, svc.Instance); err != nil {
		log.Printf("Error unregistering: %v", err)
	}

	fmt.Println("Goodbye packet sent")

	// Wait for goodbye packet to propagate (RFC 6762 §10.1)
	time.Sleep(250 * time.Millisecond)

	fmt.Println("Shutdown complete")
}()
```

**Verification with Wireshark**:
```bash
# 1. Start Wireshark capture on mDNS traffic
sudo wireshark &
# Filter: udp.port == 5353

# 2. Run example
make run

# Expected in Wireshark: PTR announcements

# 3. Press Ctrl+C

# Expected in Wireshark: PTR record with TTL=0 (goodbye packet)
```

---

### Example-Specific: Multi-Service (T022-T026)

**Path**: `examples/basic/multi-service/`
**Contract**: `contracts/examples-basic.md` (Example 4)

**Key Code Pattern**:
```go
// Define 3 services
services := []*responder.Service{
	{Instance: "Multi-Service Demo", Service: "_http._tcp", Port: 8080},
	{Instance: "Multi-Service Demo", Service: "_ssh._tcp", Port: 22},
	{Instance: "Multi-Service Demo", Service: "_myapp._tcp", Port: 9000},
}

// Register each
for i, svc := range services {
	if err := r.Register(ctx, svc); err != nil {
		log.Fatalf("Failed to register service %d: %v", i+1, err)
	}
	fmt.Printf("Registered service %d/3: %s.%s\n", i+1, svc.Instance, svc.Service)
}

// On shutdown, unregister all
for _, svc := range services {
	r.Unregister(ctx, svc.Instance)
}
```

---

### Example-Specific: Browser (T027-T031)

**Path**: `examples/basic/browser/`
**Contract**: `contracts/examples-basic.md` (Example 5)

**Key Code Pattern**:
```go
import "github.com/joshuafuller/beacon/querier"

func main() {
	ctx := context.Background()

	// Create querier
	q, err := querier.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create querier: %v", err)
	}
	defer q.Close()

	fmt.Println("Browsing for available services...")

	// Query for service enumeration (RFC 6763 §9)
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	results, err := q.Query(queryCtx, "_services._dns-sd._udp.local")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	// Parse PTR records (each points to a service type)
	services := make(map[string]bool) // Deduplicate
	for _, result := range results {
		if result.Type == "PTR" {
			services[result.PTR] = true
		}
	}

	// Print results
	for service := range services {
		fmt.Printf("Found service: %s\n", service)
	}
	fmt.Printf("Total services found: %d\n", len(services))
}
```

---

## Godoc Example Pattern (T037-T044)

### Function Naming Convention

**Pattern**: `Example<Type>_<Method>[_<suffix>]`

**Examples**:
- `ExampleResponder_Register` - Demonstrates `Responder.Register()`
- `ExampleResponder_Register_multiService` - Second example for `Responder.Register()`
- `ExampleService_Validate` - Demonstrates `Service.Validate()`
- `Example` - Package-level example

### Template

**File**: `responder/responder_test.go` (add to existing file)

```go
func ExampleResponder_Register() {
	// Create context
	ctx := context.Background()

	// Create responder
	r, err := responder.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Define service
	svc := &responder.Service{
		Instance: "My Service",
		Service:  "_http._tcp",
		Domain:   "local",
		Port:     8080,
		TXT: map[string]string{
			"version": "1.0",
		},
	}

	// Register service
	if err := r.Register(ctx, svc); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Service registered: %s.%s.%s\n", svc.Instance, svc.Service, svc.Domain)
	// Output:
	// Service registered: My Service._http._tcp.local
}
```

**Key Rules**:
1. **Output validation**: `// Output:` comment must match EXACTLY (whitespace-sensitive)
2. **No timestamps**: Use predictable output (`fmt.Printf`, not `log.Printf`)
3. **Self-contained**: No external dependencies
4. **Error handling**: Show proper error handling (`log.Fatal` for examples)

### Testing Godoc Examples

```bash
# Run all examples
go test -v -run Example ./responder ./querier

# Run specific example
go test -v -run ExampleResponder_Register ./responder

# Expected output: PASS (if // Output: matches actual output)
```

---

## Hugo Site Deployment (T045-T056)

### Quick Start (T045-T046)

```bash
# 1. Create Hugo site structure
cd /home/user/development/beacon
mkdir -p docs

# 2. Create config.toml
cat > docs/config.toml <<'EOF'
baseURL = "https://joshuafuller.github.io/beacon/"
title = "Beacon - mDNS for Go"
theme = "docsy"

[params]
description = "Production-ready mDNS library for Go with RFC 6762/6763 compliance"
github_repo = "https://github.com/joshuafuller/beacon"
github_branch = "main"

[params.ui]
sidebar_menu_compact = true
breadcrumb_disable = false

[markup]
[markup.goldmark.renderer]
unsafe = true  # Allow raw HTML
EOF

# 3. Add Docsy theme as git submodule
cd docs
git submodule add https://github.com/google/docsy.git themes/docsy
git submodule update --init --recursive

# 4. Pin Docsy version
cd themes/docsy
git checkout v0.10.0
cd ../..

# 5. Create content directories
mkdir -p content/{getting-started,guides,examples,reference,architecture}
mkdir -p static/images

# 6. Create landing page
cat > content/_index.md <<'EOF'
---
title: "Beacon"
description: "Production-ready mDNS library for Go"
---

# Beacon

Production-ready Multicast DNS (mDNS) library for Go with 100% RFC compliance.

## Features

- ✅ 100% RFC 6762 (mDNS) P0 compliance
- ✅ 100% RFC 6763 (DNS-SD) P0 compliance
- ✅ Production-ready with comprehensive testing
- ✅ Zero external dependencies (stdlib only)

## Quick Start

```bash
go get github.com/joshuafuller/beacon
```
EOF

# 7. Test Hugo build
hugo serve

# Expected: Site available at http://localhost:1313
```

### GitHub Actions Deployment (T055-T056)

**File**: `.github/workflows/docs.yml`

```yaml
name: Deploy Hugo Docs

on:
  push:
    branches:
      - main
    paths:
      - 'docs/**'
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          submodules: recursive  # Fetch Docsy theme

      - name: Setup Hugo
        uses: peaceiris/actions-hugo@v2
        with:
          hugo-version: '0.120.0'
          extended: true

      - name: Build
        run: |
          cd docs
          hugo --minify

      - name: Deploy
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs/public
          publish_branch: gh-pages
```

**Testing Deployment**:
```bash
# 1. Push to branch
git add .github/workflows/docs.yml docs/
git commit -m "docs: Add Hugo site with GitHub Actions deployment"
git push origin 008-documentation-production-polish

# 2. Verify Actions run
# Go to: https://github.com/joshuafuller/beacon/actions

# 3. Enable GitHub Pages (if not already)
# Repo Settings → Pages → Source: gh-pages branch, / (root)

# 4. Verify site live
# https://joshuafuller.github.io/beacon/
```

---

## Deployment Guide Pattern (T032-T036)

### General Structure

See `contracts/deployment-guides.md` for complete specifications.

**Quick Checklist**:
```bash
# 1. Create guide file
code docs/deployment/[guide-name].md

# 2. Follow template structure:
# - Overview (1 paragraph)
# - Quick Start (≤5 commands)
# - Detailed sections
# - Troubleshooting
# - Next Steps

# 3. Test all commands on Linux AND macOS
# (Don't document untested commands!)

# 4. Add cross-links to other guides

# 5. Commit
git add docs/deployment/[guide-name].md
git commit -m "docs: Add [guide-name] deployment guide"
```

---

## Task Execution Checklist

For each task, follow this pattern:

```bash
# 1. Mark task in PROMPT.md as in-progress
# (Update [ ] to [x] in progress tracking section)

# 2. Execute task (see task-specific steps above)

# 3. Test/verify output
# (See example-specific verification steps)

# 4. Commit immediately after task completion
git add [files]
git commit -m "[type]: [description]"

# 5. Mark task complete in PROMPT.md

# 6. Move to next task
```

**Important**: One file edit per commit (Ralph requirement)

---

## Common Commands Reference

### mDNS Discovery Tools

```bash
# macOS - Bonjour (dns-sd)
dns-sd -B _http._tcp           # Browse for HTTP services
dns-sd -L "Service" _http._tcp # Lookup specific service details

# Linux - Avahi
avahi-browse -t _http._tcp     # Browse for HTTP services
avahi-browse -t _services._dns-sd._udp  # List all service types

# Windows
# (No built-in tool - use third-party Bonjour browser)
```

### Network Diagnostics

```bash
# Check port 5353
sudo netstat -ulnp | grep 5353
sudo lsof -i :5353

# Test multicast
ping -c 3 224.0.0.251

# Capture mDNS traffic
sudo tcpdump -i any udp port 5353

# Wireshark filter
udp.port == 5353
```

### Go Testing

```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestName ./package

# Run examples
go test -v -run Example ./responder

# With coverage
go test -cover ./...
```

---

**Status**: Quickstart reference complete, ready for task execution
