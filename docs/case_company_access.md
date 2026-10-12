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
| Induk | PT Reasuransi Indonesia Utama |
| Anak usaha | Perusahaan lain dalam grup, mis. PT Asuransi ASEI Indonesia dan PT Reasuransi Syariah Indonesia |
| Legal AU | Peran `LEGAL_AU`: legal dari anak usaha. "AU" berarti anak usaha |
| Perusahaan pemilik perkara | Nilai `company_id` pada perkara |
| Perusahaan internal | Perusahaan dengan `is_internal = true`. Usernya melihat perkara semua perusahaan. Hanya induk yang seharusnya bertanda ini |

Pembatasan ditentukan oleh **perusahaan user**, bukan oleh perannya. User dari
perusahaan non-internal dibatasi ke perkara perusahaannya, apa pun perannya.

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
| Perusahaan baru bisa disimpan sebagai non-internal | Belum. Perusahaan yang dibuat non-internal tersimpan sebagai internal |
| Perusahaan pemilik perkara mengikuti pilihan di form | Belum. Selalu perusahaan pembuatnya |
| Perusahaan perkara bisa diubah oleh user induk | Belum |
| Detail, PDF, ubah, hapus, dokumen, foto, kronologi dibatasi per perusahaan | Belum |
| Legal AU bisa membuat perkara dan melampirkan dokumen serta foto | Belum |
| Anak usaha bertanda non-internal (seed dan database yang sudah ada) | Belum. Semua perusahaan bawaan bertanda internal |
| Dokumen perkara di Repositori Dokumen dibatasi per perusahaan | Belum. Semua peran melihat semua dokumen |

Selama dua baris tentang tanda non-internal belum berlaku, hampir tidak ada user
yang benar-benar dibatasi, karena semua perusahaan bawaan bertanda internal.

## 7. Yang belum diputuskan

- **Akun `EXTERNAL`** tidak terikat ke perusahaan mana pun, sehingga melihat
  perkara semua perusahaan. Belum diputuskan apakah ini harus dibatasi.

## 8. Mengatur perusahaan

Tanda internal diatur ADMIN di menu **Data Master > Perusahaan**. Centang
"Internal" hanya untuk induk. Mencentangnya pada anak usaha membuat semua user
anak usaha itu melihat perkara semua perusahaan.

Rujukan kode: penentuan perusahaan user ada di `getCompanyID()` pada
`backend/internal/handler/legal_case.handler.go`, dan tesnya di
`TestPerkaraPerPerusahaan` pada `backend/internal/router/perkara_test.go`.
