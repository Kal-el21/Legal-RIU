# Akses Perkara per Perusahaan - Legal RIU Portal

Dokumen ini menjelaskan siapa yang boleh melihat dan mengelola sebuah perkara
(Case Management) berdasarkan perusahaannya. Aturannya diputuskan pada
12 Oktober 2026.

> **Status penerapan:** aturan di dokumen ini adalah perilaku yang dituju. Per
> 12 Oktober 2026 kode belum sepenuhnya mematuhinya. Bagian 6 mencatat bagian
> mana yang sudah berlaku; perbarui tabel itu setiap kali satu bagian selesai.

## 1. Aturan

Sebuah perkara terlihat oleh dua pihak:

1. **Induk**, yaitu user PT Reasuransi Indonesia Utama. Induk melihat perkara
   semua perusahaan.
2. **Perusahaan pemilik perkara**, yaitu perusahaan yang dipilih di kolom
   "Perusahaan" pada form perkara.

Aturan ini berlaku siapa pun pembuat perkaranya. Perkara yang dibuat orang induk
untuk sebuah anak usaha terlihat oleh anak usaha itu, dan perkara yang dibuat
anak usaha terlihat oleh induk. Anak usaha tidak melihat perkara anak usaha lain.

## 2. Istilah

| Istilah | Arti |
|---|---|
| Induk | PT Reasuransi Indonesia Utama. Sistem mengenalinya dari domain email `indonesiare.co.id`, yang ditetapkan di kode (`ParentCompanyDomain` di `backend/internal/entity/entity.go`) |
| Anak usaha | Perusahaan lain dalam grup, mis. PT Asuransi ASEI Indonesia dan PT Reasuransi Syariah Indonesia |
| Legal AU | Peran `LEGAL_AU`: legal dari anak usaha. "AU" berarti anak usaha |
| Perusahaan pemilik perkara | Nilai `company_id` pada perkara |
| Perusahaan internal | Perusahaan dengan `is_internal = true`. Usernya melihat perkara semua perusahaan. Hanya induk yang seharusnya bertanda ini |

Pembatasan ditentukan oleh **perusahaan user**, bukan oleh perannya. User dari
perusahaan non-internal dibatasi ke perkara perusahaannya, apa pun perannya.

Satu pengecualian: **ADMIN tidak dibatasi**, di perusahaan mana pun ia
terdaftar. ADMIN mengelola seluruh sistem, dan admin bawaan yang dibuat saat
pemasangan bisa saja terdaftar di anak usaha.

### 2.1 Perusahaan user mengikuti perannya

Karena pembatasan mengikuti perusahaan user, perusahaan dua peran legal tidak
bebas dipilih. Aturan ini diperiksa saat ADMIN membuat atau mengubah user.

| Peran | Perusahaan | Di form user |
|---|---|---|
| `LEGAL` | Selalu induk. Pilihan yang dikirim form diabaikan | Kolom Perusahaan tidak ditampilkan |
| `LEGAL_AU` | Wajib dipilih, dan bukan induk | Induk tidak ada di pilihan |
| `USER`, `ADMIN` | Dipilih ADMIN; kalau kosong, induk | Semua perusahaan |
| `EXTERNAL` | Tanpa perusahaan | Kolom Perusahaan tidak ditampilkan |

Saat peran diganti, perusahaannya ikut ditentukan ulang: `LEGAL_AU` yang
dijadikan `LEGAL` pindah ke induk, dan `LEGAL` yang dijadikan `LEGAL_AU` ditolak
sampai anak usahanya dipilih. "Bukan induk" ditentukan dari identitas
perusahaannya, bukan dari centang "Internal".

User yang sudah ada tidak dipindahkan otomatis. `LEGAL` lama yang terdaftar di
anak usaha baru pindah ke induk setelah ADMIN membuka Edit lalu menyimpannya.
User yang masuk lewat LDAP dibuat sebagai `USER` di perusahaan yang cocok dengan
domain emailnya, atau di induk kalau tidak ada yang cocok.

## 3. Siapa melihat apa

| Perkara milik | Dibuat oleh | Induk | Legal AU ASEI | Legal AU Reasuransi Syariah |
|---|---|---|---|---|
| ASEI | orang induk | Ya | Ya | Tidak |
| ASEI | Legal AU ASEI | Ya | Ya | Tidak |
| Reasuransi Syariah | orang induk | Ya | Tidak | Ya |
| Indonesia Re | orang induk | Ya | Tidak | Tidak |

"Melihat" di sini mencakup semua yang menyangkut perkara itu: daftar, detail,
PDF, ubah, hapus, status, kronologi, dokumen, foto, dan dokumen perkara yang
muncul di Repositori Dokumen. User yang tidak boleh melihat sebuah perkara
mendapat jawaban "tidak ditemukan", sama seperti kalau perkara itu tidak ada.

## 4. Menentukan perusahaan pemilik perkara

| Pembuat atau pengubah | Perusahaan pemilik perkara |
|---|---|
| User induk | Bebas memilih di form, saat membuat maupun mengubah |
| User anak usaha | Selalu perusahaannya sendiri. Pilihan lain di form diabaikan |

User anak usaha tidak bisa memindahkan perkara ke perusahaan lain.

## 5. Perkara yang dibuat sebelum aturan ini berlaku

Sebelum aturan ini diterapkan, pilihan "Perusahaan" di form diabaikan: setiap
perkara disimpan atas perusahaan **pembuatnya**. Akibatnya untuk perkara lama:

| Perkara lama dibuat oleh | Tersimpan atas | Sudah benar? |
|---|---|---|
| Legal AU anak usaha | Anak usaha itu | Ya |
| Orang induk, untuk induk sendiri | Induk | Ya |
| Orang induk, untuk sebuah anak usaha | Induk | **Tidak.** Anak usahanya tidak akan melihat perkara ini |

Baris ketiga tidak bisa dibetulkan otomatis, karena data tidak menyimpan
perusahaan mana yang dulu dipilih di form. Yang perlu dilakukan setelah aturan
ini berlaku:

1. User induk (ADMIN atau LEGAL) membuka daftar perkara.
2. Untuk tiap perkara yang sebenarnya milik anak usaha, buka **Edit**, pilih
   perusahaan yang benar di kolom "Perusahaan", lalu simpan.
3. Setelah itu perkara tersebut terlihat oleh Legal AU anak usaha itu.

Perubahan perusahaan tercatat di audit log sebagai perubahan perkara.

## 6. Status penerapan

| Bagian | Berlaku di kode? |
|---|---|
| Daftar perkara dibatasi per perusahaan | Ya |
| Ubah status dibatasi per perusahaan | Ya |
| Perusahaan baru bisa disimpan sebagai non-internal | Ya, lewat tambah, impor, maupun ubah |
| Perusahaan pemilik perkara mengikuti pilihan di form | Ya |
| Perusahaan perkara bisa diubah oleh user induk | Ya. User anak usaha tidak bisa memindahkan perkaranya |
| Detail, PDF, ubah, hapus, status, dokumen, foto, kronologi (termasuk impor) dibatasi per perusahaan | Ya |
| File perkara yang diambil lewat alamat file (unduh dan lihat) dibatasi per perusahaan | Ya. Untuk semua pengguna, path-nya juga harus tercatat di sebuah perkara; selain itu "tidak ditemukan" |
| Legal AU bisa membuat perkara dan melampirkan dokumen serta foto | Ya di backend. Form di frontend sudah disesuaikan, tetapi belum dicoba di browser |
| Anak usaha bertanda non-internal di database baru | Ya. Dari tiga perusahaan bawaan hanya induk yang internal |
| Anak usaha bertanda non-internal di database yang sudah ada | Tidak otomatis. ADMIN mencabut centangnya lewat menu, lihat bagian 8 |
| ADMIN tidak dibatasi perusahaannya | Ya |
| Induk tidak bisa dijadikan non-internal, diganti domainnya, atau dihapus | Ya di backend. Tampilan menu Perusahaan sudah disesuaikan, tetapi belum dicoba di browser |
| Perusahaan user mengikuti perannya (2.1) | Ya di backend. Form user sudah disesuaikan, tetapi belum dicoba di browser |
| Dokumen perkara di Repositori Dokumen dibatasi per perusahaan | Ya, lihat bagian 9 |

Pembatasan ini baru berlaku nyata untuk sebuah anak usaha setelah centang
"Internal"-nya kosong. Sejak itu user anak usaha tersebut hanya melihat perkara
perusahaannya sendiri.

## 7. Yang belum diputuskan

- **Akun `EXTERNAL`** tidak terikat ke perusahaan mana pun, sehingga melihat
  perkara semua perusahaan. Belum diputuskan apakah ini harus dibatasi.

## 8. Mengatur perusahaan

Tanda internal diatur ADMIN di menu **Data Master > Perusahaan**. Centang
"Internal" hanya untuk induk. Mencentangnya pada anak usaha membuat semua user
anak usaha itu melihat perkara semua perusahaan.

Induk dikunci: baris berlabel "Induk" di menu itu tidak bisa dicabut centang
"Internal"-nya, tidak bisa diganti domain emailnya, dan tidak bisa dihapus.
Namanya tetap bisa diubah. Induk tidak bisa dipindahkan ke perusahaan lain
lewat menu.

Perusahaan baru bawaannya tidak internal.

Perubahan centang berlaku saat itu juga dan tidak berubah sendiri: server yang
dinyalakan ulang tidak mengembalikannya. Tiga perusahaan bawaan (induk, ASEI,
Reasuransi Syariah) hanya dibuat otomatis kalau belum ada, dengan hanya induk
yang internal.

Di server yang sudah berjalan sebelum aturan ini, ASEI dan Reasuransi Syariah
masih bertanda internal. ADMIN perlu mencabut centang keduanya satu kali.
Sebelum mencabut, pastikan tidak ada user LEGAL yang terdaftar di anak usaha
itu, karena ia akan ikut dibatasi.

Rujukan kode: penentuan cakupan user ada di `companyScope()` pada
`backend/internal/handler/legal_case.handler.go`, dan pemeriksaannya di
`findCase()` pada `backend/internal/service/legal_case.service.go`. Tesnya di
`backend/internal/router/perkara_test.go`: `TestAnakUsahaBawaanDibatasi`
(aturan utama), `TestOperasiPerkaraPerPerusahaan` (semua operasi), dan
`TestPerusahaanPemilikPerkara` (perusahaan pemilik). Kuncian induk ada di
`backend/internal/service/company.service.go`, dengan tes
`TestPerusahaanIndukDikunci` di `backend/internal/router/master_test.go`.
Perusahaan user per peran ditentukan `resolveCompanySelection()` di
`backend/internal/service/helpers.go`, dengan tes `TestPerusahaanUserPerPeran`
di `backend/internal/router/sesi_test.go`.

## 9. Repositori Dokumen

Repositori Dokumen mengumpulkan file dari modul lain, dan semua peran punya izin
membukanya (`document_repository.view`). Izin itu hanya membuka halamannya. Yang
tampil di dalamnya mengikuti hak akses di modul asal tiap dokumen, di daftar,
detail, maupun unduhan:

| Dokumen | Terlihat oleh |
|---|---|
| Dokumen, foto, dan lampiran kronologi perkara | Yang punya `case_management.view`, dengan batasan perusahaan yang sama seperti perkaranya (bagian 3) |
| Lampiran dan hasil Legal Opinion | Pengajunya (`legal_opinion.view.own`), atau yang punya `legal_opinion.view.all` |
| Lampiran dan hasil Review Document | Pengajunya (`document_review.view.own`), atau yang punya `document_review.view.all` |
| Lampiran perjanjian | Pengajunya (`agreement_document.download_attachment.own`), atau yang punya `.all` |
| PDF final perjanjian | Pengajunya (`agreement_document.download_pdf.own`), atau yang punya `.all` |
| DOCX final perjanjian | Hanya yang punya `agreement_document.download_docx.all` |

Dokumen yang tidak boleh dilihat dijawab "tidak ditemukan". Dokumen dari sumber
yang tidak dikenali aturan di atas tidak ditampilkan ke siapa pun, jadi modul
baru yang mengirim file ke repositori harus menambahkan aturannya lebih dulu.

Rujukan kode: `DocumentAccess` di
`backend/internal/repository/document_repository.repository.go` (saringannya)
dan `access()` di `backend/internal/handler/document_repository.handler.go`
(penyusunannya dari permission). Tesnya `TestRepositoriDokumenMengikutiSumber`
dan `TestRepositoriDokumen` di `backend/internal/router/pemantauan_test.go`.
