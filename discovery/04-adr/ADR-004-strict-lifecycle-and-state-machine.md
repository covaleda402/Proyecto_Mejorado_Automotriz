# ADR-004: Deterministic Finite State Machine & Lifecycle Invariants

## Status
Accepted

## Context
In earlier system iterations, service orders suffered from business logic inconsistencies: orders could transition to repair states without diagnostic reports, or states could be bypassed. Moreover, completed orders in `DELIVERED` status could be edited retroactively, compromising the legal and financial history of workshop operations.

## Decision
We implemented a deterministic finite state machine governed by strict domain invariants:

1. **Strict Linear Progression:**
   The service order lifecycle is rigidly sequenced:
   ```text
   RECEIVED -> IN_DIAGNOSIS -> IN_REPAIR -> READY -> DELIVERED
   ```
   Arbitrary state jumps and backwards transitions are rejected by `order.MoveTo()` with `domain.ErrInvalidTransition`.

2. **Mandatory Activity Prerequisites:**
   - Moving from `RECEIVED` to `IN_DIAGNOSIS` requires a verified diagnostic report in the database.
   - Moving from `IN_DIAGNOSIS` to `IN_REPAIR` requires at least one recorded repair intervention.
   - If prerequisites are not satisfied, the operation is blocked with a domain error.

3. **Immutability of Delivered Orders:**
   - When an order reaches `DELIVERED`, the system automatically releases the active technician assignment.
   - Any subsequent call to modify the order status, attach diagnostics, or record interventions returns `403 Forbidden` (`domain.ErrForbidden`).

## Consequences

### Positive
- Guarantees operational integrity and audit trail compliance: every repair invoice correlates with recorded parts and diagnostics.
- Prevents accidental or fraudulent post-delivery modifications.

### Negative / Mitigations
- In scenarios where a customer returns for warranty claims on a completed repair, technicians cannot edit the closed order; the system requires creating a new warranty service order linked to the existing vehicle.
