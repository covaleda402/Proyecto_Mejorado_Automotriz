# Technical Security Rules

> Mandatory technical controls that apply to all project code.
> These rules complement the security policy (`security-policy.md`) with
> concrete implementation practices.

---

## OWASP Top 10 — Controls per category

### A01 — Broken Access Control (BAC & BOLA)

```go
// ❌ BAD — trusting client input or omitting ownership checks
func (h ServiceOrderHandler) Advance(writer http.ResponseWriter, request *http.Request) {
    orderID := request.PathValue("serviceOrderId")
    // Advances order without checking caller's role or assigned technician!
    h.usecase.Advance(request.Context(), orderID, nextStatus, callerID, callerRole)
}

// ✅ GOOD — enforce RBAC and verify ownership in Use Case & Handler
func (s ServiceOrderUseCase) Advance(ctx context.Context, orderID string, next domain.ServiceOrderStatus, actorUserID string, actorRole domain.Role) (domain.ServiceOrder, error) {
    order, err := s.order.FindByID(ctx, orderID)
    if err != nil {
        return domain.ServiceOrder{}, err
    }
    if order.Status == domain.StatusDelivered {
        return domain.ServiceOrder{}, domain.ErrForbidden // Delivered orders are locked
    }
    if actorRole == domain.RoleTechnician {
        profile, err := s.technician.FindByUserID(ctx, actorUserID)
        if err != nil {
            return domain.ServiceOrder{}, domain.ErrForbidden
        }
        active, err := s.assignment.FindActiveByServiceOrder(ctx, orderID)
        if err != nil || active.TechnicianID != profile.ID {
            return domain.ServiceOrder{}, domain.ErrForbidden // Technicians only advance their assigned orders
        }
    }
    return s.executeTransition(ctx, order, next, actorUserID)
}
```

**Mandatory Rules:**
- Every protected backend endpoint MUST pass through `authMiddleware` verifying the Bearer token.
- Role-based authorization: Sensitive endpoints (`GET /api/customer`, `GET /api/vehicle`, `GET /api/technician`, `GET /api/warranty`) strictly require `requireAdministrator(request.Context())`, returning `403 Forbidden` if invoked by technicians.
- Broken Object Level Authorization (BOLA) mitigation: Technicians can only list (`ListByTechnicianUser`), view (`Find`), and update (`Advance`) service orders explicitly assigned to them.
- Delivered orders (`DELIVERED`) are strictly immutable: no further status transitions, diagnostics, or interventions are allowed.
- Frontend route protection: React Router routes must wrap sensitive pages with `<ProtectedRoute administratorOnly>` (e.g. `/customers`, `/vehicles`, `/warranties`, `/technicians`), automatically redirecting unauthorized roles to `/service-orders`.

---

### A02 — Cryptographic Failures

**Mandatory Rules:**
- Passwords MUST be hashed with **bcrypt** (`golang.org/x/crypto/bcrypt`) using default cost factor (≥ 10). Never use plain text, MD5, or SHA-1.
- Session tokens are signed using a secure secret configured via `TOKEN_SECRET` with an expiration (`TOKEN_TTL_MINUTE`).
- Sensitive data in transit: HTTPS is mandatory in staging and production environments.
- Passwords and token secrets must NEVER be logged in application logs or stored in repository commits.

---

### A03 — Injection and Input Sanitization

```go
// ❌ BAD — direct SQL string concatenation
query := fmt.Sprintf("SELECT id, plate FROM vehicles WHERE plate = '%s'", plate)

// ✅ GOOD — parameterized queries across all repositories
const query = "SELECT id, plate, vin FROM vehicles WHERE plate = ?"
err := db.QueryRowContext(ctx, query, plate).Scan(&id, &plate, &vin)

// ✅ Anti-XSS and Anti-HTML validation in domain models
if err := domain.EnsureNoHTML("reportedFailure", reportedFailure); err != nil {
    return domain.ServiceOrder{}, err // Returns 400 invalid_input
}
```

**Mandatory Rules:**
- Parameterized queries ALWAYS via `database/sql` placeholders (`?`). Direct string concatenation is forbidden.
- Domain-level validation: User inputs (names, documents, descriptions, failure reports) MUST be validated against script/HTML tag injection using `domain.EnsureNoHTML()`.
- Inputs containing `<script>` or raw HTML tags MUST be rejected immediately with `400 Bad Request (invalid_input)`.

---

### A04 — Insecure Design & Finite State Lifecycle

- Deterministic state flow: Service orders follow a strict, non-skippable progression:
  ```text
  RECEIVED -> IN_DIAGNOSIS -> IN_REPAIR -> READY -> DELIVERED
  ```
- Workflow prerequisites:
  - Transition to `IN_DIAGNOSIS` requires a previously recorded diagnostic.
  - Transition to `IN_REPAIR` requires at least one recorded intervention.
- Completed jobs: Once an order reaches `DELIVERED`, it is finalized, its technician assignment is released, and all write operations are permanently blocked.

---

### A05 — Security Misconfiguration

```text
# Verification checklist per environment
□ Database credentials loaded exclusively from environment variables (.env)
□ CORS configured with specific allowed origin (ALLOWED_ORIGIN=http://localhost:4173)
□ Debug stack traces hidden from HTTP error responses (return structured error JSON)
□ Database ports isolated and protected by strong passwords
```

---

### A06 — Vulnerable Components

**Mandatory Rules:**
- Run dependency security checks before each release:
  - Backend: `go list -m all` and Go vulnerability scanner (`govulncheck`).
  - Frontend: `npm audit` with 0 high/critical vulnerabilities.
- Pin explicit package versions in `package.json` and `go.mod`.

---

### A07 — Identification, Authentication Failures & Rate Limiting

```go
// Sliding-window in-memory rate limiter on POST /api/session
func (l *LoginRateLimiter) Allow(ip string) bool {
    // Allows up to 5 failed attempts per IP within a 15-minute window.
    // 6th attempt returns HTTP 429 Too Many Requests.
}
```

**Mandatory Rules:**
- Strict Rate Limiting: Authentication endpoint (`POST /api/session`) is protected by `LoginRateLimiter` configured for a maximum of **5 failed login attempts per IP within a 15-minute window**.
- Subsequent attempts exceeding the threshold return **HTTP 429 Too Many Requests** (`{"error": "too_many_requests"}`).
- Session tokens expire automatically according to `TOKEN_TTL_MINUTE`.

---

### A08 — Software and Data Integrity Failures

- Automated build verification: `tsc --noEmit && vite build` and `go build ./cmd/server` must pass with zero warnings or errors.
- Automated tests: Unit and integration tests must run and pass before merging into `main` (`go test ./...` and `vitest run --coverage`).

---

### A09 — Security Logging and Monitoring Failures

- Authentication and authorization failures MUST return structured HTTP errors (`401 unauthorized`, `403 forbidden`).
- Internal server errors log diagnostic details to server stderr while returning sanitized messages to clients.

---

### A10 — Server-Side Request Forgery (SSRF)

- The workshop service does not make dynamic outbound HTTP requests based on user-supplied URLs.
- External integrations (such as webhooks or third-party APIs) must use hardcoded, vetted endpoint configurations.

---

## Correlations

- Security policy → `00-governance/security-policy.md`
- Architectural Decision Records → `04-adr/`
- Definition of Done → `00-governance/definition-of-done.md`
