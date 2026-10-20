# Panduan Menyusun Template Perjanjian - ALEA

Panduan untuk Admin dan tim legal yang menyusun atau memperbarui template dokumen perjanjian di Microsoft Word,
lalu mengunggahnya di menu **Data Master → Jenis Perjanjian** (pilih jenisnya, mis. Perjanjian Kerja Sama, lalu
**Atur**).

## 1. Alur singkat

1. Pastikan isian yang dibutuhkan sudah ada di bagian **Isian Form** dan sudah disimpan (**Simpan isian**). Untuk
   jenis baru, susun isiannya lebih dulu.
2. Unduh template yang sedang aktif (ikon unduh di **Riwayat Versi**), lalu sunting di Word. Placeholder bisa disalin
   dari bagian **Placeholder untuk Word**.
3. Di **Versi Template Baru**, pilih file, isi nama versi dan **Catatan Perubahan** (wajib: apa yang berubah dibanding
   versi sebelumnya), lalu **Unggah & Uji**. Kalau Word tidak berubah (mis. hanya label atau status wajib isian yang
   diganti), pakai **Simpan sebagai versi baru**: DOCX versi terakhir dipakai ulang.
4. Sistem menguji file otomatis: struktur `.docx`, setiap placeholder dicocokkan dengan isian jenis dokumen, format
   paragraf ruang lingkup, lalu membentuk dokumen contoh dan mengubahnya ke PDF. Template yang gagal uji tidak
   tersimpan; pesannya menjelaskan sebabnya (bagian 9). Template yang lolos bisa membawa peringatan, mis. isian yang
   tidak dipakai di Word atau deretan Enter kosong; periksa sebelum mengaktifkan.
5. Versi yang lolos tersimpan berstatus **Draft**. Buka **Pratinjau** dan periksa hasilnya dengan daftar periksa di
   bagian 8.
6. Klik **Aktifkan**. Versi aktif sebelumnya otomatis diarsipkan. Dialog konfirmasi menyebut berapa pengajuan berjalan
   yang tetap memakai versi lama.

Kalau unggahan ditolak karena placeholder yang belum punya isian, tombol **Jadikan isian** menambahkannya ke Isian Form
sebagai teks singkat di bagian baru. Periksa jenis isiannya (uang, tanggal, dan sebagainya), simpan isian, lalu unggah
ulang.

Perubahan Isian Form baru dipakai pengaju setelah ada versi baru yang diaktifkan. Selama itu halaman menampilkan
penanda "Ada perubahan isian yang belum dipakai versi mana pun".

Yang perlu diketahui:

- Hanya ada satu versi aktif. Pengajuan **baru** memakai versi aktif.
- Pengajuan yang **sudah dikirim** tetap memakai versi yang dipakainya saat dikirim, sampai selesai. Mengaktifkan
  versi baru tidak mengubah pengajuan yang sedang berjalan.
- Versi lama tersimpan sebagai riwayat dan bisa diunduh. Versi yang tidak dipakai lagi boleh dihapus, kecuali versi
  aktif dan versi yang masih dipakai pengajuan yang belum selesai.
- Versi berketerangan **"Format lama, hanya bisa diunduh"** adalah template sebelum aplikasi memakai placeholder.
  Versi itu tidak bisa dipratinjau atau diaktifkan lagi.
- Yang boleh mengunggah, mengaktifkan, dan menghapus: **Admin**. LEGAL bisa melihat, mengunduh, dan mempratinjau.

## 2. Placeholder

Placeholder adalah penanda yang diganti isian pengajuan saat dokumen dibentuk, ditulis dengan dua kurung kurawal
dan huruf besar:

```
Perjanjian ini dibuat antara PIHAK PERTAMA dan {{PIHAK_KEDUA_NAMA}} tentang {{JENIS_PEKERJAAN}}.
```

Daftar lengkap placeholder beserta artinya ada di halaman jenis perjanjian, bagian **Placeholder untuk Word**. Daftar itu
yang berlaku; jangan menebak nama placeholder. Daftarnya dibentuk dari isian jenis dokumen itu: kode placeholder adalah
nama isian dalam huruf besar, ditambah turunan menurut jenis isiannya (tabel di bawah), dan placeholder sistem yang
sama untuk semua jenis (nomor perjanjian, tanggal dan tempat tanda tangan, Pihak Pertama, daftar lampiran).

Cara menulis:

- **Ketik atau tempel placeholder sekaligus**, lalu jangan ubah formatnya di tengah (mis. menebalkan separuhnya).
  Word memecah teks yang formatnya berubah. Placeholder yang ditulis benar tetap terisi walau terpecah, tetapi
  placeholder yang **salah ketik dan terpecah** tidak terdeteksi saat unggah dan akan tercetak apa adanya di
  dokumen. Kalau ragu, hapus lalu tempel ulang placeholder sebagai teks biasa.
- Format seluruh placeholder (tebal, miring, huruf) ikut menjadi format isiannya.
- Placeholder yang tidak punya isian di jenis dokumen membuat unggahan ditolak dengan pesan *"placeholder tidak
  dikenal: ..."* beserta daftarnya. Buat isiannya di jenis dokumen, atau perbaiki ejaannya di Word.

Isian sudah diformat sistem. Jangan menambahkan format sendiri di sekitar placeholder:

| Jenis | Contoh hasil | Catatan |
|---|---|---|
| Nilai uang | `Rp 750.000.000` | Sudah memuat "Rp"; jangan menulis "Rp" di depan placeholder |
| Persen dan angka | `40`, `33,5` | Tanpa tanda "%"; tulis "%" atau "persen" di template |
| Terbilang | `tujuh ratus lima puluh juta`, `tiga puluh tiga koma lima` | Placeholder berakhiran `_TERBILANG` untuk uang, persen, angka, dan tanggal; kata "rupiah" atau "persen" ditulis di template |
| Tanggal lengkap | `17 Agustus 2026` | Turunannya: `_HARI` (`Senin`) dan `_TERBILANG` (`tujuh belas Agustus dua ribu dua puluh enam`) |
| Isian yang dikosongkan | `-` | Berlaku untuk semua jenis isian, termasuk turunannya |
| Bagian tanggal tanda tangan | `Senin`, `17`, `Agustus`, `2026` | `HARI_TTD`, `TANGGAL_TTD`, `BULAN_TTD`, `TAHUN_TTD` |
| Data Pihak Pertama | Alamat, telepon, email, PIC, pejabat | Diambil dari menu Pihak Pertama, bukan diketik di template |

## 3. Ruang lingkup pekerjaan

- Tulis `{{RUANG_LINGKUP_LIST}}` di **paragraf sendiri**, dan format paragraf itu sebagai **numbered list** Word
  (Home → Numbering). Tanpa format itu, unggahan ditolak.
- Sistem menggandakan paragraf itu sekali per poin isian, sehingga penomorannya (a, b, c) mengikuti list Word.
- Kalau ruang lingkup ingin ditampilkan sebagai satu blok teks biasa, pakai `{{RUANG_LINGKUP}}`.

## 4. Pindah ke halaman baru

Gunakan **pemisah halaman**, jangan menekan Enter berulang kali:

- **Ctrl+Enter** di akhir halaman, atau
- klik kanan judul → *Paragraph* → tab *Line and Page Breaks* → centang **Page break before** (cara yang dipakai
  template bawaan untuk BAGIAN I, II, dan III).

Deretan Enter kosong ikut terdorong ketika isian di atasnya panjang (alamat, ruang lingkup), sehingga muncul ruang
kosong besar atau halaman kosong. Ruang kosong di dalam tabel (mis. tempat tanda tangan basah) tidak masalah.

## 5. Penanda di template

- **Sorotan (highlight)** boleh dipakai sebagai penanda kerja, mis. untuk menandai isian atau rujukan pasal yang
  perlu dicek. Sorotan hanya tampil di **pratinjau**; dokumen final yang disetujui dibersihkan dari sorotan. Karena
  itu sorotan tidak bisa dipakai untuk penekanan di dokumen final.
- **Jangan memasang watermark** (mis. "DRAFT") di template. Sistem menambahkan watermark DRAFT sendiri di setiap
  halaman pratinjau, dan dokumen final harus bersih. Watermark yang terpasang di template akan ikut tercetak di
  dokumen final.
- Header dan footer lain (logo, nomor halaman, kolom paraf) boleh dipakai seperti biasa.

## 6. Rujukan pasal

Kalimat seperti *"sebagaimana dimaksud dalam Pasal 9 ayat (2) huruf (h)"* adalah teks biasa; nomornya tidak berubah
sendiri. Setelah menyisipkan, menghapus, atau memindah pasal, periksa semua rujukan pasal di template. Template
bawaan menyorot rujukan-rujukan itu supaya mudah ditemukan.

## 7. Isian yang bisa kosong

Daftar isian dan mana yang wajib saat ini ditetapkan aplikasi. Isian yang tidak wajib dan dibiarkan kosong tercetak
sebagai tanda `-`, mis. *"Surat Penawaran No. - perihal - pada tanggal -"*. Letakkan placeholder isian yang tidak wajib
di kalimat yang tetap terbaca kalau isiannya kosong, atau minta isian itu dijadikan wajib.

## 8. Daftar periksa sebelum mengaktifkan

Buka **Pratinjau** versi Draft dan periksa. Pratinjau diisi data contoh: isian teks tampil sebagai labelnya dalam
kurung siku (mis. `[Nama Perusahaan]`), sehingga terlihat isian mana mengisi bagian mana.

- [ ] Tidak ada placeholder yang tercetak apa adanya (`{{...}}`).
- [ ] Setiap BAGIAN mulai di halaman baru, tanpa ruang kosong besar atau halaman kosong.
- [ ] Ruang lingkup tampil sebagai poin bernomor.
- [ ] Nilai uang dan terbilang tidak ganda (mis. tidak ada "Rp Rp").
- [ ] Rujukan pasal masih menunjuk pasal yang benar.
- [ ] Data Pihak Pertama (alamat, pejabat) terisi dari menu Pihak Pertama.
- [ ] Catatan Perubahan menjelaskan apa yang berubah dibanding versi sebelumnya.

## 9. Pesan galat saat unggah

| Pesan | Arti | Perbaikan |
|---|---|---|
| *placeholder tidak dikenal: {{...}}* | Ada placeholder yang tidak punya isian di jenis dokumen | Bandingkan ejaannya dengan **Placeholder untuk Word**; tempel ulang sebagai teks biasa, atau buat isiannya (**Jadikan isian**) |
| *Catatan perubahan wajib diisi...* | Catatan Perubahan kosong | Tulis apa yang berubah dibanding versi sebelumnya |
| *paragraf {{RUANG_LINGKUP_LIST}} harus diformat sebagai numbered list Word...* | Paragraf ruang lingkup bukan list bernomor | Pilih paragraf itu, klik Numbering di Word |
| *uji coba pembentukan dokumen gagal: ...* | File tidak bisa diisi, mis. placeholder tersisa | Baca rincian setelah titik dua; periksa placeholder yang disebut |
| *uji coba konversi PDF gagal: ...* | File tidak bisa diubah ke PDF | Buka dan simpan ulang file di Word (format `.docx`, bukan `.doc`), lalu unggah lagi |
| *template ini memakai format lama yang tidak didukung lagi...* | Versi format lama dipratinjau atau diaktifkan | Pakai atau unggah versi yang lebih baru |
