# ADR-001: Adoption of Clean Architecture in Go Backend

## Status
Accepted

## Context
The automotive technical support system requires high reliability, strict traceability, and total isolation of business logic from delivery mechanisms (REST HTTP) and persistence storage (MySQL). To ensure maintainability, independent unit testing, and adherence to SOLID design principles, a decoupled software architecture was required.

## Decision
We implemented **Clean Architecture** in the Go 1.22+ backend, structuring the codebase into concentric layers adhering to the unidirectional Dependency Rule pointing inward:

1. **`internal/domain` (Core / Entities Layer):**
   - Encapsulates pure enterprise entities (`ServiceOrder`, `Vehicle`, `Customer`, `Technician`, `Diagnostic`, `Intervention`, `Warranty`), value objects, and domain errors (`ErrInvalidTransition`, `ErrForbidden`, `ErrNotFound`, `ErrInvalidInput`).
   - Zero external dependencies: relies exclusively on the Go standard library.
   - Hosts pure domain invariant validation functions, including `EnsureNoHTML`.

2. **`internal/usecase` (Application / Use Cases Layer):**
   - Coordinates application workflows: `ServiceOrderUseCase`, `CustomerUseCase`, `VehicleUseCase`, `WarrantyUseCase`, etc.
   - Declares narrow repository interfaces for Dependency Inversion (`ServiceOrderRepository`, `AssignmentRepository`, etc.).
   - Enforces business authorization rules (e.g. validating technician assignment before order mutations).

3. **`internal/repository` (Infrastructure / Persistence Adapters Layer):**
   - Implements use case repository interfaces using SQL prepared statements with MySQL via `database/sql`.
   - Handles data mapping between relational SQL tables and domain entities.

4. **`internal/transport/http` (Delivery / Interface Adapters Layer):**
   - Standard HTTP route multiplexing, JSON payload decoding/encoding, and HTTP security middlewares (authentication, RBAC role validation, rate limiting, and CORS).

## Consequences

### Positive
- **High Testability:** Business logic and domain use cases are tested instantaneously with unit tests without needing external databases.
- **Infrastructure Agnosticism:** Database drivers or external libraries can be swapped without touching domain logic.
- **Layered Security:** Access controls are verified at both the transport boundary (roles) and the use case boundary (resource ownership).

### Negative / Mitigations
- Increased boilerplate due to interface definitions and domain-to-DTO mappings. Mitigated by keeping interfaces small and focused.
