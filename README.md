# Legal-RIU

Legal RIU Portal is a full-stack legal case management system built with Golang (Gin), React, PostgreSQL, and MinIO. It implements RBAC + PBAC, JWT auth, document versioning, and a structured workflow engine for Legal Opinion & Document Review with audit-friendly status transitions and file lifecycle tracking.

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

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key variables:
- `DB_*` — PostgreSQL connection
- `JWT_SECRET` — JWT signing secret
- `MINIO_*` — MinIO S3 storage credentials
- `ADMIN_EMAIL`, `ADMIN_PASSWORD` — Initial admin account

## Architecture Docs

- [Workflow](docs/workflow.md) — Backend & frontend request flow, conventions, how to add features
- [Permission Matrix](docs/permission_matrix.md) — Permission catalog, role baseline, override mechanics, coverage status