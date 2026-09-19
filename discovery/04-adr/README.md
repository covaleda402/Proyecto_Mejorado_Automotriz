# Architectural Decision Records (ADRs)

This directory documents the core architectural and technical decisions made during the design, refactoring, and security remediation of the Automotive Technical Support and Workshop Management System.

## Decision Index

| ID | Title | Status | Date | Core Decision |
|---|---|---|---|---|
| [ADR-001](./ADR-001-clean-architecture-go.md) | Adoption of Clean Architecture in Go | Accepted | 2026-09 | 4 decoupled layers (`domain`, `usecase`, `repository`, `transport/http`) with inward dependency inversion. |
| [ADR-002](./ADR-002-rbac-and-bola-protection.md) | Mitigation of OWASP A01: RBAC and Service Order Isolation (BOLA) | Accepted | 2026-09 | Administrator role enforcement on sensitive catalogs and strict technician-to-order ownership verification. |
| [ADR-003](./ADR-003-session-tokens-and-rate-limiting.md) | Authentication Tokens and Sliding Window Rate Limiting | Accepted | 2026-09 | Bcrypt cryptographic hashing, Bearer token session validation, and in-memory rate limiter against brute force attacks. |
| [ADR-004](./ADR-004-strict-lifecycle-and-state-machine.md) | Deterministic Finite State Machine & Lifecycle Invariants | Accepted | 2026-09 | Strict linear order transitions, mandatory diagnostic/intervention prerequisites, and immutability for delivered orders. |
| [ADR-005](./ADR-005-database-orchestration-and-migration.md) | Database Orchestration, Schema Migration and Startup Automation | Accepted | 2026-09 | Standardized MySQL schemas with relational integrity, Docker Compose containerization, and batch initialization scripts. |
