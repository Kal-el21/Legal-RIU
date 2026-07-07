# Legal RIU Portal — Project Workflow

## Document Purpose
Dokumen ini menjelaskan alur kerja development untuk frontend dan backend, mulai dari request HTTP sampai response, termasuk permission system, role-based routing, dan data flow.

---

## Table of Contents
1. [High-Level Architecture](#high-level-architecture)
2. [Backend Workflow](#backend-workflow)
3. [Frontend Workflow](#frontend-workflow)
4. [Permission & Authorization Flow](#permission--authorization-flow)
5. [Data Flow Examples](#data-flow-examples)
6. [Development Conventions](#development-conventions)
7. [Adding a New Feature](#adding-a-new-feature)

---

## High-Level Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Browser   │────▶│   Nginx     │────▶│   Frontend  │
│  (Vite)     │     │  Reverse    │     │   (Build)   │
└─────────────┘     │   Proxy     │     └─────────────┘
                    └──────┬──────┘
                           │ API Proxy
                           ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  PostgreSQL │◀────│   Backend   │────▶│    MinIO    │
│  Database   │     │   (Go/Gin)  │     │   Storage   │
└─────────────┘     └─────────────┘     └─────────────┘
```

### Core Principles
- **Backend**: Clean architecture (`cmd`, `internal/{handler,service,repository,entity}`)
- **Frontend**: Feature-based routing with role guards
- **Auth**: JWT access + refresh token
- **Authorization**: RBAC + PBAC with user permission overrides
- **Storage**: MinIO for file uploads/downloads

---

## Backend Workflow

### Directory Structure
```
backend/
├── cmd/
│   ├── api/                    # Main HTTP server entrypoint
│   │   └── main.go             # Router setup, middleware, server start
│   └── seed-admin/             # CLI seed untuk buat admin default
├── internal/
│   ├── config/                 # Environment config (viper)
│   ├── dto/                    # Request/Response DTOs
│   ├── entity/                 # Database models (GORM)
│   ├── handler/                # HTTP handlers (controllers)
│   ├── middleware/             # Auth, permission, role middleware
│   ├── repository/             # Database queries (repository pattern)
│   ├── seed/                   # Permission & role seed data
│   │   └── data/
│   ├── service/                # Business logic
│   ├── storage/                # MinIO file storage
│   ├── utils/                  # Helpers (JWT, pagination, etc.)
│   └── validator/              # Request validation
├── migrations/                 # SQL migrations
└── go.mod
```

### Request Flow (Backend)

```
HTTP Request
    │
    ▼
┌─────────────────────┐
│   main.go: Router   │
│   Group by module   │
│   /admin, /legal,   │
│   /external, etc.   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Auth Middleware     │
│  Validate JWT token  │
│  Attach user to ctx  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Role Middleware      │
│ Check user.role ==   │
│ required role        │
│ Example: ADMIN,      │
│ LEGAL, EXTERNAL      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│Permission Middleware │
│ hasPermission(code)  │
│ = role + ALLOW - DENY│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│      Handler         │
│  Parse DTO           │
│  Call service        │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│      Service         │
│  Business logic      │
│  Call repository     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│    Repository        │
│  GORM queries        │
│  Return entity       │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│     Response         │
│  JSON to frontend    │
└─────────────────────┘
```

### Key Backend Files

| File | Purpose |
|------|---------|
| `cmd/api/main.go` | Router setup, middleware chain, server bootstrap |
| `internal/middleware/auth.go` | JWT validation, user context injection |
| `internal/middleware/role.go` | Role-based access control |
| `internal/middleware/permission.go` | Permission-based access control with override calc |
| `internal/handler/*.go` | HTTP handlers per module |
| `internal/service/*.go` | Business logic layer |
| `internal/repository/*.go` | Database access layer |
| `internal/entity/*.go` | GORM models |
| `internal/seed/permissions.go` | Permission catalog + role baseline assignments |

### Backend Naming Convention

| Pattern | Example |
|---------|---------|
| Handler function | `GetAll`, `GetByID`, `Create`, `Update`, `Delete` |
| Router group | `admin`, `legal`, `external`, `legalAU` |
| Permission code | `feature.action.scope` → `case_management.create` |
| Entity name | Singular, PascalCase → `LegalCase`, `User` |

---

## Frontend Workflow

### Directory Structure
```
frontend/src/
├── app/                    # App providers/router wrapper
├── assets/                 # Static assets
├── components/
│   ├── common/             # Reusable UI components
│   │   ├── PermissionGate.tsx   # Reusable permission wrapper (optional)
│   │   ├── StatusBadge.tsx
│   │   └── ...
│   └── shared/             # Shared page-level components
│       ├── MaterialListPage.tsx
│       ├── LegalOpinionListPage.tsx
│       └── ReviewDocumentListPage.tsx
├── constants/              # App-wide constants
├── hooks/                  # React Query hooks per module
├── layouts/                # Layout components per role
│   ├── AdminLayout.tsx
│   ├── LegalLayout.tsx
│   ├── DashboardLayout.tsx
│   └── ExternalLayout.tsx
├── lib/                    # Utils, formatters
├── pages/
│   ├── admin/              # ADMIN role pages
│   ├── dashboard/          # USER role pages
│   ├── legal/              # LEGAL role pages
│   ├── legal-au/           # LEGAL_AU role pages
│   ├── external/           # EXTERNAL role pages (some dormant)
│   ├── settings/           # Shared settings page
│   ├── auth/               # Login page
│   └── public/             # Public/landing pages
├── routes/                 # Route definitions + guards
│   ├── index.tsx           # Main router
│   └── guards.tsx          # Role-based route guards
├── services/               # API service layer
├── store/                  # Zustand auth store
├── types/                  # TypeScript interfaces
└── utils/                  # Helper functions
```

### Request Flow (Frontend)

```
User Action (click button)
    │
    ▼
┌─────────────────────┐
│ React Component     │
│ Check permission    │
│ const canEdit =     │
│ hasPermission(...)  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ PermissionGate?     │
│ If yes → render     │
│ If no → fallback    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Service Layer       │
│ Call API endpoint   │
│ Example:            │
│ caseService.update()│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ TanStack Query      │
│ Cache + mutate      │
│ Auto refetch        │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Component Re-render│
│  Update UI state    │
└─────────────────────┘
```

### Frontend Routing & Guards

```
URL Entered
    │
    ▼
┌─────────────────────┐
│ createBrowserRouter │
│ Match route path    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Route Guard       │
│ Example:            │
│ AdminRoute,         │
│ LegalRoute,         │
│ UserRoute, etc.     │
└──────────┬──────────┘
           │
      ┌────┴────┐
      │ Yes     │ No
      ▼         ▼
┌─────────┐  ┌─────────────┐
│ Render  │  │ Redirect to │
│ Page    │  │ role home   │
└─────────┘  └─────────────┘
```

### Available Route Guards

| Guard | Role | Redirect To |
|-------|------|-------------|
| `UserRoute` | `USER` | `/login` if not authenticated |
| `AdminRoute` | `ADMIN` | `/login` if not authenticated |
| `LegalRoute` | `LEGAL` | `/login` if not authenticated |
| `LegalAURoute` | `LEGAL_AU` | `/login` if not authenticated |
| `ExternalRoute` | `EXTERNAL` | `/login` if not authenticated |
| `PrivateRoute` | Any authenticated | `/login` if not authenticated |
| `GuestRoute` | Any authenticated | `/dashboard` or role home |

### Frontend Service Layer Pattern

```typescript
// services/case.service.ts
import api from './api'

export const caseService = {
  getAll: (params) => api.get('/cases', { params }),
  getById: (id) => api.get(`/cases/${id}`),
  create: (data) => api.post('/cases', data),
  update: (id, data) => api.put(`/cases/${id}`, data),
  delete: (id) => api.delete(`/cases/${id}`),
}
```

### Frontend Route Base Pattern

```typescript
const getRoleHome = (role: string) => {
  switch (role) {
    case 'ADMIN': return '/admin'
    case 'LEGAL': return '/legal'
    case 'LEGAL_AU': return '/legal-au'
    case 'EXTERNAL': return '/external/legal-cases'
    default: return '/dashboard'
  }
}
```

---

## Permission & Authorization Flow

### RBAC + PBAC Model

```
User Login
    │
    ▼
┌─────────────────────┐
│ JWT Token Generated │
│ Contains:           │
│ - user_id           │
│ - role              │
│ - permissions[]     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Effective Perms    │
│  = Role Permissions │
│  + ALLOW Overrides  │
│  - DENY Overrides   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Frontend:           │
│ hasPermission(code) │
│ → show/hide UI      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Backend:            │
│ PermissionMiddleware│
│ → 403 if no access  │
└─────────────────────┘
```

### Permission Code Format
```
feature.action.scope

Examples:
- case_management.view          → scope default (all)
- case_management.view.own      → scope own
- legal_opinion.update_status.all → scope all
```

### Permission Seed Assignment

Backend seed assigns baseline permissions per role in `internal/seed/permissions.go`:

| Role | Assign Method |
|------|---------------|
| SUPER_ADMIN | `permissionSeedData()` — all permissions |
| ADMIN | `rolePermissionSeedData()` — all permissions |
| LEGAL | `rolePermissionSeedData()` — specific subset |
| LEGAL_AU | `rolePermissionSeedData()` — limited subset |
| USER | `rolePermissionSeedData()` — own submissions |
| EXTERNAL | `rolePermissionSeedData()` — case management only |

### Override System

Admin dapat mengubah permission user via:

1. **UI**: `UserManagementPage` → Detail User → Tab "Permission"
2. **API**: `PUT /admin/users/:id/permissions`

Override types:
- `DEFAULT` — baseline role permission
- `ALLOW` — tambahkan permission
- `DENY` — cabut permission

---

## Data Flow Examples

### Example 1: User Creates Legal Opinion

```
1. User login → JWT with role=USER
2. Navigate to /dashboard/legal-opinions/new
3. MaterialFormPage checks:
   - hasPermission('legal_opinion.create.own') → true
   - Render form
4. User submits form
5. Frontend calls: POST /api/dashboard/legal-opinions
6. Backend:
   - AuthMiddleware → validate JWT
   - PermissionMiddleware → check 'legal_opinion.create.own'
   - Handler → create DTO
   - Service → create entity
   - Repository → save to DB
   - Response → { id, ticket_number, status: 'PENDING' }
7. Frontend:
   - TanStack Query invalidate
   - Redirect to detail page
```

### Example 2: LEGAL Updates Case Status

```
1. LEGAL login → JWT with role=LEGAL
2. Navigate to /legal/legal-cases/:id
3. AdminLegalCaseDetailPage checks:
   - hasPermission('case_management.update_status') → true
   - Render "Update Status" card
4. LEGAL changes status to "APPROVED"
5. Frontend calls: PATCH /api/legal/legal-cases/:id/status
6. Backend:
   - AuthMiddleware → validate JWT
   - PermissionMiddleware → check 'case_management.update_status'
   - Handler → update status
   - Service → create chronology entry
   - Repository → save + audit log
   - Response → { current_status: 'APPROVED' }
7. Frontend:
   - Update local state
   - Show success toast
```

### Example 3: Admin Override Permission

```
1. ADMIN login
2. Navigate to /admin/users/:id/permissions
3. Admin finds EXTERNAL user
4. Admin grants ALLOW for 'legal_opinion.view.own'
5. Frontend calls: PUT /admin/users/:id/permissions
   Body: { effect: 'ALLOW', permission_id: '<uuid>' }
6. Backend:
   - AuthMiddleware → validate JWT (ADMIN)
   - PermissionMiddleware → check 'user_management.manage_permissions'
   - Handler → upsert permission override
   - Service → log audit (PERMISSION_UPDATE)
7. Next time EXTERNAL user loads page:
   - hasPermission('legal_opinion.view.own') → true (effective permission)
   - UI now shows Legal Opinion menu item
```

---

## Development Conventions

### Naming

| Type | Convention | Example |
|------|-----------|---------|
| Component | PascalCase | `MaterialListPage.tsx` |
| Hook | camelCase + `use` prefix | `useMaterial.ts` |
| Service | camelCase + `Service` suffix | `caseService.ts` |
| Store | camelCase + `Store` suffix | `auth.store.ts` |
| API route | kebab-case path | `/legal-opinions` |
| Permission code | snake_case | `case_management.create` |
| Backend entity | PascalCase singular | `LegalCase`, `User` |
| DB table | snake_case plural | `legal_cases`, `users` |

### Frontend Permission Check Pattern

```tsx
import { useAuthStore } from '@/store/auth.store'

export default function SomePage() {
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const canCreate = hasPermission('feature.create')
  const canEdit = hasPermission('feature.update')
  const canDelete = hasPermission('feature.delete')

  return (
    <div>
      {canCreate && <Button>Tambah</Button>}
      {canEdit && <Button>Edit</Button>}
      {canDelete && <Button>Hapus</Button>}
    </div>
  )
}
```

### Page-Level Permission Gate Pattern

```tsx
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/store/auth.store'
import { getRoleHome } from '@/routes/guards'

export default function SomePage() {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)
  const role = useAuthStore((state) => state.user?.role)

  useEffect(() => {
    if (!hasPermission('feature.view')) {
      navigate(getRoleHome(role), { replace: true })
    }
  }, [hasPermission, navigate, role])

  return <div>...</div>
}
```

---

## Adding a New Feature

### Checklist

1. **Backend**
   - [ ] Create entity in `internal/entity/`
   - [ ] Add migration in `migrations/`
   - [ ] Create DTO in `internal/dto/`
   - [ ] Create repository in `internal/repository/`
   - [ ] Create service in `internal/service/`
   - [ ] Create handler in `internal/handler/`
   - [ ] Register routes in `cmd/api/main.go`
   - [ ] Add permission codes in `internal/seed/permissions.go`
   - [ ] Assign to roles in `rolePermissionSeedData()`

2. **Frontend**
   - [ ] Create pages in appropriate role folder (`admin/`, `legal/`, etc.)
   - [ ] Create service in `services/`
   - [ ] Create hooks in `hooks/`
   - [ ] Create types in `types/`
   - [ ] Register routes in `routes/index.tsx`
   - [ ] Add sidebar menu in `layouts/*Layout.tsx`
   - [ ] Add permission checks with `hasPermission()`

3. **Documentation**
   - [ ] Update `docs/permission_matrix.md` if new permissions added
   - [ ] Update this workflow doc if architecture changes

---

## Quick Reference

### Ports
- Frontend dev: `5173`
- Backend API: `8080`
- PostgreSQL: `5432`
- MinIO: `9000` (API), `9001` (Console)

### Environment Setup
```bash
cp .env.example .env
docker compose up -d
cd backend && go run ./cmd/api
cd frontend && npm install && npm run dev
```

### Database Migrations
- Auto-migrated by GORM on startup (`cmd/api/main.go`)
- Manual migrations in `backend/migrations/`

### Seed Data
- Permissions auto-seed on first run
- Admin credentials from `.env`: `ADMIN_EMAIL`, `ADMIN_PASSWORD`
