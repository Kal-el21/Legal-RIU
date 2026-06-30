# Legal-RIU

Legal RIU Portal is a full-stack legal case management system built with Golang (Gin), React, PostgreSQL, and MinIO. It implements RBAC, JWT auth, document versioning, and a structured workflow engine for Legal Opinion & Document Review with audit-friendly status transitions and file lifecycle tracking.

## Roles

| Role | Description | Access |
|------|-------------|--------|
| **ADMIN** | Super admin | Full access - user management, all submissions, status updates, upload results |
| **LEGAL** | Legal reviewer | Review & update status of Legal Opinions and Document Reviews, upload results |
| **USER** | Regular user | Create/edit own submissions, resubmit revisions, download completed results |
| **EXTERNAL** | External user | View and download completed results only |

## Features

- **Legal Opinion** - Submit legal opinion requests with document attachments
- **Document Review** - Submit document review requests with draft agreements
- **Dashboard** - Statistics and recent activity tracking per role
- **File Storage** - MinIO integration for document management
- **Authentication** - JWT-based auth with refresh tokens

## Tech Stack

- **Backend**: Go (Gin framework), GORM, PostgreSQL, MinIO
- **Frontend**: React, TypeScript, TanStack Query, Tailwind CSS
- **Infrastructure**: Docker Compose

## Development

```bash
# Start all services
docker compose up -d

# Or run backend manually
cd backend && go run ./cmd/api
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Key variables:
- `DB_*` - PostgreSQL connection
- `JWT_SECRET` - JWT signing secret
- `MINIO_*` - MinIO S3 storage credentials
- `ADMIN_EMAIL`, `ADMIN_PASSWORD` - Initial admin account