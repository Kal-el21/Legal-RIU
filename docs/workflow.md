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
│   │   └── main.go             # Bootstrap: config, DB, MinIO, migrasi, server start
│   └── seed-admin/             # CLI seed untuk buat admin default
├── internal/
│   ├── assets/templates/       # Template dokumen perjanjian bawaan
│   ├── config/                 # Environment config (godotenv) + koneksi database
│   ├── dto/                    # Request/Response DTOs
│   ├── entity/                 # Database models (GORM)
│   ├── handler/                # HTTP handlers (controllers)
│   ├── middleware/             # Auth, role, permission, audit log, rate limit, CSRF
│   ├── repository/             # Database queries (repository pattern)
│   ├── router/                 # Router setup: semua route dan middleware chain
│   ├── seed/                   # Migrasi skema, katalog permission, data master
│   │   └── data/
│   ├── service/                # Business logic
│   ├── storage/                # MinIO file storage
│   └── utils/                  # Helpers (JWT, response, nomor tiket, Excel)
└── go.mod
```

### Request Flow (Backend)

```
HTTP Request
    │
    ▼
┌─────────────────────┐
│  router.go: Router  │
│   Prefix /api/v1    │
│   Group: /admin,    │
│   /legal, /legal-au,│
│   /external         │
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
│ RequirePermission    │
│ (dipasang per route) │
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
| `cmd/api/main.go` | Server bootstrap (config, database, MinIO, migrasi) |
| `internal/router/router.go` | Router setup, middleware chain |
| `internal/middleware/auth.middleware.go` | `AuthMiddleware` (JWT validation, user context injection) dan `RoleMiddleware` (role-based access control) |
| `internal/middleware/permission.middleware.go` | `PermissionContextMiddleware` memuat permission efektif user, `RequirePermission` menolak dengan 403 |
| `internal/middleware/audit_log.middleware.go` | Mencatat aksi ke audit log |
| `internal/middleware/rate_limit.go` | Batas percobaan login |
| `internal/middleware/security.go` | CSRF dan security headers |
| `internal/handler/*.handler.go` | HTTP handlers per module |
| `internal/service/*.service.go` | Business logic layer |
| `internal/repository/*.repository.go` | Database access layer |
| `internal/entity/*.go` | GORM models |
| `internal/seed/schema.go` | `RunAllMigrationsAndSeeds`: AutoMigrate + migrasi manual + seed |
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
│   ├── features/           # Komponen khusus satu fitur (dashboard, legal-opinion, review-document)
│   ├── shared/             # Shared page-level components
│   │   ├── MaterialListPage.tsx
│   │   ├── LegalOpinionListPage.tsx
│   │   └── ReviewDocumentListPage.tsx
│   └── ui/                 # Komponen dasar shadcn
├── constants/              # App-wide constants
├── hooks/                  # React Query hooks per module
├── layouts/                # Layout components per role
│   ├── AdminLayout.tsx
│   ├── LegalLayout.tsx
│   ├── LegalAULayout.tsx
│   ├── DashboardLayout.tsx
│   ├── ExternalLayout.tsx
│   └── PublicLayout.tsx
├── lib/                    # Utils, formatters
├── pages/
│   ├── admin/              # ADMIN role pages
│   ├── dashboard/          # USER role pages
│   ├── legal/              # LEGAL role pages
│   ├── legal-au/           # LEGAL_AU role pages
│   ├── external/           # EXTERNAL role pages (some dormant)
│   ├── materials/          # Form materi yang dipakai lintas peran
│   ├── reports/            # Halaman laporan
│   ├── settings/           # Shared settings page
│   ├── auth/               # Login page
│   └── public/             # Public/landing pages
├── routes/                 # Route definitions + guards
│   ├── index.tsx           # Main router
│   ├── guards.tsx          # Role-based route guards
│   └── role-home.ts        # Halaman awal per peran
├── services/               # API service layer
├── store/                  # Zustand auth store
└── types/                  # TypeScript interfaces
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
│ legalCaseService    │
│   .update()         │
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

Semua guard ada di `routes/guards.tsx`. Belum login selalu diarahkan ke `/login`;
sudah login tapi perannya tidak cocok diarahkan ke halaman awal perannya sendiri.

| Guard | Yang boleh masuk | Kalau tidak boleh |
|-------|------------------|-------------------|
| `UserRoute` | `USER` | `/login`, atau role home kalau peran lain |
| `AdminRoute` | `ADMIN` | `/login`, atau role home kalau peran lain |
| `LegalRoute` | `LEGAL` | `/login`, atau role home kalau peran lain |
| `LegalAURoute` | `LEGAL_AU` | `/login`, atau role home kalau peran lain |
| `ExternalRoute` | `EXTERNAL` | `/login`, atau role home kalau peran lain |
| `PrivateRoute` | Siapa pun yang sudah login | `/login` |
| `GuestRoute` | Yang belum login | Role home |

### Frontend Service Layer Pattern

`services/api.ts` adalah instance axios dengan `baseURL: '/api/v1'`. Tiap modul punya
satu file `services/<modul>.service.ts` yang mengembalikan isi `data`, bukan response axios.

```typescript
// services/legal-case.service.ts (disederhanakan)
import api from './api'

export const legalCaseService = {
  getAll: async (params?: LegalCaseFilters) => {
    const res = await api.get<ApiResponse<PaginatedData<LegalCase>>>(getLegalCaseRouteBase(), { params })
    return res.data.data!
  },
}
```

Modul yang dipakai beberapa peran memilih prefix endpoint lewat fungsi route base
(mis. `getLegalCaseRouteBase()`), karena backend memisahkan route per peran.

### Frontend Route Base Pattern

```typescript
// routes/role-home.ts
export function getRoleHome(role?: UserRole) {
  if (role === 'ADMIN') return '/admin'
  if (role === 'LEGAL') return '/legal'
  if (role === 'LEGAL_AU') return '/legal-au'
  if (role === 'EXTERNAL') return '/external/legal-cases'
  return '/dashboard'
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
│ - sub (user id)     │
│ - email             │
│ - role              │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ GET /auth/permissions│
│ (tidak ada di JWT)  │
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
│ RequirePermission   │
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

Katalog permission didefinisikan di `permissionSeedData()`, lalu dibagikan ke tiap role
oleh `rolePermissionSeedData()`. Keduanya ada di `internal/seed/permissions.go`:

| Role | Assign Method |
|------|---------------|
| ADMIN | `rolePermissionSeedData()` — all permissions |
| LEGAL | `rolePermissionSeedData()` — specific subset |
| LEGAL_AU | `rolePermissionSeedData()` — limited subset |
| USER | `rolePermissionSeedData()` — own submissions |
| EXTERNAL | `rolePermissionSeedData()` — case management only |

### Override System

Admin dapat mengubah permission user via:

1. **UI**: `UserManagementPage` → Detail User → Tab "Permission"
2. **API**: `PUT /api/v1/admin/users/:id/permissions`

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
3. LegalOpinionFormPage checks:
   - hasPermission('legal_opinion.create.own') → true
   - Render form
4. User submits form
5. Frontend calls: POST /api/v1/legal-opinions
6. Backend:
   - AuthMiddleware → validate JWT
   - RoleMiddleware("USER")
   - RequirePermission → check 'legal_opinion.create.own'
   - Handler → create DTO
   - Service → create entity; nomor tiket berikutnya diambil lewat denganNomorTiket
     (diulang dengan nomor baru, maksimal 3 kali, kalau bentrok dengan pengajuan bersamaan)
   - Repository → save to DB; lampiran ke MinIO
   - Service → lampiran langsung dicatat di Repositori Dokumen (muat → SyncLegalOpinion)
   - AuditMiddleware → catat aksi dari SetAuditContext di handler
   - Response → { id, ticket_number, status: 'SUBMITTED' }
7. Frontend:
   - TanStack Query invalidate
   - Redirect to detail page
```

### Example 1b: User Menyimpan Draf lalu Mengirimnya

```
1. Di form pengajuan (mis. /dashboard/legal-opinions/new), user klik "Simpan draf"
   - Validasi form dilewati; isian boleh belum lengkap
2. Frontend calls: POST /api/v1/legal-opinions/drafts (multipart, lampiran boleh)
   - RequirePermission → 'legal_opinion.create.own'
   - Service → status DRAFT, nomor sementara DRAF-<uuid>, lampiran ke MinIO
3. Frontend pindah ke /dashboard/legal-opinions/drafts/:draftId
   - Simpan lagi → PUT /api/v1/legal-opinions/:id/draft (isian diganti, lampiran baru ditambah)
   - Hapus lampiran → DELETE /api/v1/legal-opinions/:id/attachments/:attachmentId
4. User klik "Ajukan" → simpan draf terakhir, lalu POST /api/v1/legal-opinions/:id/submit
   - Validasi penuh (sama dengan pembuatan langsung)
   - Nomor tiket asli, status SUBMITTED, tanggal pengajuan = waktu kirim
   - Audit log seperti pembuatan pengajuan
5. Draf yang dibuang → DELETE /api/v1/legal-opinions/:id (hapus permanen beserta file)
```

Draf hanya terlihat pemiliknya; daftar approver, dashboard, laporan, notifikasi, dan Repositori Dokumen tidak
memuatnya. Perjanjian memakai pola yang sama di `/agreement-documents`: draf selalu mengikuti versi template aktif,
dan versi yang usang dijawab 409 supaya wizard memindahkan isian. Rinciannya di `docs/permission_matrix.md` bagian 2.14.

### Example 2: LEGAL_AU Updates Case Status

```
1. LEGAL_AU login → JWT with role=LEGAL_AU
2. Navigate to /legal-au/cases/:id (LegalAUCaseDetailPage)
3. Request: PATCH /api/v1/legal-au/cases/:id/status
   Body: { status: '<status baru>' }
4. Backend:
   - AuthMiddleware → validate JWT
   - RoleMiddleware("LEGAL_AU")
   - RequirePermission → check 'case_management.update_status'
   - Handler → pastikan kasus milik perusahaan user (403 kalau bukan)
   - Service → UpdateStatus
   - Repository → simpan status + waktu perubahan status
   - Audit log → STATUS_CHANGE pada 'legal_case'
   - Response → "Status kasus berhasil diupdate"
```

Catatan: endpoint `/status` untuk perkara hanya ada di grup `/legal-au`. Peran lain
(ADMIN, LEGAL, USER, EXTERNAL) mengubah perkara lewat `PUT /legal-cases/:id` dengan
permission `case_management.update`.

### Example 3: Admin Override Permission

```
1. ADMIN login
2. Navigate to /admin/users/:id/permissions
3. Admin finds EXTERNAL user
4. Admin grants ALLOW for 'legal_opinion.view.own'
5. Frontend calls: PUT /api/v1/admin/users/:id/permissions
   Body: { effect: 'ALLOW', permission_id: '<uuid>' }
6. Backend:
   - AuthMiddleware → validate JWT (ADMIN)
   - RequirePermission → check 'user_management.manage_permissions'
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
| Service | kebab-case + `.service.ts`, objeknya camelCase + `Service` | `legal-case.service.ts` → `legalCaseService` |
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
import { getRoleHome } from '@/routes/role-home'

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
   - [ ] Daftarkan entity di `db.AutoMigrate(...)` dalam `internal/seed/schema.go`; kalau perlu rename/backfill/NOT NULL, tulis fungsi migrasi di file yang sama dan panggil dari `RunAllMigrationsAndSeeds`
   - [ ] Create DTO in `internal/dto/`
   - [ ] Create repository in `internal/repository/`
   - [ ] Create service in `internal/service/`
   - [ ] Create handler in `internal/handler/`
   - [ ] Register routes in `internal/router/router.go`
   - [ ] Add permission codes in `internal/seed/permissions.go`
   - [ ] Assign to roles in `rolePermissionSeedData()`
   - [ ] Aksi yang mengubah data memanggil `middleware.SetAuditContext` (atau `SetAuditContextWithValues` untuk nilai
         lama dan baru) di handler, dan tesnya memeriksa baris `audit_logs` yang terbentuk
   - [ ] Query lewat `db.Table(...)` atau SQL mentah menambah `deleted_at IS NULL` sendiri, dan query atas tabel
         pengajuan mengecualikan draf (`bukanDraf`)
   - [ ] Aksi yang menambah atau menghapus file mencatatnya di Repositori Dokumen saat itu juga (fungsi `Sync*` dan
         `Remove*` di `internal/service/document_repository.service.go`), dengan isi yang sama dengan backfill di
         `internal/seed/repository_document_seed.go`, dan sumber barunya punya aturan akses
         (`docs/case_company_access.md` bagian 9)

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
- Tidak ada file SQL. `cmd/api/main.go` memanggil `seed.RunAllMigrationsAndSeeds` setiap server start
- Skema dibuat GORM AutoMigrate; daftar entity-nya ada di `internal/seed/schema.go`
- Perubahan yang tidak bisa ditangani AutoMigrate (rename kolom, backfill, NOT NULL) ditulis sebagai fungsi di `internal/seed/schema.go` dan harus aman dijalankan berulang kali

### Seed Data
- Permissions auto-seed on first run
- Admin credentials from `.env`: `ADMIN_EMAIL`, `ADMIN_PASSWORD`
