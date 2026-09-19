# ADR-005: Database Orchestration, Schema Migration, and Automated Startup

## Status
Accepted

## Context
Running the full automotive workshop stack requires coordinating a relational database engine (MySQL 8.0), a compiled Go backend API, and a React/TypeScript Vite frontend. Setting up relational schemas manually across different evaluation environments often leads to misconfigured credentials, missing foreign keys, or database connection failures.

## Decision
We established a standardized environment orchestration strategy tailored to both containerized workflows and local Windows environments:

1. **Unified Schema Initialization (`database/init_complete_db.sql`):**
   - Consolidated table definitions, indexes, foreign key constraints with `ON DELETE RESTRICT`, and seed records (administrator and technician accounts, demo vehicles, and service catalog) into a single idempotent script.
   - Ensures deterministic database state from the initial execution.

2. **Containerized Deployment (`docker-compose.yml` & `database/docker-compose.db.yml`):**
   - Configured Docker Compose definitions with persistent data volumes and automated health checks, enabling one-command provisioning of the database service.

3. **Automated Windows Startup Tooling:**
   - **`setup_database.bat`:** Interactive script prompting for MySQL root credentials and applying the complete schema initialization.
   - **`start-backend.bat` & `start-frontend.bat`:** Dedicated runners configuring paths and environment variables for each service.
   - **`start-all.bat`:** Root master orchestration script that initializes backend and frontend services concurrently in parallel command windows with clear access URLs.

## Consequences

### Positive
- One-click startup for evaluation and grading without complex manual terminal commands.
- Consistent schema state across local bare-metal MySQL installations and containerized Docker environments.
- Relational integrity enforced at the database level through foreign key constraints matching Go domain rules.

### Negative / Mitigations
- Batch scripts (`.bat`) are specific to Windows environments; containerized `docker-compose.yml` serves as the cross-platform standard for Linux/macOS.
