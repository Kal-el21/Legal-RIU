# ALEA

ALEA is a full-stack legal case management system built with Golang (Gin), React, PostgreSQL, and MinIO. It implements RBAC + PBAC, JWT auth, document versioning, and a structured workflow engine for Legal Opinion & Document Review with audit-friendly status transitions and file lifecycle tracking.

## Quick Links

- [Workflow Documentation](docs/workflow.md) — Backend & frontend request flow, conventions, and how to add features
- [Permission Matrix](docs/permission_matrix.md) — Full permission catalog, role matrix, and override system docs

## System Roles

| Role | Description | Scope |
|------|-------------|-------|
| **ADMIN** | System administrator | Full access — user management, all modules, permission overrides |
| **LEGAL** | Legal reviewer | Review & update status of legal opinions, review documents, case management, materials |
| **LEGAL_AU** | Subsidiary legal user | Own-company case management, materials |
| **USER** | Regular user | Own submissions (legal opinion, review document), public materials |
| **EXTERNAL** | External/corporate user | Case management only — view case, upload document, manage chronology |

## Modules & Access

| Module | ADMIN | LEGAL | LEGAL_AU | USER | EXTERNAL |
|--------|-------|-------|----------|------|----------|
| Dashboard | ✅ | ✅ | ❌ | ✅ | ❌ |
| Case Management | ✅ | ✅ | ✅ | ❌ | ✅ |
| Legal Opinion | ✅ | ✅ | ❌ | ✅ | ❌ |
| Review Document | ✅ | ✅ | ❌ | ✅ | ❌ |
| Legal Material | ✅ | ✅ | ✅ | ✅ public | ❌ |
| User Management | ✅ | ❌ | ❌ | ❌ | ❌ |
| Audit Log | ✅ | ✅ | ❌ | ❌ | ❌ |
| Settings | ✅ | ✅ | ❌ | ✅ | ✅ |

## Key Features

- **RBAC + PBAC Authorization** — Role baseline + per-user permission override (ALLOW/DENY)
- **Case Management** — Full case tracking with chronology, document upload, status updates
- **Legal Opinion** — Submission workflow with review, approval, and PDF generation
- **Document Review** — Draft agreement review with revision cycles
- **Material Management** — Legal material CRUD with role-based access
- **Audit Logging** — All sensitive actions logged with `PERMISSION_UPDATE` tracking
- **File Storage** — MinIO integration for secure document management
- **JWT Auth** — Access + refresh token with rotation

## Tech Stack

- **Backend**: Go (Gin framework), GORM, PostgreSQL, MinIO
- **Frontend**: React, TypeScript, TanStack Query, Tailwind CSS, Zustand
- **Infrastructure**: Docker Compose, Nginx reverse proxy

## Permission System

The system uses a 3-layer authorization model:

1. **Role** — Baseline permissions per role
2. **Permission Override** — Admin can ALLOW or DENY granular permissions per user
3. **Effective Permissions** — Final permissions = Role + ALLOW - DENY

Permission format: `feature.action.scope`

Example: `case_management.create`, `legal_opinion.update.own`, `document_review.download.all`

See [Permission Matrix](docs/permission_matrix.md) for full catalog.

## Project Structure

```
├── backend/               # Go backend (Gin)
│   ├── cmd/api/           # Main server entrypoint
│   ├── internal/          # Core application code
│   │   ├── handler/       # HTTP handlers
│   │   ├── service/       # Business logic
│   │   ├── repository/    # Data access
│   │   ├── entity/        # Database models
│   │   ├── middleware/    # Auth, role, permission
│   │   └── seed/          # Permission & role seeds
│   └── migrations/        # SQL migrations
├── frontend/              # React frontend (Vite)
│   └── src/
│       ├── pages/         # Route pages by role
│       ├── components/    # Shared UI components
│       ├── services/      # API service layer
│       ├── store/         # Zustand state
│       └── routes/        # Router + guards
├── docs/                  # Documentation
│   ├── workflow.md        # Full workflow docs
│   └── permission_matrix.md
├── docker-compose.yml     # Infrastructure
└── README.md
```

## Development

```bash
# Start all services (PostgreSQL, MinIO, backend, frontend)
docker compose up -d

# Backend manual dev (with auto-migration + seed)
cd backend && go run ./cmd/api

# Frontend manual dev
cd frontend && npm install && npm run dev
```

Frontend runs at `http://localhost:5173`, Backend API at `http://localhost:8080`.

Before pushing, run every check (go vet, go test, tsc, lint, build) from PowerShell:
`.\scripts\cek.ps1`. It stops at the first check that fails.

The backend API tests (`backend/internal/router`) need PostgreSQL and MinIO running.
They read connection settings from `backend/.env`, recreate their own database
(`legal_riu_test`) on every run and use their own bucket (`legal-riu-test`), so
development data is never touched. Set `TES_API_LEWATI=1` to skip them.

GitHub Actions cannot pull the MinIO image any more, so it runs with
`TES_TANPA_MINIO=1`: everything that only needs PostgreSQL still runs there, and
tests that touch file storage run locally only. Do not set it on your machine.

Before deploying, also run `.\scripts\cek-docker.ps1`. It runs the same API tests inside
the `legal_riu_backend` container, the only place with PostgreSQL, MinIO and LibreOffice
together, so it is the only run that covers agreement approval (DOCX to PDF). Use
`-Tes <name>` to run matching tests only and `-Rinci` to list every test.

Tests that describe correct behaviour the code does not have yet are skipped with
`bugDiketahui(...)`; run `go test ./internal/router -v` and look for "BUG DIKETAHUI" to
list them. After fixing one, check it with `TES_BUG=1` and then delete that call.

`testdata/akses.golden` records which roles can reach each endpoint. When an access
change is intentional, refresh it from `backend/` and commit the result:
`go test ./internal/router -run TestFotoHakAkses -update`.

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key variables:

- `DB_*` — PostgreSQL connection
- `JWT_SECRET` — JWT signing secret
- `MINIO_*` — MinIO S3 storage credentials
- `LDAP_*` — Active Directory connection for signing in with an office username; `LDAP_BIND_PASSWORD` must be set, and `LDAP_DEFAULT_DIVISION` should stay empty so the division comes from the user's OU
- `ADMIN_EMAIL`, `ADMIN_PASSWORD` — Initial admin account

## Architecture Docs

- [Workflow](docs/workflow.md) — Backend & frontend request flow, conventions, how to add features
- [Permission Matrix](docs/permission_matrix.md) — Permission catalog, role baseline, override mechanics, coverage status
