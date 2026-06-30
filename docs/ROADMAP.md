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
- [x] Buat tabel audit_log di database
- [x] Middleware/hook untuk log setiap aksi penting
- [x] Endpoint API untuk get audit log
- [x] UI di admin panel untuk view audit log

**Status**: ✅ Implemented

---

### 4. Submission Reminder Notification

**Description**: Notifikasi pengingat untuk pengajuan Legal Opinion dan Document Review yang belum diproses oleh tim legal.

**Requirements**:
- Hitung durasi sejak pengajuan/submission
- Warning kuning setelah 3 hari tanpa update dari role LEGAL
- Warning merah setelah 14 hari tanpa status COMPLETED
- Tampilkan di dashboard Legal dan Admin
- Indicator di sidebar untuk notifikasi

**Implementation Plan**:
- [ ] Hitung warning level di backend (dashboard service)
- [ ] Endpoint API `/api/v1/dashboard/reminders`
- [ ] WarningBadge component (kuning/merah)
- [ ] Integrasi ke LegalDashboardPage (list reminder)
- [ ] Integrasi ke AdminDashboardPage (stat card)
- [ ] Sidebar indicator di LegalLayout

---

### 5. Contract Status Monitoring

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

### Notes for Submission Reminder
- Hitung durasi: `(NOW() - created_at)` atau `(NOW() - updated_at)`
- YELLOW warning: ≥ 3 hari tanpa update
- RED warning: ≥ 14 hari tanpa status COMPLETED
- Query menggunakan GORM dengan kondisi status NOT IN ('COMPLETED', 'REJECTED')

### Notes for PDF Download & Auto-Complete (CASE_MANAGEMENT.md section 12)
- PDF library: gofpdf/unipdf/unidoc atau HTML to PDF
- Auto-complete: trigger di `AdminUploadResult()` service
- Cek status sebelum upload, jika bukan COMPLETED maka set COMPLETED
- Audit log untuk tracking perubahan status otomatis

---

## Priority

1. **High**: Submission Reminder Notification (penting untuk SLA)
2. **High**: PDF Download & Auto-Complete Status
3. **High**: Chart/Analytics Dashboard
4. **Medium**: Contract Generation
5. **Low**: Contract Status Monitoring