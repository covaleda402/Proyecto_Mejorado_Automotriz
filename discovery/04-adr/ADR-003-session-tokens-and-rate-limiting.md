# ADR-003: Authentication Tokens, Cryptographic Hashing, and Rate Limiting

## Status
Accepted

## Context
Prior security evaluations noted the absence of brute force mitigation on the authentication endpoint (`POST /api/session`), allowing automated credential guessing attacks. Additionally, safe password handling and session token integrity are paramount for the workshop environment.

## Decision
We implemented a multi-layered authentication defense:

1. **In-Memory Sliding Window Rate Limiter:**
   - Implemented `LoginRateLimiter` protecting `POST /api/session`.
   - Tracks failed authentication attempts per remote IP address.
   - Restricts login attempts to a maximum of **5 failed attempts within a 15-minute sliding window**.
   - Any further attempt returns **HTTP 429 Too Many Requests** (`{"error": "too_many_requests"}`).
   - Successful authentications automatically reset the counter for that IP.

2. **Cryptographic Password Storage:**
   - Passwords are encrypted using **bcrypt** with default cost (10+ rounds), ensuring resilience against rainbow table and offline dictionary attacks.

3. **Session Token Issuance:**
   - Generated tokens are cryptographically signed using an HMAC secret configured via environment variables (`TOKEN_SECRET`).
   - Tokens have an explicit expiration time configured by `TOKEN_TTL_MINUTE` (defaulting to 8 hours).
   - Validated on every private request by `authMiddleware`.

## Consequences

### Positive
- Fully neutralizes credential stuffing and automated brute force attacks on user logins.
- Secure, stateless authentication scalable across microservice or containerized instances.

### Negative / Mitigations
- In-memory rate limiting state is tied to the server instance lifecycle; acceptable for single-instance deployments and development evaluation. For multi-node distributed clusters, a Redis-backed rate limiter can be substituted.
