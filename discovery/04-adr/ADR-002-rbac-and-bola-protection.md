# ADR-002: Mitigation of OWASP A01: Role-Based Access Control and Service Order Isolation (BOLA)

## Status
Accepted

## Context
During security audits, the system exhibited two major vulnerabilities classified under OWASP A01 (Broken Access Control):
1. **Vertical Privilege Escalation:** Technicians could read sensitive customer PII, vehicle registries, and warranty policies via direct HTTP calls to `GET /api/customer`, `GET /api/vehicle`, and `GET /api/warranty`.
2. **Broken Object Level Authorization (BOLA):** Any authenticated technician could query orders belonging to peers and advance their state (`POST /api/service-order/:id/status`) without being assigned to the job.

## Decision
We implemented comprehensive access control across both frontend and backend layers:

1. **Role Verification at the API Boundary:**
   - Introduced `requireAdministrator(ctx)` in HTTP handlers for customer, vehicle, technician, and warranty endpoints.
   - Any attempt by a technician to query or mutate these resources returns an immediate `403 Forbidden`.

2. **Resource Ownership Enforcement (BOLA Mitigation):**
   - In `ServiceOrderUseCase.List`: If the caller is a technician, the query delegates to `ListByTechnicianUser`, returning exclusively orders assigned to that technician.
   - In `ServiceOrderUseCase.Find`: If the caller is a technician, the use case verifies that the active assignment's `technicianId` matches the caller's technician profile. Unassigned access returns `403 Forbidden`.
   - In `ServiceOrderUseCase.Advance`: Only the designated assigned technician (or an administrator) can trigger lifecycle transitions.

3. **Client-Side Route Protection:**
   - Encapsulated private routes in React Router using `<ProtectedRoute administratorOnly>` for `/customers`, `/vehicles`, `/warranties`, and `/technicians`.
   - Unauthorized attempts immediately redirect the user to `/service-orders`.

## Consequences

### Positive
- Strict compliance with data privacy regulations by hiding customer PII and vehicle registries from non-administrative staff.
- Prevents cross-technician tampering and maintains full audit traceability throughout the repair pipeline.

### Negative / Mitigations
- Requires an additional database lookup to verify assignment during order operations; mitigated by database indexing on foreign keys (`service_order_id`, `technician_id`).
