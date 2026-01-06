# Tasks: Documentation & Production Polish

**Input**: Design documents from `/home/user/development/beacon/specs/008-documentation-production-polish/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Note**: This is a documentation project. Tests are not applicable - verification is via manual testing (examples compile/run, guides work, links resolve).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Documentation project: `examples/`, `docs/`, `responder/`, `querier/` at repository root
- All paths are relative to `/home/user/development/beacon/`

---

## Phase 1: Setup (Minimal Infrastructure)

**Purpose**: No special setup required - working with existing Beacon repository

- [ ] T001 Verify Go 1.21+ installed and Beacon project compiles with `make test`
- [ ] T002 Verify Hugo extended 0.120+ available for documentation site (P1 task)
- [ ] T003 [P] Verify Docker and docker-compose available for deployment examples

**Checkpoint**: Environment ready - documentation work can begin

---

## Phase 2: Foundational (No Blocking Prerequisites)

**Purpose**: Documentation projects have no foundational blockers - skip to user stories

**Note**: All user stories can begin immediately after Phase 1

---

## Phase 3: User Story 1 - README Accuracy & Discoverability (Priority: P0) 🎯 MVP CORE

**Goal**: Update README with accurate RFC compliance stats (100% P0, 97.9%/96.9% overall) and add compliance badges

**Independent Test**: View README.md on GitHub, verify badges render correctly with accurate percentages and link to compliance certificates

**Template**: quickstart.md (README Update Checklist)
**Contract**: spec.md (FR-001 through FR-005)

### Implementation for User Story 1

- [ ] T004 [US1] Update README.md line ~30: Change "72.2% RFC compliance" to "97.9% RFC 6762 compliance, 96.9% RFC 6763 compliance" in README.md
- [ ] T005 [US1] Update README.md line ~35: Change "81.3% test coverage" to "68.6% test coverage" in README.md
- [ ] T006 [P] [US1] Add RFC 6762 P0 badge after title in README.md: `[![RFC 6762 P0](https://img.shields.io/badge/RFC%206762%20P0-100%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC_COMPLIANCE_CERTIFICATE_v1.0.md)`
- [ ] T007 [P] [US1] Add RFC 6762 Overall badge after title in README.md: `[![RFC 6762 Overall](https://img.shields.io/badge/RFC%206762-97.9%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC_COMPLIANCE_CERTIFICATE_v1.0.md)`
- [ ] T008 [P] [US1] Add RFC 6763 P0 badge after title in README.md: `[![RFC 6763 P0](https://img.shields.io/badge/RFC%206763%20P0-100%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC6763_KEY_REQUIREMENTS.md)`
- [ ] T009 [P] [US1] Add RFC 6763 Overall badge after title in README.md: `[![RFC 6763 Overall](https://img.shields.io/badge/RFC%206763-96.9%25-brightgreen?style=flat-square&logo=checkmarx)](./RFC6763_KEY_REQUIREMENTS.md)`
- [ ] T010 [US1] Add compliance note in Features section of README.md: "Beacon achieves **100% compliance with all mandatory (MUST) requirements** for both RFC 6762 and RFC 6763. The overall percentages (97.9%/96.9%) include optional (SHOULD/MAY) features. See [RFC_COMPLIANCE_CERTIFICATE_v1.0.md](./RFC_COMPLIANCE_CERTIFICATE_v1.0.md) for details."
- [ ] T011 [US1] Test README quick start: Run `go get github.com/joshuafuller/beacon && cd examples/basic/hello-querier && go run main.go` and verify success

**Checkpoint**: README displays accurate compliance stats with working badges - first impression fixed

---

## Phase 4: User Story 2 - Basic Example Availability (Priority: P0) 🎯 MVP CORE

**Goal**: Provide 5 basic examples demonstrating fundamental Beacon responder usage

**Independent Test**: Each example compiles with `go build main.go` and runs with `go run main.go`, producing expected output within 30 seconds

**Templates**: data-model.md (Example Template Structure, main.go Template, go.mod Template, Makefile Template)
**Contracts**: contracts/examples-basic.md (Examples 1-5)
**Quick Reference**: quickstart.md (Example Creation Steps, Example-Specific sections)

### Example 1: Hello Responder

- [ ] T012 [P] [US2] Create examples/basic/hello-responder/README.md using template from data-model.md (README.md Template), following contract from contracts/examples-basic.md (Example 1)
- [ ] T013 [P] [US2] Create examples/basic/hello-responder/main.go (~50 lines) per quickstart.md (Hello Responder section) - minimal service registration
- [ ] T014 [P] [US2] Create examples/basic/hello-responder/go.mod with replace directive per data-model.md (go.mod Template)
- [ ] T015 [P] [US2] Create examples/basic/hello-responder/Makefile with standard targets per data-model.md (Makefile Template)
- [ ] T016 [US2] Test hello-responder: Run `cd examples/basic/hello-responder && go run main.go`, verify "Service registered: Hello World._http._tcp.local" output, service visible via `dns-sd -B _http._tcp` (macOS) or `avahi-browse -t _http._tcp` (Linux)

### Example 2: Error Handling

- [ ] T017 [P] [US2] Create examples/basic/error-handling/README.md demonstrating all error types per contracts/examples-basic.md (Example 2)
- [ ] T018 [P] [US2] Create examples/basic/error-handling/main.go (~80 lines) showing NetworkError, ValidationError, context cancellation per quickstart.md (Error Handling section)
- [ ] T019 [P] [US2] Create examples/basic/error-handling/go.mod per data-model.md template
- [ ] T020 [P] [US2] Create examples/basic/error-handling/Makefile per data-model.md template
- [ ] T021 [US2] Test error-handling: Run `cd examples/basic/error-handling && go run main.go`, verify all 4 error scenarios trigger correctly

### Example 3: Graceful Shutdown

- [ ] T022 [P] [US2] Create examples/basic/graceful-shutdown/README.md showing clean termination patterns per contracts/examples-basic.md (Example 3)
- [ ] T023 [P] [US2] Create examples/basic/graceful-shutdown/main.go (~60 lines) with signal handling, goodbye packets per quickstart.md (Graceful Shutdown section)
- [ ] T024 [P] [US2] Create examples/basic/graceful-shutdown/go.mod per data-model.md template
- [ ] T025 [P] [US2] Create examples/basic/graceful-shutdown/Makefile per data-model.md template
- [ ] T026 [US2] Test graceful-shutdown: Run `cd examples/basic/graceful-shutdown && go run main.go`, press Ctrl+C, verify goodbye packet sent and service disappears from discovery within 1 second

### Example 4: Multi-Service

- [ ] T027 [P] [US2] Create examples/basic/multi-service/README.md for multiple service registration per contracts/examples-basic.md (Example 4)
- [ ] T028 [P] [US2] Create examples/basic/multi-service/main.go (~70 lines) registering 3+ services per quickstart.md (Multi-Service section)
- [ ] T029 [P] [US2] Create examples/basic/multi-service/go.mod per data-model.md template
- [ ] T030 [P] [US2] Create examples/basic/multi-service/Makefile per data-model.md template
- [ ] T031 [US2] Test multi-service: Run `cd examples/basic/multi-service && go run main.go`, verify all 3 services visible in `dns-sd -B` with correct ports

### Example 5: Service Browser

- [ ] T032 [P] [US2] Create examples/basic/browser/README.md for service enumeration per contracts/examples-basic.md (Example 5)
- [ ] T033 [P] [US2] Create examples/basic/browser/main.go (~70 lines) querying _services._dns-sd._udp.local per quickstart.md (Browser section)
- [ ] T034 [P] [US2] Create examples/basic/browser/go.mod per data-model.md template
- [ ] T035 [P] [US2] Create examples/basic/browser/Makefile per data-model.md template
- [ ] T036 [US2] Test browser: Run `cd examples/basic/browser && go run main.go`, verify discovers known services on network within 2 seconds

**Checkpoint**: All 5 basic examples compile, run, and demonstrate core Beacon responder functionality

---

## Phase 5: User Story 3 - Godoc API Documentation (Priority: P0) 🎯 MVP CORE

**Goal**: Provide runnable Godoc examples for all public APIs that appear on pkg.go.dev

**Independent Test**: Run `go test -v -run Example ./responder ./querier` and verify all examples compile and execute with expected output

**Template**: data-model.md (Godoc Example Template)
**Contract**: spec.md (FR-012 through FR-015)
**Quick Reference**: quickstart.md (Godoc Example Pattern)

### Responder API Examples

- [ ] T037 [P] [US3] Add ExampleResponder_Register() to responder/responder_test.go per data-model.md (Godoc Example Template)
- [ ] T038 [P] [US3] Add ExampleResponder_Unregister() to responder/responder_test.go demonstrating service unregistration with goodbye packets
- [ ] T039 [P] [US3] Add ExampleResponder_UpdateService() to responder/responder_test.go showing dynamic TXT record updates
- [ ] T040 [P] [US3] Add ExampleService_Validate() to responder/service_test.go demonstrating service validation logic

### Querier API Examples

- [ ] T041 [P] [US3] Add ExampleQuerier_Query() to querier/querier_test.go demonstrating basic mDNS query
- [ ] T042 [P] [US3] Add ExampleQuerier_QueryAll() to querier/querier_test.go showing service enumeration

### Advanced API Examples

- [ ] T043 [P] [US3] Add ExampleConflictDetector() to responder/conflict_detector_test.go demonstrating RFC 6762 §8.2 tie-breaking logic

### Validation

- [ ] T044 [US3] Test all Godoc examples: Run `go test -v -run Example ./responder ./querier`, verify all compile and execute with expected output matching `// Output:` comments

**Checkpoint**: 100% of public APIs have Godoc examples appearing on pkg.go.dev

---

## Phase 6: User Story 4 - Production Deployment Guidance (Priority: P0) 🎯 MVP CORE

**Goal**: Provide deployment guides for production environments (Docker, monitoring, troubleshooting)

**Independent Test**: Each guide should be followable step-by-step, with working examples that execute successfully

**Templates**: data-model.md (Deployment Guide Template, docker-compose.yml Template, Dockerfile Template)
**Contracts**: contracts/deployment-guides.md (Guides 1-4)
**Quick Reference**: quickstart.md (Deployment Guide Pattern)

### Guide 1: Production Checklist

- [ ] T045 [P] [US4] Create docs/deployment/production-checklist.md per contracts/deployment-guides.md (Guide 1) with pre-deployment validation checklist covering network, interfaces, services, resources, security, monitoring
- [ ] T046 [US4] Validate production-checklist.md: Verify all checklist items have validation commands and troubleshooting links

### Guide 2: Docker Deployment

- [ ] T047 [P] [US4] Create docs/deployment/docker.md per contracts/deployment-guides.md (Guide 2) explaining multicast requirements, network_mode: "host", and macvlan alternative
- [ ] T048 [P] [US4] Create docs/deployment/docker-example/Dockerfile per data-model.md (Dockerfile Template) with multi-stage build
- [ ] T049 [P] [US4] Create docs/deployment/docker-example/docker-compose.yml per data-model.md (docker-compose.yml Template) with network_mode: "host" for multicast
- [ ] T050 [P] [US4] Create docs/deployment/docker-example/main.go with simple Beacon responder service for Docker testing
- [ ] T051 [US4] Test Docker deployment: Run `cd docs/deployment/docker-example && docker-compose up`, verify service announces via mDNS and is discoverable from host

### Guide 3: Monitoring

- [ ] T052 [P] [US4] Create docs/deployment/monitoring.md per contracts/deployment-guides.md (Guide 3) demonstrating structured logging with slog, key metrics, health checks, alerting thresholds
- [ ] T053 [US4] Validate monitoring.md: Verify log schema documented, metrics list complete, health check example provided

### Guide 4: Troubleshooting

- [ ] T054 [P] [US4] Create docs/deployment/troubleshooting.md per contracts/deployment-guides.md (Guide 4) with at least 10 real-world scenarios: service not visible, can't connect, port conflicts, name conflicts, high CPU, etc.
- [ ] T055 [US4] Validate troubleshooting.md: Verify each scenario follows Problem Format template (Symptom → Diagnosis → Solution → Prevention)

**Checkpoint**: Production deployment guides complete - teams can confidently deploy to Docker with monitoring and troubleshooting support

---

## Phase 7: User Story 5 - Professional Documentation Site (Priority: P1)

**Goal**: Deploy Hugo documentation site to GitHub Pages with Docsy theme

**Independent Test**: Site accessible at https://joshuafuller.github.io/beacon/, loads in <3 seconds, all internal links resolve, search works

**Template**: data-model.md (Hugo Content Template)
**Contract**: spec.md (FR-021 through FR-026)
**Quick Reference**: quickstart.md (Hugo Site Deployment)

### Hugo Site Setup

- [ ] T056 [P] [US5] Create docs/config.toml per quickstart.md (Hugo Site Deployment) with Docsy theme configuration, site title "Beacon - mDNS for Go"
- [ ] T057 [US5] Add Docsy theme as git submodule: Run `cd docs && git submodule add https://github.com/google/docsy.git themes/docsy && git submodule update --init --recursive`
- [ ] T058 [US5] Pin Docsy version: Run `cd docs/themes/docsy && git checkout v0.10.0`
- [ ] T059 [P] [US5] Create docs/content/_index.md landing page with value proposition, quick start, badges per quickstart.md template
- [ ] T060 [P] [US5] Create docs/content/getting-started/_index.md section index
- [ ] T061 [P] [US5] Create docs/content/guides/_index.md section index
- [ ] T062 [P] [US5] Create docs/content/examples/_index.md section index
- [ ] T063 [P] [US5] Create docs/content/reference/_index.md section index
- [ ] T064 [P] [US5] Create docs/content/architecture/_index.md section index
- [ ] T065 [P] [US5] Create docs/static/images/ directory for diagrams
- [ ] T066 [US5] Test Hugo build: Run `cd docs && hugo serve`, verify site renders at http://localhost:1313

### GitHub Actions Deployment

- [ ] T067 [US5] Create .github/workflows/docs.yml per quickstart.md (GitHub Actions Deployment) for auto-deploy to GitHub Pages on push to main, paths: `docs/**`
- [ ] T068 [US5] Test GitHub Actions: Push to branch, verify workflow runs successfully and deploys to joshuafuller.github.io/beacon

### Content Migration

- [ ] T069 [P] [US5] Migrate docs/guides/getting-started.md → docs/content/getting-started/quickstart.md with Hugo front matter
- [ ] T070 [P] [US5] Migrate docs/guides/architecture.md → docs/content/architecture/overview.md with Hugo front matter
- [ ] T071 [P] [US5] Migrate docs/guides/troubleshooting.md → docs/content/guides/troubleshooting.md with Hugo front matter
- [ ] T072 [P] [US5] Convert RFC_COMPLIANCE_CERTIFICATE_v1.0.md → docs/content/reference/rfc-compliance.md with Hugo front matter
- [ ] T073 [US5] Update all internal links in migrated docs (search/replace relative paths to Hugo paths)

**Checkpoint**: Hugo site live at joshuafuller.github.io/beacon with searchable documentation

---

## Phase 8: User Story 6 - Intermediate Examples (Priority: P1)

**Goal**: Provide 5 intermediate examples showing production-ready patterns

**Independent Test**: Each example compiles, runs, and demonstrates production integration patterns

**Templates**: data-model.md (Example Template Structure)
**Contracts**: contracts/examples-intermediate.md (Examples 6-10)
**Quick Reference**: quickstart.md (Example Creation Steps)

### Example 6: Web Server with mDNS

- [ ] T074 [P] [US6] Create examples/intermediate/web-server/README.md per contracts/examples-intermediate.md (Example 6)
- [ ] T075 [P] [US6] Create examples/intermediate/web-server/main.go (~100 lines) with http.Server + responder integration
- [ ] T076 [P] [US6] Create examples/intermediate/web-server/go.mod per template
- [ ] T077 [P] [US6] Create examples/intermediate/web-server/Makefile per template
- [ ] T078 [US6] Test web-server: Run `cd examples/intermediate/web-server && go run main.go`, verify HTTP responds on :8080 AND service visible via mDNS

### Example 7: Service Updates

- [ ] T079 [P] [US6] Create examples/intermediate/service-updates/README.md per contracts/examples-intermediate.md (Example 7)
- [ ] T080 [P] [US6] Create examples/intermediate/service-updates/main.go (~90 lines) with UpdateService() showing dynamic TXT changes
- [ ] T081 [P] [US6] Create examples/intermediate/service-updates/go.mod per template
- [ ] T082 [P] [US6] Create examples/intermediate/service-updates/Makefile per template
- [ ] T083 [US6] Test service-updates: Run example, verify TXT records change every 5 seconds visible in `dns-sd -L`

### Example 8: Multi-Interface Subnet Bridge (IoT - USER REQUESTED)

- [ ] T084 [P] [US6] Create examples/intermediate/multi-interface-bridge/README.md per contracts/examples-intermediate.md (Example 8) explaining IoT use case (WiFi ↔ Ethernet bridging)
- [ ] T085 [P] [US6] Create examples/intermediate/multi-interface-bridge/main.go (~150 lines) with bridge orchestration, signal handling
- [ ] T086 [P] [US6] Create examples/intermediate/multi-interface-bridge/bridge.go (~100 lines) implementing Bridge type with forwarding logic, filtering, RFC 6762 §15 compliance
- [ ] T087 [P] [US6] Create examples/intermediate/multi-interface-bridge/config.yaml with interface names, service allowlist, subnet exclusions
- [ ] T088 [P] [US6] Create examples/intermediate/multi-interface-bridge/go.mod per template
- [ ] T089 [P] [US6] Create examples/intermediate/multi-interface-bridge/Makefile per template
- [ ] T090 [US6] Test multi-interface-bridge: Run on multi-interface machine (WiFi + Ethernet), verify queries forwarded between interfaces with interface-specific IP addressing

### Example 9: Custom Service Type

- [ ] T091 [P] [US6] Create examples/intermediate/custom-service-type/README.md per contracts/examples-intermediate.md (Example 9)
- [ ] T092 [P] [US6] Create examples/intermediate/custom-service-type/main.go (~80 lines) defining _myapp._tcp service with custom TXT schema
- [ ] T093 [P] [US6] Create examples/intermediate/custom-service-type/go.mod per template
- [ ] T094 [P] [US6] Create examples/intermediate/custom-service-type/Makefile per template
- [ ] T095 [US6] Test custom-service-type: Run example, verify custom service visible with `dns-sd -B _myapp._tcp`

### Example 10: Logging Integration

- [ ] T096 [P] [US6] Create examples/intermediate/logging-integration/README.md per contracts/examples-intermediate.md (Example 10)
- [ ] T097 [P] [US6] Create examples/intermediate/logging-integration/main.go (~100 lines) with slog structured logging for responder events
- [ ] T098 [P] [US6] Create examples/intermediate/logging-integration/go.mod per template
- [ ] T099 [P] [US6] Create examples/intermediate/logging-integration/Makefile per template
- [ ] T100 [US6] Test logging-integration: Run example, verify JSON structured logs output to stdout

### Migration Guide

- [ ] T101 [P] [US6] Create docs/content/migration/from-hashicorp-mdns.md with API comparison table per spec.md (FR-032, FR-033)
- [ ] T102 [P] [US6] Add at least 3 side-by-side code examples (hashicorp → Beacon) to migration guide
- [ ] T103 [P] [US6] Create docs/content/migration/from-grandcat.md (zeroconf library migration guide)

### Architecture Diagrams

- [ ] T104 [P] [US6] Create docs/content/architecture/message-flow.md with Mermaid sequence diagram (probe → announce → query → response)
- [ ] T105 [P] [US6] Create docs/content/architecture/state-machine.md with FSM diagram (Probing → Announcing → Established)
- [ ] T106 [P] [US6] Create docs/content/architecture/multi-interface.md with interface-specific addressing diagram (RFC 6762 §15)
- [ ] T107 [P] [US6] Create docs/content/architecture/buffer-pooling.md with before/after performance diagram
- [ ] T108 [US6] Export all Mermaid diagrams as SVG: Run `mmdc -i file.mmd -o docs/static/images/file.svg` for each diagram (requires mermaid-cli)

**Checkpoint**: 5 intermediate examples complete, migration guides available, architecture diagrams rendered

---

## Phase 9: User Story 7 - Advanced Content & Community Infrastructure (Priority: P2)

**Goal**: Provide advanced examples, Kubernetes deployment, video tutorials, and community contribution infrastructure

**Independent Test**: IoT example runs on Raspberry Pi, Kubernetes templates work in minikube, CONTRIBUTING.md enables first-time PRs

**Templates**: data-model.md (Example Template Structure)
**Contract**: spec.md (FR-036 through FR-040)

### Example 11: IoT Device Registration

- [ ] T109 [P] [US7] Create examples/advanced/iot-device/README.md for Raspberry Pi scenario
- [ ] T110 [P] [US7] Create examples/advanced/iot-device/main.go (~120 lines) with GPIO service, hardware detection
- [ ] T111 [P] [US7] Create examples/advanced/iot-device/go.mod per template
- [ ] T112 [P] [US7] Create examples/advanced/iot-device/Makefile per template
- [ ] T113 [US7] Test on Raspberry Pi: Run `cd examples/advanced/iot-device && go run main.go`, verify service registered

### Example 12: Microservice Discovery

- [ ] T114 [P] [US7] Create examples/advanced/microservices/README.md for service mesh scenario
- [ ] T115 [P] [US7] Create examples/advanced/microservices/main.go (~150 lines) with multi-service registration + discovery
- [ ] T116 [P] [US7] Create examples/advanced/microservices/docker-compose.yml with 3 services communicating
- [ ] T117 [P] [US7] Create examples/advanced/microservices/go.mod per template
- [ ] T118 [US7] Test microservices: Run `cd examples/advanced/microservices && docker-compose up`, verify cross-service discovery

### Example 13: Load Balancing

- [ ] T119 [P] [US7] Create examples/advanced/load-balancing/README.md for client-side load balancing
- [ ] T120 [P] [US7] Create examples/advanced/load-balancing/main.go (~130 lines) with round-robin across instances
- [ ] T121 [P] [US7] Create examples/advanced/load-balancing/go.mod per template
- [ ] T122 [P] [US7] Create examples/advanced/load-balancing/Makefile per template
- [ ] T123 [US7] Test load-balancing: Run example, verify distributes requests across multiple service instances

### Example 14: Printer Discovery

- [ ] T124 [P] [US7] Create examples/real-world/printer-discovery/README.md
- [ ] T125 [P] [US7] Create examples/real-world/printer-discovery/main.go (~90 lines) querying _ipp._tcp.local
- [ ] T126 [P] [US7] Create examples/real-world/printer-discovery/go.mod per template
- [ ] T127 [P] [US7] Create examples/real-world/printer-discovery/Makefile per template
- [ ] T128 [US7] Test printer-discovery: Run example, verify finds network printers

### Example 15: Chromecast Discovery

- [ ] T129 [P] [US7] Create examples/real-world/chromecast/README.md
- [ ] T130 [P] [US7] Create examples/real-world/chromecast/main.go (~100 lines) querying _googlecast._tcp
- [ ] T131 [P] [US7] Create examples/real-world/chromecast/go.mod per template
- [ ] T132 [P] [US7] Create examples/real-world/chromecast/Makefile per template
- [ ] T133 [US7] Test chromecast: Run example, verify discovers Chromecast devices

### Example 16: Home Assistant Integration

- [ ] T134 [P] [US7] Create examples/real-world/home-assistant/README.md
- [ ] T135 [P] [US7] Create examples/real-world/home-assistant/main.go (~110 lines) for smart home device registration
- [ ] T136 [P] [US7] Create examples/real-world/home-assistant/go.mod per template
- [ ] T137 [P] [US7] Create examples/real-world/home-assistant/Makefile per template
- [ ] T138 [US7] Test home-assistant: Run example, verify integrates with Home Assistant

### Kubernetes Deployment

- [ ] T139 [P] [US7] Create docs/content/deployment/kubernetes.md with DaemonSet + Service examples
- [ ] T140 [P] [US7] Create examples/deployment/kubernetes/deployment.yaml for Beacon DaemonSet
- [ ] T141 [P] [US7] Create examples/deployment/kubernetes/service.yaml (headless service for mDNS)
- [ ] T142 [P] [US7] Create examples/deployment/kubernetes/configmap.yaml for Beacon configuration
- [ ] T143 [US7] Test Kubernetes: Run `minikube start && kubectl apply -f examples/deployment/kubernetes/`, verify pods announce services

### Video Tutorials

- [ ] T144 [P] [US7] Record "5-Minute Beacon Introduction" video (overview, quick start, first service)
- [ ] T145 [P] [US7] Record "Step-by-Step Service Registration" video (detailed walkthrough with debugging)
- [ ] T146 [P] [US7] Record "Debugging Common mDNS Issues" video (firewall, multicast, subnet problems)
- [ ] T147 [US7] Upload videos to YouTube, embed in docs/content/tutorials/videos.md

### Community Contribution Guides

- [ ] T148 [P] [US7] Create CONTRIBUTING.md with PR workflow, testing requirements, code style per spec.md (FR-038)
- [ ] T149 [P] [US7] Create CODE_OF_CONDUCT.md using Contributor Covenant v2.1
- [ ] T150 [P] [US7] Create .github/ISSUE_TEMPLATE/bug_report.md per spec.md (FR-039)
- [ ] T151 [P] [US7] Create .github/ISSUE_TEMPLATE/feature_request.md per spec.md (FR-039)
- [ ] T152 [P] [US7] Create .github/PULL_REQUEST_TEMPLATE.md with checklist per spec.md (FR-040)
- [ ] T153 [P] [US7] Create docs/content/contributing/development-setup.md (Go setup, make targets, testing)

### Documentation Quality Infrastructure

- [ ] T154 [P] [US7] Create .markdownlint.json config (line length 120, no bare URLs)
- [ ] T155 [P] [US7] Create .github/workflows/docs-quality.yml (markdownlint + htmltest link checker) per spec.md (FR-026)
- [ ] T156 [P] [US7] Create .htmltest.yml config (ignore external links, check internal only)
- [ ] T157 [US7] Test CI: Run `make docs-quality` (if Makefile target added), verify markdown lints and links valid

**Checkpoint**: All 6 advanced examples complete, Kubernetes deployment ready, community infrastructure enables contributions

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements affecting multiple user stories

- [ ] T158 [P] Run full test suite: `make test` to ensure no regressions from documentation work
- [ ] T159 [P] Run example compilation test: Build all examples with `for dir in examples/*/*/; do (cd "$dir" && go build main.go); done`
- [ ] T160 [P] Verify all README badges render correctly on GitHub
- [ ] T161 [P] Check Hugo site: Verify all internal links resolve, no broken links via htmltest
- [ ] T162 [P] Performance check: Verify Hugo site loads in <3 seconds on 3G connection (use Chrome DevTools Network throttling)
- [ ] T163 Update DOCUMENTATION_STRATEGY.md to mark all P0, P1, P2 items as complete
- [ ] T164 Update BIG_PICTURE_ANALYSIS.md to close documentation gap findings
- [ ] T165 Final validation: Run through quickstart.md validation checklist for each user story

---

**Total Tasks**: 165
**P0 Tasks (MVP)**: 52 (US1: 8, US2: 25, US3: 8, US4: 11)
**P1 Tasks**: 55 (US5: 18, US6: 37)
**P2 Tasks**: 49 (US7: 49)
**Polish Tasks**: 8

**Format Validation**: ✅ All 165 tasks follow checklist format with checkbox, ID, [P] marker (where applicable), [Story] label (for user story phases), and exact file paths

**Suggested MVP Scope (v1.0)**: Complete P0 only (Tasks T001-T055) - 52 tasks
**Incremental Delivery (v1.1)**: Add P1 (Tasks T056-T108) - +55 tasks
**Future Enhancements (v1.2+)**: Add P2 (Tasks T109-T157) - +49 tasks
