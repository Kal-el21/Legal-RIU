# Roadmap & Task Planning

## Upcoming Features

### 1. Chart/Analytics Dashboard

**Description**: Bar chart atau grafik visualisasi untuk admin/legal mengecek statistik kasus.

**Requirements**:
- Grafik batang untuk visualisasi data
- Filter berdasarkan:
  - Status pengajuan (SUBMITTED, UNDER_REVIEW, COMPLETED, dll)
  - Divisi/Requestor Division
  - Date range
  - Jenis dokumen (Legal Opinion vs Document Review)
- Role access: ADMIN dan LEGAL bisa lihat grafik lengkap

**Implementation Plan**:
- [ ] Tambah endpoint API untuk statistik grafik di backend
- [ ] Buat komponen chart di frontend (menggunakan library chart seperti recharts/apexcharts)
- [ ] Integrasi filter ke dashboard admin/legal
- [ ] Cache data untuk performa

---

### 2. Contract Generation & Email

**Description**: Generate contract/document yang sudah template, hanya perlu ganti field tertentu, lalu kirim ke email.

**Requirements**:
- Template contract yang bisa di-custom field
- Multiple email recipient support
- Field placeholder yang bisa diganti:
  - Nama pihak
  - Tanggal
  - Terms & conditions
  - Value/number
- Auto-generate PDF/document
- Send to specified emails

**Implementation Plan**:
- [ ] Buat template engine untuk contract
- [ ] Tambah field "contract_recipients" di document review/legal opinion
- [ ] Integrasi email service
- [ ] Generate PDF endpoint
- [ ] UI untuk mengisi field dan memilih recipient

---

### 3. Audit Log

**Description**: Log aktivitas untuk memantau semua perubahan kasus.

**Requirements**:
- Track semua aksi:
  - Status change
  - File upload
  - User update
  - Login/logout
- Timestamp dan user yang melakukan aksi
- Viewable di admin panel

**Implementation Plan**:
- [ ] Buat tabel audit_log di database
- [ ] Middleware/hook untuk log setiap aksi penting
- [ ] Endpoint API untuk get audit log
- [ ] UI di admin panel untuk view audit log

---

### 4. Contract Status Monitoring

**Description**: Tracking progres contract dari draft sampai signed.

**Requirements**:
- Status workflow tambahan:
  - DRAFT → IN_REVIEW → APPROVED → SIGNED → ARCHIVED
- Milestone tracking
- Notification/alert untuk delay
- Admin/Legal bisa update milestone

**Implementation Plan**:
- [ ] Tambah status workflow baru
- [ ] Timeline view di detail page
- [ ] Notification system
- [ ] Update status workflow endpoint

---

## Technical Notes

### Notes for Chart Feature
- Pertimbangkan menggunakan Recharts (lightweight) atau Chart.js
- Data structure: array of { label, value, date }

### Notes for Contract Generation
- Template format: HTML template atau DOCX template
- Generate menggunakan library seperti unidoc/unipdf atau puppeteer
- Email integration: SMTP atau sesuai config yang ada

### Notes for Audit Log
- Gunakan middleware Gin untuk capture request
- Simpan di tabel terpisah dengan foreign key ke user/submission
- Index database untuk query cepat

---

## Priority

1. **High**: Audit Log (penting untuk compliance)
2. **Medium**: Chart/Analytics (dashboard enhancement)
3. **Low**: Contract Generation (feature baru)
4. **Low**: Contract Status Monitoring (workflow extension)