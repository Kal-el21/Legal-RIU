# Permission Matrix - Legal RIU Portal

## 1. Mekanisme Permission (RBAC + PBAC)

### 1.1 Prinsip Dasar
Sistem menggunakan kombinasi Role-Based Access Control (RBAC) dan Permission-Based Access Control (PBAC):

- **Role** menentukan baseline permission yang dimiliki user.
- **Permission Matrix** (per-user override) memungkinkan Admin menambahkan (`ALLOW`) atau mencabut (`DENY`) permission secara granular tanpa mengubah role user.
- Hasil akhir disebut **Effective Permissions** = Role Permissions + Overrides (ALLOW/DENY).

### 1.2 Aturan Admin
- Role **ADMIN memiliki akses penuh (full access)** ke seluruh sistem.
- Admin tidak perlu diatur melalui Permission Matrix.
- Hanya **non-Admin** yang hak aksesnya ditentukan melalui Permission Matrix.

### 1.3 Struktur Permission
Format kode permission: `feature.action.scope`

| Komponen | Contoh | Keterangan |
|----------|--------|-----------|
| feature | `case_management` | Nama modul/fitur |
| action | `create` | Jenis operasi |
| scope | `own` / `all` | Cakupan akses |

### 1.4 Tipe Override
| Effect | Deskripsi |
|--------|-----------|
| `DEFAULT` | Memberikn akses seperti awal |
| `ALLOW` | Menambahkan permission yang belum ada di role |
| `DENY` | Mencabut permission yang ada di role |

---

## 2. Feature Catalog & Aksi Default

Setiap fitur memiliki kumpulan aksi default. Admin dapat memberikan/mencabut aksi ini secara independen per user.

### 2.1 Dashboard
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `dashboard.user.view` | View User Dashboard | Melihat dashboard user |
| `dashboard.legal.view` | View Legal Dashboard | Melihat dashboard legal |
| `dashboard.admin.view` | View Admin Dashboard | Melihat dashboard admin |

### 2.2 Case Management
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `case_management.view` | View Case | Melihat daftar case |
| `case_management.create` | Create Case | Membuat case baru |
| `case_management.update` | Edit Case | Mengedit case |
| `case_management.delete` | Delete Case | Menghapus case |
| `case_management.update_status` | Update Status Case | Mengubah status case |
| `case_management.manage_document` | Manage Documents | Upload/hapus dokumen case |
| `case_management.manage_chronology` | Manage Chronology | Kelola kronologi case |
| `case_management.manage_reference` | Manage References | Kelola cedant & lokasi case |
| `case_management.download` | Download Documents | Mengunduh dokumen case |
| `case_management.view_regencies` | View Regencies | Daftar kabupaten/kota (master data lookup) |
| `case_management.view_cedants` | View Cedants | Daftar pihak terkait (master data lookup) |

### 2.3 Legal Opinion
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `legal_opinion.view.own` | View Own | Melihat legal opinion milik sendiri |
| `legal_opinion.view.all` | View All | Melihat semua legal opinion |
| `legal_opinion.create.own` | Create Own | Membuat pengajuan legal opinion; menyimpan draf baru dan mengirim draf (2.14) |
| `legal_opinion.update.own` | Edit Own | Mengedit pengajuan milik sendiri; mengubah draf dan menghapus lampirannya (2.14) |
| `legal_opinion.delete.own` | Delete Own | Menghapus pengajuan milik sendiri; draf dihapus permanen beserta lampirannya |
| `legal_opinion.resubmit.own` | Resubmit Own | Mengajukan ulang yang ditolak/perlu revisi |
| `legal_opinion.update_status.all` | Update Status | Mengubah status (review/approve/reject) |
| `legal_opinion.upload_result.all` | Upload Result | Mengunggah hasil kajian |
| `legal_opinion.download.all` | Download | Mengunduh dokumen atau PDF |
| `legal_opinion.generate_pdf.all` | Generate PDF | Menghasilkan PDF hasil kajian |

### 2.4 Document Review
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `document_review.view.own` | View Own | Melihat review dokumen milik sendiri |
| `document_review.view.all` | View All | Melihat semua review dokumen |
| `document_review.create.own` | Create Own | Membuat pengajuan review dokumen; menyimpan draf baru dan mengirim draf (2.14) |
| `document_review.update.own` | Edit Own | Mengedit pengajuan milik sendiri; mengubah draf dan menghapus lampirannya (2.14) |
| `document_review.delete.own` | Delete Own | Menghapus pengajuan milik sendiri; draf dihapus permanen beserta lampirannya |
| `document_review.resubmit.own` | Resubmit Own | Mengajukan ulang yang ditolak/perlu revisi |
| `document_review.update_status.all` | Update Status | Mengubah status (review/approve/reject) |
| `document_review.upload_result.all` | Upload Result | Mengunggah hasil review |
| `document_review.download.all` | Download | Mengunduh dokumen review |

### 2.5 User Management
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `user_management.view` | View Users | Melihat daftar user |
| `user_management.create` | Create User | Membuat user baru |
| `user_management.update` | Edit User | Mengedit data user |
| `user_management.update_status` | Update Status User | Mengubah status (aktif/nonaktif) |
| `user_management.reset_password` | Reset Password | Mereset password user |
| `user_management.delete` | Delete User | Menghapus user |
| `user_management.manage_permissions` | Manage Permissions | Mengatur permission override user |

### 2.6 Audit Log
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `audit_log.view` | View Audit Log | Melihat log aktivitas sistem |

### 2.7 Master Data
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `master_data.view` | View Master Data | Melihat master data |
| `master_data.manage` | Manage Master Data | Mengelola master data (create/edit/delete) |
| `company.set_parent.all` | Perusahaan - Ganti Induk | Menjadikan perusahaan lain sebagai induk dan memindah semua LEGAL ke sana. Lihat `docs/case_company_access.md` bagian 8.1 |

Catatan: Master data mencakup Company, Purpose Type, Case Type, Case Category, Document Type, Regency, Cedant, Division.

Catatan `company.set_parent.all`: dipisah dari `master_data.manage` karena dampaknya memindahkan data user, bukan hanya mengubah satu baris master. Endpoint-nya (`GET /admin/companies/:id/make-parent/preview` dan `POST /admin/companies/:id/make-parent`) ada di grup `/admin`, jadi hanya ADMIN yang bisa memakainya. Sesuai aturan 1.2, override tidak berlaku untuk ADMIN, sehingga semua ADMIN bisa mengganti induk.

### 2.8 Notification Settings
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `notification_setting.view` | View Settings | Melihat pengaturan notifikasi |
| `notification_setting.manage` | Manage Settings | Mengubah threshold notifikasi |

Lonceng notifikasi tidak punya kode permission sendiri. Ketiga endpoint-nya (`GET /notifications`, `PATCH /notifications/read`, `PATCH /notifications/read-all`) terbuka bagi siapa pun yang memegang salah satu permission lihat pengajuan (`legal_opinion.view.own`/`.view.all`, `document_review.view.own`/`.view.all`, `agreement_document.view.own`/`.view.all`), apa pun perannya. Isinya mengikuti permission per modul: pemegang `view` menerima kabar status pengajuannya sendiri, pemegang `update_status.all` menerima pengajuan yang masuk dan peringatan. Frontend menyembunyikan lonceng dengan aturan yang sama (`NOTIFICATION_PERMISSIONS` di `frontend/src/lib/reminders.ts`).

### 2.9 Legal Material
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `legal_material.view` | View Materials | Melihat daftar materi legal |
| `legal_material.manage` | Manage Materials | Mengelola materi legal (create/edit/delete) |

### 2.10 Settings (Profile)
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `settings.profile.update` | Update Profile | Mengubah profil sendiri |
| `settings.notifications.update` | Update Notifications | Mengubah preferensi notifikasi |
| `settings.two_fa.manage` | Manage 2FA | Mengaktifkan/nonaktifkan 2FA |

### 2.11 Auth
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `auth.login` | Login | Masuk ke sistem |
| `auth.logout` | Logout | Keluar dari sistem |
| `auth.change_password` | Change Password | Mengubah password sendiri |

---

### 2.12 Perjanjian: jenis dokumen, template, dan Pihak Pertama
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `agreement_type.manage.all` | Jenis Dokumen - Kelola | Membuat, mengubah, dan menonaktifkan jenis dokumen perjanjian beserta field kerjanya (`GET/POST /admin/agreement-types`, `PUT /admin/agreement-types/:code`) |
| `agreement_template.view.all` | Template Perjanjian - Lihat | Melihat versi template, daftar placeholder, unduh, dan pratinjau |
| `agreement_template.manage.all` | Template Perjanjian - Kelola | Mengunggah, membuat versi baru dari DOCX versi terakhir, mengaktifkan, dan menghapus versi template. Hanya ADMIN sejak PR 9 (keputusan T1) |
| `agreement_company_master.view.all` | Pihak Pertama - Lihat | Melihat master Pihak Pertama (`GET /admin/agreement-company-master`). LEGAL membacanya lewat `GET /legal/agreement-company-master` untuk isian awal persetujuan (sejak PR 11). Juga membaca Daftar Penandatangan (`GET /admin/agreement-signatories`, `GET /legal/agreement-signatories`) |
| `agreement_company_master.manage.all` | Pihak Pertama - Kelola | Mengelola master Pihak Pertama dan Daftar Penandatangan (`POST /admin/agreement-signatories`, `PUT` dan `DELETE /admin/agreement-signatories/:id`; perubahannya tercatat di audit log sebagai `SIGNATORY_CHANGE`) |

Catatan `agreement_type.manage.all`: hanya ADMIN (keputusan K3, PR 9). Endpoint-nya di grup `/admin`, jadi override ke peran lain tidak membuka aksesnya. Field yang disimpan di jenis dokumen adalah field kerja; pengajuan memakai salinan field di versi template yang dikuncinya.

### 2.13 Perjanjian: pengajuan
| Action Code | Label | Keterangan |
|-------------|-------|-----------|
| `agreement_document.view.own` | Perjanjian - Lihat Milik Sendiri | Melihat pengajuan sendiri beserta definisi isian versi templatenya (`GET /agreement-documents/:id/schema`) |
| `agreement_document.view.all` | Perjanjian - Lihat Semua | Melihat semua pengajuan beserta definisi isian versinya (alamat `/legal` dan `/admin`) |
| `agreement_document.create.own` | Perjanjian - Buat | Membuat pengajuan dan membentuk pratinjau sebelum dikirim (`POST /agreement-documents/preview-draft`); menyimpan draf baru dan mengirim draf (2.14) |
| `agreement_document.update.own` | Perjanjian - Edit | Mengubah pengajuan sendiri yang masih Diajukan atau Perlu Revisi; mengubah draf dan menghapus lampirannya (2.14) |
| `agreement_document.delete.own` | Perjanjian - Hapus | Menghapus pengajuan sendiri yang masih Diajukan; draf dihapus permanen beserta lampirannya |
| `agreement_document.resubmit.own` | Perjanjian - Ajukan Ulang | Mengajukan ulang pengajuan yang dikembalikan untuk revisi, beserta lampiran baru |
| `agreement_document.preview.own` | Perjanjian - Preview Sendiri | Pratinjau PDF pengajuan sendiri |
| `agreement_document.preview.all` | Perjanjian - Preview Semua | Pratinjau PDF semua pengajuan |
| `agreement_document.download_attachment.own` | Perjanjian - Attachment Sendiri | Mengunduh lampiran pengajuan sendiri |
| `agreement_document.download_attachment.all` | Perjanjian - Semua Attachment | Mengunduh lampiran semua pengajuan |
| `agreement_document.update_meta.all` | Perjanjian - Review Metadata | Mengisi data persetujuan: nomor perjanjian, tanggal dan tempat tanda tangan, penandatangan Pihak Pertama |
| `agreement_document.update_status.all` | Perjanjian - Ubah Status | Memeriksa, mengembalikan untuk revisi, menolak, dan menyetujui (membentuk dokumen final) |
| `agreement_document.download_pdf.own` | Perjanjian - PDF Sendiri | Mengunduh PDF final pengajuan sendiri |
| `agreement_document.download_pdf.all` | Perjanjian - Semua PDF | Mengunduh PDF final semua pengajuan |
| `agreement_document.download_docx.all` | Perjanjian - DOCX | Mengunduh DOCX final |

### 2.14 Draf pengajuan (Legal Opinion, Review Dokumen, Perjanjian)

Draf adalah pengajuan berstatus `DRAFT` yang belum dikirim. **Tidak ada kode permission baru**: alamat draf memakai
`create.own`, `update.own`, dan `delete.own` modulnya, dan hanya ada di grup USER. Secara bawaan hanya peran USER yang
punya `create.own`.

| Alamat (contoh Legal Opinion; Review Dokumen di `/review-documents`, Perjanjian di `/agreement-documents`) | Permission | Fungsi |
|---|---|---|
| `GET /legal-opinions?status=DRAFT` | `view.own` | Daftar draf milik sendiri |
| `POST /legal-opinions/drafts` | `create.own` | Simpan draf baru; isian boleh belum lengkap, lampiran boleh |
| `PUT /legal-opinions/:id/draft` | `update.own` | Ganti isian draf dan tambah lampiran |
| `DELETE /legal-opinions/:id/attachments/:attachmentId` | `update.own` | Hapus satu lampiran draf (permanen) |
| `POST /legal-opinions/:id/submit` | `create.own` | Kirim draf: validasi penuh, nomor tiket, status `SUBMITTED` |
| `DELETE /legal-opinions/:id` | `delete.own` | Untuk draf: hapus permanen beserta lampiran dan filenya |

Aturan akses yang berlaku di backend:
- Draf hanya terlihat pemiliknya. Pengguna lain, termasuk yang punya `view.all`, mendapat 404 untuk detail, unduhan,
  PDF, pratinjau, dan skema draf. `?status=DRAFT` selalu hanya memuat draf milik sendiri.
- Draf tidak muncul di daftar biasa, dashboard, laporan, lonceng notifikasi, maupun Repositori Dokumen.
- Alamat approver (`update_status.all`, `upload_result.all`, `update_meta.all`) menjawab draf dengan "pengajuan tidak
  ditemukan".
- Nomor tiket asli dan tanggal pengajuan baru diberikan saat draf dikirim. Selama draf, nomornya sementara (`DRAF-<uuid>`).

## 3. Default Permission by Role (Baseline)

Tabel ini menunjukkan permission default yang dimiliki setiap role **sebelum override**. Admin selalu memiliki semua permission secara efektif.

| Fitur | Action | ADMIN | LEGAL | LEGAL_AU | USER | EXTERNAL |
|-------|--------|-------|-------|----------|------|----------|
| **Dashboard** | View User Dashboard | ✅ | ❌ | ❌ | ✅ | ❌ |
| | View Legal Dashboard | ✅ | ✅ | ❌ | ❌ | ❌ |
| | View Admin Dashboard | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Case Management** | View | ✅ | ✅ | ✅ | ❌ | ✅ |
| | Create | ✅ | ✅ | ✅ | ❌ | ❌ |
| | Edit | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Delete | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Update Status | ✅ | ✅ | ✅ | ❌ | ❌ |
| | Manage Documents | ✅ | ✅ | ✅ | ❌ | ✅ |
| | Manage Chronology | ✅ | ✅ | ✅ | ❌ | ✅ |
| | Manage References | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Download | ✅ | ✅ | ✅ | ❌ | ✅ |
| | View Regencies | ✅ | ✅ | ✅ | ❌ | ✅ |
| | View Cedants | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Legal Opinion** | View Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | View All | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Create Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Edit Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Delete Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Resubmit Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Update Status | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Upload Result | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Download | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Generate PDF | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Document Review** | View Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | View All | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Create Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Edit Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Delete Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Resubmit Own | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Update Status | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Upload Result | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Download | ✅ | ✅ | ❌ | ❌ | ❌ |
| **User Management** | View Users | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Create User | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Edit User | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Update Status | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Reset Password | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Delete User | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Manage Permissions | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Audit Log** | View | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Master Data** | View | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Manage | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Ganti Perusahaan Induk | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Notifications** | View Settings | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Manage Settings | ✅ | ❌ | ❌ | ❌ | ❌ |
| | View Reminders | ✅ | ✅ | ❌ | ✅ | ❌ |
| | Mark Read | ✅ | ✅ | ❌ | ✅ | ❌ |
| | Mark All Read | ✅ | ✅ | ❌ | ✅ | ❌ |
| **Legal Material** | View | ✅ | ✅ | ✅ | ❌ | ❌ |
| | Manage | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Settings** | Update Profile | ✅ | ❌ | ❌ | ✅ | ✅ |
| | Update Notifications | ✅ | ❌ | ❌ | ✅ | ✅ |
| | Manage 2FA | ✅ | ❌ | ❌ | ✅ | ✅ |
| **Auth** | Change Password | ✅ | ❌ | ❌ | ✅ | ✅ |
| **Perjanjian** | Kelola Jenis Dokumen | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Lihat Template | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Kelola Template | ✅ | ❌ | ❌ | ❌ | ❌ |
| | Lihat Pihak Pertama | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Kelola Pihak Pertama | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Pengajuan Perjanjian** | Lihat Milik Sendiri, Buat, Edit, Hapus, Ajukan Ulang | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Preview, Attachment, dan PDF Milik Sendiri | ✅ | ❌ | ❌ | ✅ | ❌ |
| | Lihat Semua, Preview Semua, Semua Attachment, Semua PDF | ✅ | ✅ | ❌ | ❌ | ❌ |
| | Review Metadata, Ubah Status, DOCX | ✅ | ✅ | ❌ | ❌ | ❌ |

---

## 4. Contoh Skenario Permission Override

Berikut contoh penerapan Permission Matrix untuk user dengan role yang sama:

### Skenario 1: User A (Role: USER)
- **Case Management**: View Case ✅, Create Case ✅, Edit Case ✅, Delete Case ❌
- **Legal Opinion**: View Own ✅, Create Own ✅, Resubmit Own ✅
- **Document Review**: View Own ✅, Create Own ✅

Cara implementasi:
- Role USER default sudah ada `case_management.create.own` — namun karena case management untuk USER tidak ada di baseline, Admin memberikan override `ALLOW` untuk:
  - `case_management.view`
  - `case_management.create`
  - `case_management.update`
- Admin memberikan override `DENY` untuk:
  - `case_management.delete`

### Skenario 2: User B (Role: USER)
- **Case Management**: Hanya View Case
- **Legal Opinion**: View Own, Create Own

Cara implementasi:
- Admin memberikan override `ALLOW` untuk `case_management.view`
- Tidak ada override untuk create/update/delete case

### Skenario 3: User C (Role: EXTERNAL)
- **Legal Opinion**: View Own, Approve (Update Status) ✅
- **Case Management**: View, Manage Document, Manage Chronology

Cara implementasi:
- Admin memberikan override `ALLOW` untuk:
  - `legal_opinion.view.own`
  - `legal_opinion.update_status.all` (untuk Approve/Reject)

---

## 5. Penjelasan Kelanjutan Sistem Permission

### 5.1 Arsitektur Saat Ini
Sistem saat ini sudah mendukung granular permission melalui:

1. **Permission Catalog** — Daftar semua permission yang tersedia (fitur, action, scope, label).
2. **Role Permissions** — Permission baseline yang dimiliki setiap role.
3. **User Permission Overrides** — Tambahan atau pembatalan permission per user (ALLOW/DENY).
4. **Effective Permissions** — Hasil akhir gabungan role permissions dan overrides.

### 5.2 Mekanisme Override
```
Effective Permissions = Role Permissions 
                        + ALLOW Overrides 
                        - DENY Overrides
```

Contoh kasus:
- User dengan role LEGAL memiliki `case_management.update` secara default.
- Jika Admin memberikan override `DENY` untuk `case_management.update`, maka user tersebut **tidak bisa** mengedit case.
- Jika Admin memberikan override `ALLOW` untuk `case_management.delete` ke user dengan role USER, maka user tersebut **bisa** menghapus case.

### 5.3 API untuk Manajemen Permission
Admin dapat mengelola permission user melalui endpoint:

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| GET | `/admin/permissions` | Menampilkan katalog semua permission |
| GET | `/admin/users/:id/permissions` | Melihat permission akses user |
| PUT | `/admin/users/:id/permissions` | Mengupdate permission override user |

### 5.4 Implementasi di Frontend
Frontend sudah menggunakan `hasPermission(code)` untuk menampilkan/menyembunyikan UI berdasarkan effective permissions:

```typescript
const hasPermission = useAuthStore((state) => state.hasPermission)
const canEdit = hasPermission('case_management.update')
```

### 5.5 Catatan Pengembangan ke Depan
- Saat ini scope hanya `.own` dan `.all`. Untuk menambah scope baru (misal `.company`), tambahkan permission code di backend dan sesuaikan logic pengecekan di frontend.
- Jika ingin menambah fitur baru seperti **Handover**, buat permission codes baru di `backend/internal/seed/permissions.go` dan tambahkan routes + handler sesuai kebutuhan.
- Override bersifat **transaksional**, setiap perubahan tercatat di Audit Log dengan action `PERMISSION_UPDATE`.

---

## 6. Analisis Kelengkapan Permission per Fitur

Bagian ini menjelaskan status implementasi permission untuk setiap fitur, termasuk:
- ✅ **Sudah ada**: Permission code sudah didefinisikan dan digunakan di backend/frontend.
- ⚠️ **Belum ada di code, perlu ditambahkan**: Permission code belum ada di seed, route, atau UI.
- 🔄 **Parsial**: Sudah ada sebagian, tetapi aksi tertentu masihgunakan akses role biasa tanpa pengecekan permission granular.

### 6.1 Case Management
| Aksi | Status | Detail |
|------|--------|--------|
| View | ✅ | `case_management.view` |
| Create | ✅ | `case_management.create` |
| Edit | ✅ | `case_management.update` |
| Delete | ✅ | `case_management.delete` |
| Update Status | ✅ | `case_management.update_status` |
| Manage Document | ✅ | `case_management.manage_document` |
| Manage Chronology | ✅ | `case_management.manage_chronology` |
| Manage Reference (Cedant/Lokasi) | ✅ | `case_management.manage_reference` |
| Download | ✅ | `case_management.download` |
| View Regencies | ⚠️ | Belum ada permission code khusus, saat ini route-nya terbuka untuk user dengan akses case management |
| View Cedants | ⚠️ | Belum ada permission code khusus, saat ini route-nya terbuka untuk user dengan akses case management |
| **Kekurangan / Belum Ada** | | |
| `case_management.approve` | ⚠️ | Belum ada. Saat ini approve/reject dilakukan via `update_status`, tidak ada permission terpisah untuk aksi approve vs reject. |
| `case_management.assign` | ⚠️ | Belum ada. Saat ini tidak ada mechanism assign PIC atau handler khusus untuk case. |
| `case_management.export` | ⚠️ | Belum ada export data case (misal: export Excel/PDF list case). |
| `case_management.import` | ⚠️ | Belum ada import data case. |
| Handover | ❌ | Fitur handover belum ada di sistem sama sekali. Perlu dibuat modul, entity, handler, route, dan permission codesnya terlebih dahulu. |

### 6.2 Legal Opinion
| Aksi | Status | Detail |
|------|--------|--------|
| View Own | ✅ | `legal_opinion.view.own` |
| View All | ✅ | `legal_opinion.view.all` |
| Create Own | ✅ | `legal_opinion.create.own` |
| Edit Own | ✅ | `legal_opinion.update.own` |
| Delete Own | ✅ | `legal_opinion.delete.own` |
| Resubmit Own | ✅ | `legal_opinion.resubmit.own` |
| Update Status | ✅ | `legal_opinion.update_status.all` |
| Upload Result | ✅ | `legal_opinion.upload_result.all` |
| Download | ✅ | `legal_opinion.download.all` |
| Generate PDF | ✅ | `legal_opinion.generate_pdf.all` |
| **Kekurangan / Belum Ada** | | |
| `legal_opinion.approve` | ⚠️ | Belum ada permission khusus approve. Saat ini menggunakan `update_status.all` yang juga digunakan untuk reject. |
| `legal_opinion.reject` | ⚠️ | Sama seperti approve, belum terpisah. |
| `legal_opinion.assign` | ⚠️ | Belum ada. Tidak ada mechanism untuk assign legal opinion ke特定的legal officer. |
| `legal_opinion.export` | ⚠️ | Belum ada export data legal opinion. |
| `legal_opinion.import` | ⚠️ | Belum ada import data legal opinion. |

### 6.3 Document Review
| Aksi | Status | Detail |
|------|--------|--------|
| View Own | ✅ | `document_review.view.own` |
| View All | ✅ | `document_review.view.all` |
| Create Own | ✅ | `document_review.create.own` |
| Edit Own | ✅ | `document_review.update.own` |
| Delete Own | ✅ | `document_review.delete.own` |
| Resubmit Own | ✅ | `document_review.resubmit.own` |
| Update Status | ✅ | `document_review.update_status.all` |
| Upload Result | ✅ | `document_review.upload_result.all` |
| Download | ✅ | `document_review.download.all` |
| **Kekurangan / Belum Ada** | | |
| `document_review.approve` | ⚠️ | Belum ada permission khusus approve. Saat ini menggunakan `update_status.all`. |
| `document_review.reject` | ⚠️ | Belum terpisah. |
| `document_review.assign` | ⚠️ | Belum ada. |
| `document_review.export` | ⚠️ | Belum ada export data review dokumen. |
| `document_review.import` | ⚠️ | Belum ada import data review dokumen. |

### 6.4 User Management
| Aksi | Status | Detail |
|------|--------|--------|
| View Users | ✅ | `user_management.view` |
| Create User | ✅ | `user_management.create` |
| Edit User | ✅ | `user_management.update` |
| Update Status | ✅ | `user_management.update_status` |
| Reset Password | ✅ | `user_management.reset_password` |
| Delete User | ✅ | `user_management.delete` |
| Manage Permissions | ✅ | `user_management.manage_permissions` |
| **Kekurangan / Belum Ada** | | |
| `user_management.export` | ⚠️ | Belum ada export daftar user. |
| `user_management.import` | ⚠️ | Belum ada import daftar user. |

### 6.5 Audit Log
| Aksi | Status | Detail |
|------|--------|--------|
| View | ✅ | `audit_log.view` |
| **Kekurangan / Belum Ada** | | |
| `audit_log.export` | ⚠️ | Belum ada export audit log. |
| `audit_log.filter` | ⚠️ | Saat ini filter sudah ada di UI, tetapi tidak ada permission khusus untuk akses filter lanjutan. |

### 6.6 Master Data
| Aksi | Status | Detail |
|------|--------|--------|
| View | ✅ | `master_data.view` |
| Manage (CRUD) | ✅ | `master_data.manage` |
| Ganti perusahaan induk | ✅ | `company.set_parent.all`, dijaga di route (`requirePermission`) dan di menu Perusahaan (`hasPermission`) |
| **Kekurangan / Belum Ada** | | |
| Route tambah, ubah, dan hapus perusahaan | ⚠️ | Hanya memeriksa peran ADMIN, tidak memeriksa `master_data.manage`. Karena override tidak berlaku untuk ADMIN, hasilnya sama untuk saat ini, tetapi penjagaannya tidak mengikuti katalog di atas |
| `master_data.export` | ⚠️ | Belum ada export master data per entitas. |
| `master_data.import` | ⚠️ | Belum ada import master data. |

### 6.7 Notification Settings
| Aksi | Status | Detail |
|------|--------|--------|
| View Settings | ✅ | `notification_setting.view` (implied dari admin access) |
| Manage Settings | ✅ | `notification_setting.manage` |
| View Reminders | ✅ | `GET /notifications`, mengikuti permission lihat Legal Opinion atau Review Dokumen (lihat 2.8) |
| Mark Read | ✅ | `PATCH /notifications/read`, penjaga yang sama |
| Mark All Read | ✅ | `PATCH /notifications/read-all`, penjaga yang sama |
| **Kekurangan / Belum Ada** | | |
| `notification.export_reminders` | ⚠️ | Belum ada export daftar reminders. |

### 6.8 Legal Material
| Aksi | Status | Detail |
|------|--------|--------|
| View | ✅ | `legal_material.view` |
| Manage (CRUD) | ✅ | `legal_material.manage` |
| **Kekurangan / Belum Ada** | | |
| `legal_material.export` | ⚠️ | Belum ada export materi legal. |
| `legal_material.import` | ⚠️ | Belum ada import materi legal. |

### 6.9 Dashboard
| Aksi | Status | Detail |
|------|--------|--------|
| User Dashboard | ✅ | `dashboard.user.view` |
| Legal Dashboard | ✅ | `dashboard.legal.view` |
| Admin Dashboard | ✅ | `dashboard.admin.view` |
| **Kekurangan / Belum Ada** | | |
| `dashboard.legal_au.view` | ⚠️ | Belum ada. LEGAL_AU saat ini tidak memiliki dashboard khusus; home-nya langsung ke list case. |
| `dashboard.export` | ⚠️ | Belum ada export data dashboard (misal: export statistik). |

### 6.10 Settings (Profile)
| Aksi | Status | Detail |
|------|--------|--------|
| Update Profile | ✅ | `settings.profile.update` |
| Update Notifications | ✅ | `settings.notifications.update` |
| Manage 2FA | ✅ | `settings.two_fa.manage` |
| **Kekurangan / Belum Ada** | | |
| `settings.manage_for_user` | ⚠️ | Belum ada. Admin belum bisa mengubah settings atas nama user lain. |

### 6.11 Auth
| Aksi | Status | Detail |
|------|--------|--------|
| Login | ✅ | `auth.login` |
| Logout | ✅ | `auth.logout` |
| Change Password | ✅ | `auth.change_password` |
| **Kekurangan / Belum Ada** | | |
| `auth.impersonate` | ⚠️ | Belum ada fitur impersonate (login sebagai user lain untuk troubleshooting). |

### 6.12 Handover
| Aksi | Status | Detail |
|------|--------|--------|
| Modul Handover | ❌ | Fitur ini belum ada di sistem. Perlu definisi requirement terlebih dahulu. |
| **Kekurangan / Belum Ada** | | |
| `handover.view` | ❌ | Perlu dibuat. |
| `handover.create` | ❌ | Perlu dibuat. |
| `handover.edit` | ❌ | Perlu dibuat. |
| `handover.delete` | ❌ | Perlu dibuat. |
| `handover.approve` | ❌ | Perlu dibuat. |
| `handover.reject` | ❌ | Perlu dibuat. |
| `handover.assign` | ❌ | Perlu dibuat. |
| `handover.export` | ❌ | Perlu dibuat. |
| `handover.import` | ❌ | Perlu dibuat. |

---

## 7. Rekomendasi Implementasi Lanjutan

### 7.1 Prioritas Tinggi
1. **Handover Module** — Jika fitur ini sudahFinal requirement, segera buat entity, service, handler, route, dan permission codesnya.
2. **Dashboard LEGAL_AU** — Tambahkan `dashboard.legal_au.view` dan buat halaman dashboard khusus LEGAL_AU jika dibutuhkan.
3. **Approve/Reject Permission Terpisah** — Pisahkan permission `approve` dan `reject` dari `update_status` agar admin bisa mengontrol siapa yang bisa approve tanpa bisa reject, atau sebaliknya.

### 7.2 Prioritas Medium
4. **Export Permission** — Tambahkan permission `export` untuk setiap modul yang membutuhkan export data.
5. **Import Permission** — Tambahkan permission `import` untuk fitur import data.
6. **Assign Permission** — Tambahkan permission `assign` untuk modul yang memungkinkan penugasan (assignment) tugas/kasus kepada user tertentu.

### 7.3 Prioritas Rendah
7. **Filter Permission** — Jika nanti ada filter sensitif (misal: filter by user, filter by date range yang luas), pertimbangkan permission khusus untuk akses filter lanjutan.
8. **Settings for User** — Tambahkan permission agar admin bisa mengubah settings atas nama user lain.

### 7.4 Catatan Teknis Implementasi
- Semua permission code baru harus ditambahkan di `backend/internal/seed/permissions.go` fungsi `permissionSeedData()`.
- Role baseline baru (jika ada role baru) harus ditambahkan di fungsi `rolePermissionSeedData()`.
- Frontend sudah siap menggunakan permission codes baru — tinggal tambahkan pengecekan `hasPermission('feature.action')` di komponen yang relevan.
- Setiap perubahan permission akan tercatat di Audit Log secara otomatis.
