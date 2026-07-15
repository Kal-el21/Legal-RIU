# Document Generation Perjanjian — Implementation Plan

## 1. Ringkasan Keputusan

Feature yang dibangun adalah **Document Generation** untuk dokumen perjanjian. Implementasi pertama mendukung tipe **Perjanjian Kerja Sama (PKS)**, tetapi struktur backend dan frontend disiapkan agar tipe dokumen lain dapat ditambahkan kemudian dengan form dan template yang berbeda.

Keputusan yang sudah disepakati:

- Requester adalah role `USER`.
- Approver adalah role `ADMIN` dan `LEGAL` dengan kewenangan review yang sama.
- Setiap tipe dokumen hanya memiliki **satu template DOCX aktif**.
- Tidak ada upload template, template versioning, kalibrasi koordinat, atau editor template melalui UI.
- Template dan konfigurasi form ditambahkan developer melalui repository dan perubahan berlaku setelah build/deploy.
- Revisi pengajuan tetap tersedia, tetapi tidak ada revisi/versi template melalui aplikasi.
- Requester hanya dapat melihat preview PDF dan mengunduh PDF final; requester tidak dapat mengakses DOCX.
- ADMIN dan LEGAL dapat mengunduh PDF final dan DOCX hasil generate.
- Requester dapat mengunggah beberapa attachment. Attachment disimpan dan diunduh terpisah dari dokumen perjanjian.
- Bagian III dokumen berisi daftar attachment, bukan penggabungan isi attachment ke PDF final.
- Master Pihak Pertama dikelola ADMIN.
- ADMIN dan LEGAL dapat mengoreksi data Pihak Pertama untuk satu pengajuan tanpa mengubah master global.
- Nomor Pihak Pertama dibuat otomatis dan dapat dikoreksi approver sebelum approval.

## 2. Kondisi Repository Saat Planning

Branch aktif `doc_agree_static` belum memiliki modul Agreement aktif. Baseline yang tersedia dan dapat dijadikan pola adalah:

- workflow `DocumentReview`;
- status submission bersama;
- repository/service/handler pattern;
- MinIO storage;
- audit middleware dan audit log;
- RBAC + PBAC dengan effective permission;
- permission code berbentuk `feature.action.scope`;
- frontend route guard, `hasPermission`, dan `PermissionGate`.

Implementasi Agreement lama tersedia pada branch `doc_agree_dynamic`, terutama pada commit:

```text
bc89418 feat: document aggreement
0d3dacd feat: document agreement dynamic occurence
```

Branch tersebut hanya digunakan sebagai referensi workflow dan UI. Jangan melakukan cherry-pick penuh karena implementasi lama memiliki template versioning, upload template, koordinat PDF, kalibrasi, route `EXTERNAL`, dan komponen lain yang tidak lagi sesuai kebutuhan.

## 3. Sasaran dan Batasan

### 3.1 Sasaran

- USER dapat memilih tipe dokumen dan mengirim pengajuan.
- Form berubah sesuai konfigurasi tipe dokumen.
- ADMIN/LEGAL dapat mereview data dan preview PDF.
- ADMIN/LEGAL dapat mengoreksi nomor perjanjian, tempat/tanggal tanda tangan, pejabat, dan jabatan Pihak Pertama.
- Approval menghasilkan DOCX dan PDF dari sumber dan data yang sama.
- PDF final dapat diunduh USER pemilik pengajuan.
- DOCX hanya dapat diunduh ADMIN/LEGAL.
- Semua akses dilindungi permission dan ownership check.
- Penambahan tipe dokumen baru tidak membutuhkan perubahan tabel untuk setiap field baru.

### 3.2 Di luar cakupan tahap pertama

- Upload atau edit template dari browser.
- Template versioning dan template history pada UI/database.
- Kalibrasi posisi field PDF.
- Menggabungkan attachment ke PDF final.
- Digital signature/e-signature.
- Approval berjenjang antara LEGAL dan ADMIN.
- Akses role `EXTERNAL` atau `LEGAL_AU`.
- Admin form-builder untuk membuat schema secara visual.

## 4. Sumber Template dan Fondasi Multi-Tipe

### 4.1 Struktur sumber

Satu tipe dokumen memiliki satu folder definisi:

```text
.docs/agreement-templates/
└── pks/
    ├── template.docx
    └── form-schema.json
```

Untuk migrasi awal, file berikut dipindahkan menjadi template PKS resmi:

```text
.docs/Draft - PKS RIU - .......docx
```

Tidak boleh ada salinan template sumber lain yang ikut diedit. File di image Docker hanya merupakan build artifact.

### 4.2 Registry tipe dokumen

Backend memiliki registry read-only, misalnya:

```go
type DocumentTypeDefinition struct {
    Code         string
    Name         string
    TemplatePath string
    SchemaPath   string
    Generator    string
}
```

Definisi tahap pertama:

```text
code: PKS
name: Perjanjian Kerja Sama
template: .docs/agreement-templates/pks/template.docx
schema: .docs/agreement-templates/pks/form-schema.json
```

Registry bukan template versioning. Setiap `code` hanya menunjuk satu template aktif.

### 4.3 Penambahan tipe baru

Developer menambahkan:

1. satu folder tipe dokumen;
2. satu `template.docx`;
3. satu `form-schema.json`;
4. mapping placeholder dan business validator bila diperlukan;
5. test fixture untuk tipe tersebut.

Frontend tidak membuat halaman form baru per tipe. Frontend mengambil schema dari backend dan merender field berdasarkan schema.

## 5. Schema Form Dinamis

`form-schema.json` minimal mendukung:

```json
{
  "code": "PKS",
  "name": "Perjanjian Kerja Sama",
  "sections": [
    {
      "title": "Pihak Kedua",
      "fields": [
        {
          "name": "pihak_kedua_nama",
          "label": "Nama Perusahaan",
          "type": "text",
          "required": true,
          "max_length": 255
        }
      ]
    }
  ]
}
```

Tipe field tahap pertama:

- `text`;
- `email`;
- `tel`;
- `date`;
- `integer` atau `money`;
- `decimal` untuk persentase;
- `textarea`;
- `file`/attachment sebagai komponen khusus, bukan bagian `form_data` biasa.

Backend tetap menjadi sumber validasi utama. Validasi frontend hanya untuk pengalaman pengguna.

Schema tidak boleh menentukan permission atau status workflow. Permission dan workflow tetap berada di backend.

## 6. Field PKS

### 6.1 Informasi umum

| Field | Diisi oleh | Wajib | Catatan |
|---|---|---:|---|
| `nomor_pihak_pertama` | Sistem, dapat dikoreksi approver | Ya saat approve | Harus unik |
| `nomor_pihak_kedua` | USER | Tidak | Tampilkan `-` bila kosong |
| `tempat_ttd` | Default master/approver | Ya saat approve | USER boleh mengusulkan bila form membutuhkannya |
| `tanggal_ttd` | Approver | Ya saat approve | Sumber hari, tanggal, bulan, dan tahun |

### 6.2 Pihak Pertama

Data default berasal dari Master Pihak Pertama:

- nama perusahaan;
- alamat;
- NPWP bila kelak dibutuhkan;
- telepon;
- email;
- PIC;
- nama pejabat default;
- jabatan default;
- tempat tanda tangan default.

ADMIN/LEGAL dapat mengubah data berikut hanya untuk satu pengajuan:

- nama pejabat;
- jabatan;
- nomor Pihak Pertama;
- tempat tanda tangan;
- tanggal tanda tangan.

### 6.3 Pihak Kedua

| Field | Wajib |
|---|---:|
| `pihak_kedua_nama` | Ya |
| `pihak_kedua_bidang` | Ya |
| `pihak_kedua_alamat` | Ya |
| `pihak_kedua_telepon` | Tidak |
| `pihak_kedua_email` | Tidak |
| `pihak_kedua_pic` | Tidak |
| `pihak_kedua_pejabat` | Ya |
| `pihak_kedua_jabatan` | Ya |

Template tidak menambah field bentuk badan hukum, NIB, NPWP Pihak Kedua, atau dasar kewenangan. Paragraf identitas Pihak Kedua disusun dari field yang sudah ada.

### 6.4 Pekerjaan dan dokumen dasar

| Field | Wajib |
|---|---:|
| `jenis_pekerjaan` | Ya |
| `ruang_lingkup` | Ya |
| `surat_penawaran_nomor` | Tidak |
| `surat_penawaran_perihal` | Tidak |
| `surat_penawaran_tanggal` | Tidak |
| `surat_penunjukan_nomor` | Tidak |
| `surat_penunjukan_perihal` | Tidak |
| `surat_penunjukan_tanggal` | Tidak |

### 6.5 Jangka waktu dan pembayaran

| Field | Wajib |
|---|---:|
| `jangka_waktu_mulai` | Ya |
| `jangka_waktu_selesai` | Ya |
| `nilai_kontrak` | Ya |
| `termin_1_persen` | Ya |
| `termin_1_nilai` | Ya |
| `termin_2_persen` | Ya |
| `termin_2_nilai` | Ya |
| `bank` | Ya |
| `nomor_rekening` | Ya |
| `atas_nama` | Ya |

Aturan bisnis PKS:

- tanggal selesai tidak boleh sebelum tanggal mulai;
- nilai uang tidak boleh negatif;
- persentase termin tidak boleh negatif atau lebih dari 100;
- jumlah persentase kedua termin harus 100;
- jumlah nilai kedua termin harus sama dengan nilai kontrak, dengan aturan toleransi pembulatan eksplisit;
- nomor rekening disimpan sebagai string agar angka nol di depan tidak hilang;
- nilai uang jangan memakai floating point untuk penyimpanan; gunakan integer satuan Rupiah atau decimal yang presisi.

### 6.6 Attachment

USER dapat mengunggah beberapa attachment, misalnya:

- surat penawaran;
- surat penunjukan;
- proposal;
- TOR/KAK;
- dokumen pendukung lain.

Setiap attachment menyimpan:

- nama file asli;
- object path;
- MIME type;
- ukuran file;
- deskripsi opsional;
- upload round;
- uploader;
- waktu upload.

Daftar nama/deskripsi attachment dimasukkan otomatis ke Bagian III. File tetap disimpan dan diunduh terpisah.

## 7. Kontrak Placeholder PKS

Placeholder menggunakan format:

```text
{{NAMA_PLACEHOLDER}}
```

### 7.1 Informasi umum

```text
{{NOMOR_PIHAK_PERTAMA}}
{{NOMOR_PIHAK_KEDUA}}
{{HARI_TTD}}
{{TANGGAL_TTD}}
{{BULAN_TTD}}
{{TAHUN_TTD}}
{{TANGGAL_TTD_LENGKAP}}
{{TEMPAT_TTD}}
```

### 7.2 Pihak Pertama

```text
{{PIHAK_PERTAMA_NAMA}}
{{PIHAK_PERTAMA_ALAMAT}}
{{PIHAK_PERTAMA_TELEPON}}
{{PIHAK_PERTAMA_EMAIL}}
{{PIHAK_PERTAMA_PIC}}
{{PIHAK_PERTAMA_PEJABAT}}
{{PIHAK_PERTAMA_JABATAN}}
```

### 7.3 Pihak Kedua

```text
{{PIHAK_KEDUA_NAMA}}
{{PIHAK_KEDUA_BIDANG}}
{{PIHAK_KEDUA_ALAMAT}}
{{PIHAK_KEDUA_TELEPON}}
{{PIHAK_KEDUA_EMAIL}}
{{PIHAK_KEDUA_PIC}}
{{PIHAK_KEDUA_PEJABAT}}
{{PIHAK_KEDUA_JABATAN}}
```

### 7.4 Pekerjaan dan dokumen dasar

```text
{{JENIS_PEKERJAAN}}
{{RUANG_LINGKUP}}
{{SURAT_PENAWARAN_NOMOR}}
{{SURAT_PENAWARAN_PERIHAL}}
{{SURAT_PENAWARAN_TANGGAL}}
{{SURAT_PENUNJUKAN_NOMOR}}
{{SURAT_PENUNJUKAN_PERIHAL}}
{{SURAT_PENUNJUKAN_TANGGAL}}
```

### 7.5 Jangka waktu dan pembayaran

```text
{{JANGKA_WAKTU_MULAI}}
{{JANGKA_WAKTU_SELESAI}}
{{NILAI_KONTRAK}}
{{NILAI_KONTRAK_TERBILANG}}
{{TERMIN_1_PERSEN}}
{{TERMIN_1_PERSEN_TERBILANG}}
{{TERMIN_1_NILAI}}
{{TERMIN_1_NILAI_TERBILANG}}
{{TERMIN_2_PERSEN}}
{{TERMIN_2_PERSEN_TERBILANG}}
{{TERMIN_2_NILAI}}
{{TERMIN_2_NILAI_TERBILANG}}
{{BANK}}
{{NOMOR_REKENING}}
{{ATAS_NAMA}}
```

### 7.6 Lampiran

```text
{{DAFTAR_LAMPIRAN}}
```

### 7.7 Aturan generator

- Placeholder yang sama dapat muncul berulang kali.
- Generator harus menangani placeholder yang terpecah menjadi beberapa Word XML run.
- Penggantian dilakukan di body, table, header, dan footer.
- Nilai harus di-escape sebagai XML.
- Nilai multiline harus menghasilkan Word line break yang valid.
- Field opsional kosong menghasilkan `-` atau teks kosong sesuai konteks yang didefinisikan mapper.
- Field wajib kosong menghentikan preview/final generation dengan validation error yang jelas.
- Output tidak boleh menyisakan pola `{{...}}`.
- Template sumber tidak pernah dimodifikasi ketika request berjalan.

## 8. Master dan Snapshot Pihak Pertama

### 8.1 Master global

Sediakan page ADMIN khusus untuk:

- melihat data Pihak Pertama;
- mengubah nama perusahaan dan alamat;
- mengubah telepon, email, dan PIC;
- mengubah pejabat dan jabatan default;
- mengubah tempat tanda tangan default.

Tahap pertama cukup memiliki satu record aktif karena Pihak Pertama adalah PT Reasuransi Indonesia Utama (Persero).

### 8.2 Snapshot per pengajuan

Pengajuan tidak boleh bergantung pada master live setelah approval. Saat approval, simpan snapshot lengkap yang digunakan untuk generate final:

```json
{
  "name": "PT Reasuransi Indonesia Utama (Persero)",
  "address": "...",
  "phone": "...",
  "email": "...",
  "pic": "...",
  "signatory_name": "...",
  "signatory_position": "...",
  "signing_place": "..."
}
```

Perubahan master berikutnya tidak mengubah preview/final dokumen yang sudah `COMPLETED`.

## 9. Model Data

### 9.1 AgreementDocument

Field minimum:

```text
id
ticket_number
requester_id
document_type_code
form_data JSONB
party_one_snapshot JSONB
agreement_number
status
status_updated_at
approver_note
generated_docx_path
generated_pdf_path
generated_file_name
template_checksum
approved_by
approved_at
created_at
updated_at
```

Catatan:

- `document_type_code` memungkinkan tipe baru tanpa tabel pengajuan baru.
- `form_data` menyimpan field khusus tipe dokumen.
- field yang dipakai untuk filter/report dapat ditambahkan sebagai kolom terindeks bila benar-benar dibutuhkan.
- `template_checksum` hanya untuk audit template yang menghasilkan final, bukan template versioning atau pemilihan versi.
- final DOCX dan PDF bersifat immutable setelah approval.

### 9.2 AgreementAttachment

```text
id
agreement_document_id
file_name
file_path
mime_type
file_size
description
upload_round
uploaded_by
created_at
```

### 9.3 CompanyMaster

Gunakan entity terpisah dari entity `Company` yang saat ini dipakai untuk afiliasi user:

```text
id
name
address
npwp
phone
email
pic
default_signatory_name
default_signatory_position
default_signing_place
is_active
created_at
updated_at
```

## 10. Workflow

```text
SUBMITTED
    ├── UNDER_REVIEW
    │      ├── NEED_REVISION ── USER memperbaiki ── RESUBMITTED
    │      ├── REJECTED
    │      └── COMPLETED
    └── dapat diedit/dihapus USER selama belum diambil untuk review
```

Aturan:

- Create menghasilkan `SUBMITTED`.
- USER hanya dapat mengubah pengajuan miliknya pada `SUBMITTED` atau `NEED_REVISION`.
- USER hanya dapat menghapus pengajuan miliknya pada `SUBMITTED`.
- Resubmit hanya dari `NEED_REVISION` dan menghasilkan `RESUBMITTED`.
- `REJECTED` bersifat final. Jika ingin mengajukan kembali, USER membuat pengajuan baru.
- ADMIN/LEGAL dapat memulai review dari `SUBMITTED` atau `RESUBMITTED`.
- ADMIN/LEGAL memiliki kewenangan decision yang sama.
- Approval hanya dari `UNDER_REVIEW`.
- Approval memvalidasi seluruh data, menghasilkan final DOCX dan PDF, menyimpan snapshot, lalu mengubah status menjadi `COMPLETED`.
- Return/reject wajib memiliki catatan approver.
- Endpoint status harus memvalidasi transisi di service, bukan hanya percaya pada UI.

## 11. Nomor Perjanjian

Nomor Pihak Pertama dibuat otomatis dengan pola awal:

```text
001/RM.01.01/HR/IndonesiaRe/2026
```

Aturan:

- sequence harus aman terhadap request bersamaan;
- gunakan unique constraint pada nomor final;
- nomor dapat dikoreksi ADMIN/LEGAL sebelum approval;
- setiap koreksi dicatat di audit log;
- koreksi harus tetap lolos uniqueness validation;
- format nomor disimpan dalam konfigurasi backend agar tidak tersebar di handler/frontend.

## 12. Generation Pipeline

```text
Agreement data + Party One effective data + attachment list
                         ↓
                 Placeholder mapper
                         ↓
             Salinan sementara template DOCX
                         ↓
              DOCX generator/validator
                         ↓
                 LibreOffice converter
                         ↓
           PDF preview atau DOCX + PDF final
```

### 12.1 Preview

- Menggunakan data terbaru dan template aktif.
- Menghasilkan PDF sementara.
- Memiliki watermark `DRAFT`.
- Tidak menyimpan atau mengekspos DOCX kepada USER.
- Dapat dilihat USER pemilik serta ADMIN/LEGAL.

### 12.2 Approval/final

- Mengambil lock atau menggunakan optimistic concurrency agar dua approver tidak menyetujui dokumen yang sama bersamaan.
- Memvalidasi status dan field wajib.
- Membentuk snapshot Pihak Pertama.
- Menghasilkan DOCX final tanpa watermark.
- Mengonversi DOCX yang sama menjadi PDF final tanpa watermark.
- Mengunggah kedua file ke MinIO.
- Menyimpan checksum template dan identitas approver.
- Mengubah status menjadi `COMPLETED` hanya setelah kedua file berhasil dibuat dan disimpan.
- Bila update database gagal setelah upload, service harus membersihkan object yatim atau menyediakan proses cleanup.
- Request approval berulang harus idempotent atau ditolak dengan conflict yang jelas.

## 13. Permission Architecture

### 13.1 Prinsip

Ikuti arsitektur saat ini:

```text
Effective Permissions = Role Permissions + ALLOW Overrides - DENY Overrides
```

- Semua permission ditambahkan ke catalog `backend/internal/seed/permissions.go`.
- ADMIN memperoleh semua permission melalui mekanisme `allCodes` yang sudah ada.
- USER dan LEGAL memperoleh baseline permission sesuai tabel di bawah.
- Middleware permission tetap digunakan pada setiap route.
- Frontend memakai `hasPermission`/`PermissionGate` untuk navigasi dan tombol.
- Pemeriksaan permission tidak menggantikan ownership dan validasi status pada service.
- Tidak ada permission Agreement untuk `EXTERNAL` dan `LEGAL_AU` pada tahap pertama.

### 13.2 Catalog permission Agreement

| Permission | USER | LEGAL | ADMIN | Keterangan |
|---|:---:|:---:|:---:|---|
| `agreement_document.view.own` | Ya | Tidak | Ya | Melihat pengajuan milik sendiri |
| `agreement_document.view.all` | Tidak | Ya | Ya | Melihat seluruh pengajuan |
| `agreement_document.create.own` | Ya | Tidak | Ya | Membuat pengajuan sendiri |
| `agreement_document.update.own` | Ya | Tidak | Ya | Mengubah pengajuan sendiri sesuai status |
| `agreement_document.delete.own` | Ya | Tidak | Ya | Menghapus pengajuan sendiri sesuai status |
| `agreement_document.resubmit.own` | Ya | Tidak | Ya | Mengirim ulang setelah revisi |
| `agreement_document.preview.own` | Ya | Tidak | Ya | Preview PDF milik sendiri |
| `agreement_document.preview.all` | Tidak | Ya | Ya | Preview PDF seluruh pengajuan |
| `agreement_document.upload_attachment.own` | Ya | Tidak | Ya | Upload attachment milik sendiri |
| `agreement_document.download_attachment.own` | Ya | Tidak | Ya | Download attachment milik sendiri |
| `agreement_document.download_attachment.all` | Tidak | Ya | Ya | Download seluruh attachment |
| `agreement_document.update_meta.all` | Tidak | Ya | Ya | Koreksi data review/Pihak Pertama per pengajuan |
| `agreement_document.update_status.all` | Tidak | Ya | Ya | Start review, return, reject, dan approve |
| `agreement_document.download_pdf.own` | Ya | Tidak | Ya | Download PDF final milik sendiri |
| `agreement_document.download_pdf.all` | Tidak | Ya | Ya | Download PDF final seluruh pengajuan |
| `agreement_document.download_docx.all` | Tidak | Ya | Ya | Download DOCX final; tidak diberikan kepada USER |

### 13.3 Permission Master Pihak Pertama

| Permission | USER | LEGAL | ADMIN | Keterangan |
|---|:---:|:---:|:---:|---|
| `agreement_company_master.view.all` | Tidak | Tidak | Ya | Melihat master global Pihak Pertama |
| `agreement_company_master.manage.all` | Tidak | Tidak | Ya | Mengubah master global Pihak Pertama |

LEGAL tidak memperoleh akses page master global. LEGAL hanya melihat effective snapshot/data Pihak Pertama dalam halaman review dan mengubahnya melalui `agreement_document.update_meta.all`.

### 13.4 Defense in depth

- Endpoint DOCX tidak didaftarkan pada group USER.
- Handler download DOCX tetap memeriksa `download_docx.all`.
- Presigned URL DOCX hanya dibuat sesaat setelah authorization berhasil.
- Object path MinIO tidak pernah dikirim sebagai URL publik permanen.
- USER yang memiliki permission `.own` tetap harus cocok dengan `requester_id`.
- Permission override tidak boleh membuat scope `.own` berubah menjadi akses seluruh data.

## 14. Rancangan Route

### 14.1 USER

```text
GET    /agreement-document-types
GET    /agreement-document-types/:code/schema
GET    /agreement-documents
POST   /agreement-documents
GET    /agreement-documents/:id
PUT    /agreement-documents/:id
DELETE /agreement-documents/:id
POST   /agreement-documents/:id/resubmit
POST   /agreement-documents/:id/attachments
GET    /agreement-documents/:id/attachments/:attachmentId
GET    /agreement-documents/:id/preview
GET    /agreement-documents/:id/pdf
```

Tidak ada route DOCX pada group USER.

### 14.2 ADMIN

```text
GET    /admin/agreement-documents
GET    /admin/agreement-documents/:id
PATCH  /admin/agreement-documents/:id/meta
PATCH  /admin/agreement-documents/:id/status
GET    /admin/agreement-documents/:id/preview
GET    /admin/agreement-documents/:id/pdf
GET    /admin/agreement-documents/:id/docx
GET    /admin/agreement-documents/:id/attachments/:attachmentId

GET    /admin/agreement-company-master
PUT    /admin/agreement-company-master
```

### 14.3 LEGAL

```text
GET    /legal/agreement-documents
GET    /legal/agreement-documents/:id
PATCH  /legal/agreement-documents/:id/meta
PATCH  /legal/agreement-documents/:id/status
GET    /legal/agreement-documents/:id/preview
GET    /legal/agreement-documents/:id/pdf
GET    /legal/agreement-documents/:id/docx
GET    /legal/agreement-documents/:id/attachments/:attachmentId
```

Route ADMIN juga harus memakai `requirePermission`; jangan hanya mengandalkan role group walaupun ADMIN secara default full access.

## 15. Backend Components

Komponen baru yang direncanakan:

```text
backend/internal/dto/agreement_document.dto.go
backend/internal/entity/agreement_document.go
backend/internal/repository/agreement_document.repository.go
backend/internal/handler/agreement_document.handler.go
backend/internal/service/agreement_document.service.go
backend/internal/service/agreement_docx_generator.go
backend/internal/service/document_type_registry.go
backend/internal/service/agreement_validator.go
backend/internal/service/docx_converter.go

backend/internal/dto/agreement_company_master.dto.go
backend/internal/repository/agreement_company_master.repository.go
backend/internal/handler/agreement_company_master.handler.go
backend/internal/service/agreement_company_master.service.go
```

Tanggung jawab dipisahkan:

- registry membaca tipe dan schema;
- validator memvalidasi schema serta aturan bisnis per tipe;
- mapper membentuk placeholder;
- generator hanya menghasilkan DOCX;
- converter hanya mengonversi DOCX ke PDF;
- agreement service mengatur workflow, permission-independent domain checks, storage, dan transaksi;
- handler mengatur HTTP binding/response;
- repository mengatur persistence.

## 16. Frontend Components

### 16.1 USER

- menu `Dokumen Perjanjian` dengan permission `agreement_document.view.own`;
- list pengajuan milik sendiri;
- pilihan tipe dokumen;
- dynamic form berdasarkan schema;
- multiple attachment upload;
- detail, status timeline, dan catatan revisi;
- preview PDF watermark;
- download PDF final;
- tidak ada tombol atau API client download DOCX.

### 16.2 ADMIN/LEGAL

- list seluruh pengajuan;
- detail/review dua panel: editor metadata dan PDF preview;
- koreksi nomor, pejabat, jabatan, tempat, dan tanggal;
- akses attachment;
- aksi Start Review, Return, Reject, dan Approve;
- download PDF dan DOCX setelah selesai.

### 16.3 ADMIN Master Pihak Pertama

- page khusus Master Pihak Pertama;
- route dan menu dilindungi `agreement_company_master.view.all`;
- tombol simpan dilindungi `agreement_company_master.manage.all`.

## 17. Template PKS yang Harus Disiapkan

- Ganti seluruh garis bawah, marker `[*]`, dan marker `**` dengan placeholder resmi.
- Ubah paragraf identitas Pihak Kedua agar menggunakan nama, bidang, alamat, pejabat, dan jabatan yang sudah tersedia.
- Jadikan nama/alamat/kontak Pihak Pertama dinamis dari master, bukan teks statis.
- Tambahkan `{{RUANG_LINGKUP}}` di bagian yang saat ini kosong.
- Tambahkan `{{DAFTAR_LAMPIRAN}}` pada Bagian III.
- Pastikan placeholder termin memisahkan persen terbilang dan nilai terbilang.
- Pertahankan layout, page break, header, footer, dan blok tanda tangan.
- Validasi template dengan membuka DOCX hasil generate di Microsoft Word/LibreOffice.

## 18. Storage dan File Access

MinIO digunakan untuk:

- attachment;
- DOCX final;
- PDF final.

MinIO tidak digunakan untuk:

- template sumber;
- template version;
- base PDF template;
- cache koordinat.

Object path dipisahkan, misalnya:

```text
agreement-documents/{agreement-id}/attachments/...
agreement-documents/{agreement-id}/final/agreement.docx
agreement-documents/{agreement-id}/final/agreement.pdf
```

Validasi attachment meliputi MIME allowlist, ukuran per file, jumlah file, total ukuran, sanitasi nama file, dan pemeriksaan bahwa metadata file dimiliki pengajuan yang diminta.

## 19. Docker dan Konfigurasi

- Ubah backend build context agar dapat menyalin folder `.docs/agreement-templates`.
- Salin template/schema ke lokasi runtime read-only.
- Tambahkan konfigurasi root template, misalnya `AGREEMENT_TEMPLATE_ROOT`.
- LibreOffice tetap tersedia untuk konversi headless.
- Aplikasi harus gagal start dengan error yang jelas jika registry, schema, atau template PKS tidak valid.
- Startup validation menghitung checksum, membaca ZIP DOCX, dan memeriksa placeholder wajib.

## 20. Audit Log

Catat minimal:

- create/update/delete/resubmit pengajuan;
- upload/download attachment;
- start review;
- perubahan metadata approver, termasuk nilai sebelum/sesudah;
- return dan reject beserta catatan;
- approve dan identitas approver;
- generate/download PDF;
- generate/download DOCX;
- perubahan Master Pihak Pertama;
- koreksi nomor perjanjian.

Jangan menaruh isi file, token presigned URL, atau data sensitif penuh di audit log.

## 21. Tahapan Implementasi

### Fase 0 — Baseline dan keputusan migrasi

- [ ] Pastikan build/test branch aktif lulus.
- [ ] Catat file untracked yang sudah ada dan jangan menimpa perubahan developer.
- [ ] Gunakan branch lama hanya sebagai referensi selektif.
- [ ] Tentukan apakah database target pernah menjalankan migration Agreement branch lama.
- [ ] Jika pernah, buat migration kompatibilitas terpisah; jangan mengandalkan AutoMigrate untuk cleanup destruktif.

### Fase 1 — Finalisasi template dan schema PKS

- [ ] Buat folder resmi PKS.
- [ ] Pindahkan satu template sumber resmi.
- [ ] Buat `form-schema.json`.
- [ ] Masukkan seluruh placeholder.
- [ ] Perbaiki paragraf Pihak Kedua.
- [ ] Tambahkan daftar lampiran.
- [ ] Buka dan periksa DOCX secara visual.

### Fase 2 — Domain dan database

- [ ] Tambahkan entity Agreement, attachment, dan Company Master.
- [ ] Tambahkan migration dan index.
- [ ] Tambahkan unique constraint nomor perjanjian.
- [ ] Tambahkan seed Master Pihak Pertama.
- [ ] Tambahkan registry tipe dokumen.

### Fase 3 — Permission dan route skeleton

- [ ] Tambahkan permission catalog.
- [ ] Tambahkan baseline USER dan LEGAL.
- [ ] Pastikan ADMIN tetap memperoleh seluruh catalog.
- [ ] Tambahkan route USER/ADMIN/LEGAL dengan middleware.
- [ ] Tambahkan ownership check dan status validation.
- [ ] Tambahkan test akses negatif, terutama DOCX untuk USER.

### Fase 4 — Request workflow

- [ ] Implementasikan create/list/detail/update/delete.
- [ ] Implementasikan multiple attachment.
- [ ] Implementasikan start review/return/reject/resubmit.
- [ ] Implementasikan metadata override per pengajuan.
- [ ] Implementasikan nomor otomatis dan koreksi approver.

### Fase 5 — DOCX generator

- [ ] Implementasikan mapping placeholder PKS.
- [ ] Tangani XML run terpecah.
- [ ] Tangani table/header/footer.
- [ ] Tangani multiline dan XML escaping.
- [ ] Implementasikan format tanggal, Rupiah, persen, dan terbilang.
- [ ] Validasi placeholder tersisa.
- [ ] Tambahkan unit test generator.

### Fase 6 — Preview dan approval

- [ ] Konversi DOCX ke PDF dengan LibreOffice.
- [ ] Tambahkan watermark preview.
- [ ] Implementasikan approval concurrency guard.
- [ ] Simpan snapshot, DOCX final, PDF final, dan checksum.
- [ ] Pastikan final tanpa watermark.
- [ ] Implementasikan cleanup failure dan idempotency.

### Fase 7 — Frontend USER

- [ ] Tambahkan menu dan route berbasis permission.
- [ ] Tambahkan list/detail.
- [ ] Tambahkan pemilih tipe dan dynamic form.
- [ ] Tambahkan multiple attachment.
- [ ] Tambahkan preview PDF dan download PDF final.
- [ ] Pastikan tidak ada client method/tombol DOCX untuk USER.

### Fase 8 — Frontend approver dan master

- [ ] Tambahkan list/review ADMIN dan LEGAL.
- [ ] Tambahkan editor metadata dan preview.
- [ ] Tambahkan decision actions.
- [ ] Tambahkan download PDF/DOCX.
- [ ] Tambahkan page Master Pihak Pertama khusus ADMIN.

### Fase 9 — Regression dan handover

- [ ] Jalankan test backend/frontend.
- [ ] Uji visual DOCX/PDF.
- [ ] Uji permission override ALLOW/DENY.
- [ ] Uji dua approver bersamaan.
- [ ] Dokumentasikan cara menambah tipe dokumen baru.
- [ ] Dokumentasikan cara mengganti template PKS melalui repository/deploy.

## 22. Rencana Pengujian

### 22.1 Unit test

- schema validation per tipe;
- placeholder mapping lengkap;
- placeholder berulang dan lintas XML run;
- karakter XML khusus;
- multiline;
- tanggal Indonesia;
- Rupiah dan terbilang;
- persen dan persen terbilang;
- aturan total termin;
- nomor rekening dengan nol di depan;
- field wajib kosong;
- placeholder tersisa;
- valid/invalid status transition;
- ownership checks.

### 22.2 Integration test

- create PKS dengan beberapa attachment;
- preview watermark;
- return dan resubmit;
- reject final;
- approve menghasilkan DOCX dan PDF dari data yang sama;
- final file tanpa watermark;
- attachment tetap terpisah;
- snapshot tidak berubah setelah master diedit;
- dua approval bersamaan hanya menghasilkan satu final state;
- failure MinIO/converter tidak menghasilkan status `COMPLETED` palsu.

### 22.3 Permission test

- USER hanya melihat data miliknya;
- USER tidak dapat mengakses ID milik USER lain;
- USER dapat preview PDF miliknya;
- USER dapat download PDF final miliknya;
- USER menerima `403` pada endpoint DOCX, termasuk jika mengetahui URL/ID;
- LEGAL dapat review seluruh pengajuan;
- LEGAL dapat download DOCX final;
- LEGAL tidak dapat mengubah master global;
- ADMIN dapat mengubah master global;
- override `DENY` menghilangkan akses sebagaimana effective permission;
- UI tersembunyi dan API tetap menolak ketika permission tidak tersedia.

### 22.4 Visual test

- cover dan judul tidak bergeser;
- nomor kedua pihak tampil benar;
- paragraf identitas kedua pihak terbaca benar;
- page break, header, footer, dan nomor halaman tetap benar;
- tabel tanda tangan sejajar;
- ruang lingkup multiline tidak merusak layout;
- nilai dan terbilang tidak terpotong;
- Bagian III menampilkan daftar attachment;
- tidak ada marker `_`, `[*]`, `**`, atau `{{...}}` tersisa;
- PDF preview memiliki watermark;
- DOCX/PDF final tidak memiliki watermark.

## 23. Acceptance Criteria

- [ ] PKS dapat diajukan USER dan direview ADMIN/LEGAL.
- [ ] Setiap tipe hanya memiliki satu template aktif dari repository.
- [ ] Tidak ada upload/versioning/kalibrasi template.
- [ ] Form PKS berasal dari schema tipe dokumen.
- [ ] Struktur mendukung penambahan tipe dokumen tanpa menambah kolom per field.
- [ ] Request revision tersedia dan template revision tidak tersedia.
- [ ] Nomor Pihak Pertama otomatis, unik, dan dapat dikoreksi approver.
- [ ] Master Pihak Pertama hanya dapat dikelola ADMIN.
- [ ] Override Pihak Pertama per pengajuan dapat dilakukan ADMIN/LEGAL.
- [ ] Data Pihak Pertama final disimpan sebagai snapshot.
- [ ] Attachment dapat lebih dari satu dan tetap terpisah dari dokumen final.
- [ ] Bagian III memuat daftar attachment.
- [ ] Preview menggunakan PDF watermark.
- [ ] Approval menghasilkan DOCX dan PDF final tanpa watermark.
- [ ] USER hanya dapat mengunduh PDF final dan tidak dapat mengakses DOCX.
- [ ] ADMIN/LEGAL dapat mengunduh PDF dan DOCX final.
- [ ] Permission mengikuti catalog, role baseline, override, middleware, dan frontend guard yang sudah ada.
- [ ] Ownership dan status diperiksa di backend.
- [ ] Tidak ada placeholder tersisa pada output.
- [ ] Build dan test backend/frontend lulus.

## 24. Definition of Done

Feature selesai ketika USER dapat mengajukan PKS beserta beberapa attachment, ADMIN/LEGAL dapat mereview dan mengoreksi data, approval menghasilkan DOCX dan PDF final yang konsisten, USER hanya dapat mengakses PDF final, attachment tetap terpisah, serta seluruh akses mengikuti effective permission dan ownership check yang berlaku pada aplikasi.

Arsitektur dianggap siap untuk tipe berikutnya ketika tipe baru dapat ditambahkan melalui satu template, satu schema, satu mapper/validator bila diperlukan, dan test—tanpa membuat modul workflow, tabel pengajuan, atau halaman form baru.
