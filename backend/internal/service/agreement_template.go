package service

// agreementTemplate is the verbatim "Perjanjian Kerja Sama" (PKS) document with
// placeholder tokens ({{...}}) that are replaced at render time. Punctuation has
// been normalized to ASCII for gofpdf core fonts.
const agreementTemplate = `[C][B]PERJANJIAN KERJA SAMA[N][/C]
[C]Antara[/C]
[C][B]PT REASURANSI INDONESIA UTAMA (Persero)[N][/C]
[C]Dengan[/C]
[C][B]{{PIHAK_KEDUA_NAMA}}[N][/C]
[C]Tentang[/C]
[C][B]{{JENIS_PEKERJAAN}}[N][/C]
PT REASURANSI INDONESIA UTAMA (PERSERO)

NOMOR PIHAK PERTAMA
{{NOMOR_PP}}
NOMOR PIHAK KEDUA
{{NOMOR_PK}}

Pada hari ini, {{HARI}}, tanggal {{TANGGAL}} bulan {{BULAN}} tahun {{TAHUN}} (dd/mm/yyyy) bertempat di {{TEMPAT}} yang bertandatangan di bawah ini:

PT REASURANSI INDONESIA UTAMA (PERSERO), suatu perseroan terbatas yang didirikan dan dijalankan berdasarkan hukum Negara Republik Indonesia, berkedudukan dan beralamat kantor di Jalan Salemba Raya Nomor 30, Jakarta Pusat, dalam hal ini diwakili oleh {{PP_PEJABAT}}, selaku {{PP_JABATAN}} dalam jabatannya tersebut berhak dan berwenang untuk bertindak untuk dan atas nama PT Reasuransi Indonesia Utama (Persero) selanjutnya disebut PIHAK PERTAMA; dan
{{PIHAK_KEDUA_NAMA}}, {{PK_BIDANG}}, selanjutnya disebut PIHAK KEDUA.
PIHAK PERTAMA dan PIHAK KEDUA secara bersama-sama disebut sebagai PARA PIHAK sedangkan masing-masing diantaranya akan disebut sebagai PIHAK. PARA PIHAK terlebih dahulu menerangkan bahwa:
PIHAK PERTAMA adalah Badan Usaha Milik Negara (BUMN) yang bergerak di bidang usaha reasuransi;
PIHAK KEDUA adalah perusahaan yang bergerak di bidang {{PK_BIDANG}};
untuk kepentingan dan manfaat bersama, PIHAK PERTAMA berencana untuk mengadakan kerja sama mengenai pengadaan {{JENIS_PEKERJAAN}} (Pekerjaan);
untuk merealisasikan Pekerjaan tersebut, PIHAK KEDUA telah mengajukan penawaran kepada PIHAK PERTAMA melalui Surat Penawaran No. {{PENAWARAN_NO}} perihal {{PENAWARAN_PERIHAL}} pada tanggal {{PENAWARAN_TGL}} (Proposal); dan
PIHAK PERTAMA telah menunjuk PIHAK KEDUA sebagai pelaksana Pekerjaan berdasarkan Surat Penunjukan No. {{PENUNJUKAN_NO}} perihal {{PENUNJUKAN_PERIHAL}} pada tanggal {{PENUNJUKAN_TGL}}.

Berdasarkan hal-hal tersebut di atas, PARA PIHAK setuju untuk saling mengikatkan diri satu sama lain dalam Perjanjian Kerja Sama antara PT Reasuransi Indonesia Utama (Persero) dengan {{PIHAK_KEDUA_NAMA}} tentang {{JENIS_PEKERJAAN}} PT Reasuransi Indonesia Utama (Persero), selanjutnya disebut Perjanjian. Perjanjian ini terdiri dari beberapa bagian sebagaimana disebut di bawah ini dan bagian-bagian tersebut merupakan satu-kesatuan yang tidak terpisahkan dari Perjanjian ini:

Bagian I: Syarat-syarat dan Ketentuan
Bagian II: Ketentuan Khusus
Bagian III: Lampiran-Lampiran

Demikian Perjanjian ini dibuat dan ditandatangani pada tempat, hari, tanggal, bulan, dan tahun sebagaimana tersebut pada awal Perjanjian ini. Perjanjian ini berlaku mengikat PARA PIHAK sejak ditandatangani dan dibuat dalam rangkap 2 (dua) masing-masing bermeterai cukup serta mempunyai kekuatan hukum yang sama.

PIHAK KEDUA
{{PIHAK_KEDUA_NAMA}}


[meterai Rp10.000]


Nama: {{PK_PEJABAT}}
Jabatan: {{PK_JABATAN}}
PIHAK PERTAMA
PT REASURANSI INDONESIA UTAMA (Persero)



Nama: {{PP_PEJABAT}}
Jabatan: {{PP_JABATAN}}


BAGIAN I
SYARAT-SYARAT DAN KETENTUAN

PASAL 1
DEFINISI DAN PENAFSIRAN

Kecuali ditentukan lain dalam konteks kalimat, maka istilah-istilah yang diawali dengan huruf besar yang dipergunakan dalam Perjanjian ini mempunyai pengertian sebagai berikut:
Berita Acara Serah Terima Pekerjaan adalah dokumen serah terima atas penyelesaian Pekerjaan oleh PIHAK KEDUA kepada PIHAK PERTAMA sebagaimana dibuktikan dengan penandatanganan oleh PARA PIHAK;
Data Pribadi berarti setiap data tentang orang perseorangan yang teridentifikasi atau dapat diidentifikasi secara tersendiri atau dikombinasi dengan informasi lainnya baik secara langsung maupun tidak langsung melalui sistem elektronik atau nonelektronik;
Data Pribadi Pihak Pertama adalah sebagaimana dimaksud dalam Pasal 9 ayat (2) huruf (h) dari Perjanjian ini;
Dokumen Tagihan adalah sebagaimana dimaksud dalam Pasal 7 ayat (3) Perjanjian ini;
Hak Kekayaan Intelektual berarti setiap dan semua paten, hak cipta, merek dagang, nama dagang, hak database, hak penemuan, hak rahasia dagang, hak desain dan semua bentuk lain dari hak kekayaan intelektual dan hak yang serupa atau mirip lainnya yang memberikan manfaat di mana saja di seluruh dunia, bersama-sama dengan semua hak terkait;
Hasil Laporan Pekerjaan adalah dokumen reporting atas Pekerjaan yang telah dilakukan oleh PIHAK KEDUA dan disetujui oleh PIHAK PERTAMA;
Hari Kalender adalah tujuh hari dalam setiap minggu yang dimulai pada hari Senin dan berakhir pada hari Minggu;
Hari Kerja adalah hari selain hari Sabtu, Minggu, dan hari libur resmi nasional di Indonesia sebagaimana ditetapkan oleh Pemerintah Republik Indonesia;
Informasi Rahasia adalah sebagaimana dimaksud dalam Pasal 9 Perjanjian ini;
Jangka Waktu Perjanjian adalah sebagaimana dimaksud dalam Pasal 5 Perjanjian ini;
Keadaan Memaksa adalah sebagaimana dimaksud dalam Pasal 14 ayat (1) Perjanjian ini;
Nilai Kontrak adalah sebagaimana dimaksud dalam Pasal 6 ayat (1) Perjanjian ini;
Pasal yang Tetap Berlaku berarti Pasal 1, Pasal 9, Pasal 10, Pasal 11, Pasal 12, Pasal 15, Pasal 16, Pasal 17, Pasal 18, Pasal 20 ayat (4), Pasal 20 ayat (5), dan Pasal 21 Perjanjian ini;
Pekerjaan memiliki arti sebagaimana dimaksud dalam huruf c pada bagian pendahuluan di atas dengan ruang lingkup sebagaimana disepakati oleh PARA PIHAK dan diatur dalam Perjanjian ini;
Perwakilan adalah pejabat, anggota, karyawan, afiliasi, agen, penasihat, auditor, bankir, penyedia aktual atau potensial keuangan atau asuransi, agen pemeringkat, dan konsultan, dan/atau kuasa dari suatu PIHAK;
Ruang Lingkup Pekerjaan adalah sebagaimana dimaksud dalam Pasal 2 ayat (1) dari Perjanjian ini;
Tenaga Ahli adalah personel-personel PIHAK KEDUA yang ditugaskan dan memiliki kompetensi serta kualifikasi profesional sesuai keahlian yang dibutuhkan dalam melakukan Pekerjaan; dan
Undang-Undang Perlindungan Data Pribadi berarti Undang-Undang Republik Indonesia Nomor 27 Tahun 2022 tentang Perlindungan Data Pribadi termasuk segala peraturan turunannya.

Interpretasi
Apabila suatu hari atau tanggal yang ditetapkan dalam Perjanjian ini dalam kaitannya dengan pelaksanaan suatu hak atau kewajiban, jatuh pada hari libur (yaitu Sabtu, Minggu atau hari libur resmi nasional di Indonesia), maka pelaksanaan hak atau kewajiban tersebut dianggap jatuh pada Hari Kerja berikutnya.
Rujukan terhadap peraturan perundang-undangan yang berlaku mencakup hukum, undang-undang, keputusan, peraturan, konvensi yang berlaku, perintah, pedoman, kode etik, standar, pemberitahuan, petunjuk, aturan, dan peraturan yang diberlakukan negara, pemerintah, atau badan pemerintah, pemerintah daerah, departemen, atau badan pembuat undang-undang atau badan pengatur dan instrumen sejenis.
Pasal-Pasal dan Lampiran-Lampiran adalah untuk pasal-pasal dan lampiran-lampiran dari Perjanjian ini dan rujukan-rujukan terhadap Perjanjian ini mencakup Lampiran-Lampiran darinya.
Judul-judul Pasal dan Lampiran dimasukkan hanya untuk kemudahan dan tidak akan mempengaruhi penafsiran dari Perjanjian ini.
Bentuk kata tunggal mencakup rujukan terhadap bentuk kata jamak dan sebaliknya.
PIHAK PERTAMA atau PIHAK KEDUA atau pihak lainnya harus ditafsirkan sebagaimana mencakup para pengganti kepemilikan, dan penerima penyerahan yang diizinkan.
Setiap referensi terhadap suatu dokumen dalam Perjanjian ini adalah untuk dokumen tersebut sebagaimana yang diubah, dinovasi, ditambah, diperpanjang, dinyatakan kembali atau diganti dari waktu ke waktu.
Perjanjian adalah untuk Perjanjian sebagaimana yang diubah, dinovasi, ditambah, diperpanjang, dinyatakan kembali atau diganti dari waktu ke waktu.

PASAL 2
RUANG LINGKUP

Ruang lingkup Pekerjaan yang harus dilaksanakan oleh PIHAK KEDUA berdasarkan Perjanjian ini adalah sebagaimana dimaksud dalam Bagian II - Ketentuan Khusus Perjanjian ini (selanjutnya disebut sebagai Ruang Lingkup Pekerjaan).
Penjelasan lebih lanjut mengenai Ruang Lingkup Pekerjaan adalah sebagaimana dimaksud dalam Bagian III - Lampiran-Lampiran Perjanjian ini.

PASAL 3
HAK DAN KEWAJIBAN

Hak dan kewajiban PIHAK PERTAMA adalah sebagai berikut:
PIHAK PERTAMA berhak menerima hasil atau pelaksanaan Pekerjaan dari PIHAK KEDUA sesuai dengan ketentuan dan persyaratan yang disepakati dalam Perjanjian ini;
PIHAK PERTAMA berhak memperoleh penjelasan terkait dengan hasil pelaksanaan Pekerjaan dari PIHAK KEDUA;
PIHAK PERTAMA berhak untuk memberikan tanggapan, penolakan dan meminta klarifikasi atas Pekerjaan yang dilaksanakan oleh PIHAK KEDUA sesuai dengan ketentuan dan persyaratan yang disepakati dalam Perjanjian ini; dan
PIHAK PERTAMA wajib melakukan pembayaran atas Nilai Kontrak yang disepakati sebagaimana diatur dalam Perjanjian ini.

Hak dan kewajiban PIHAK KEDUA adalah sebagai berikut:
PIHAK KEDUA berhak untuk menerima pembayaran atas hasil Pekerjaan sebagaimana diatur dalam Perjanjian ini;
PIHAK KEDUA berhak memperoleh informasi, data dan/atau dokumen dari PIHAK PERTAMA yang benar, akurat dan tidak menyesatkan yang diperlukan dalam melaksanakan Pekerjaan;
PIHAK KEDUA wajib untuk memiliki kemampuan dan pemahaman secara keseluruhan sebagaimana dibutuhkan untuk melaksanakan Pekerjaan berdasarkan Perjanjian ini;
PIHAK KEDUA wajib melaksanakan dan menyelesaikan Pekerjaan secara profesional;
PIHAK KEDUA wajib mematuhi semua ketentuan sebagaimana diatur dalam Perjanjian ini, peraturan perundang-undangan yang berlaku, termasuk mengenai Hak Kekayaan Intelektual, serta praktik industri terbaik;
PIHAK KEDUA wajib melaksanakan Ruang Lingkup Pekerjaan dengan penjelasan lebih lanjut sebagaimana dimaksud dalam Bagian II - Ketentuan Khusus Perjanjian ini;
PIHAK KEDUA wajib merahasiakan seluruh data dan informasi yang diberikan oleh PIHAK PERTAMA, termasuk hasil pelaksanaan Pekerjaan, dan tidak boleh digunakan untuk keperluan apa pun, kecuali telah mendapatkan persetujuan tertulis sebelumnya dari PIHAK PERTAMA;
PIHAK KEDUA wajib berkoordinasi dengan PIHAK PERTAMA dalam melaksanakan Pekerjaan;
PIHAK KEDUA wajib menjaga perilaku dan tata tertib selama bekerja di lokasi dan/atau seluruh wilayah kerja PIHAK PERTAMA;
PIHAK KEDUA wajib menyampaikan Berita Acara Serah Terima Pekerjaan;
PIHAK KEDUA wajib menjaga nama baik PIHAK PERTAMA selama pelaksanaan Pekerjaan berdasarkan Perjanjian ini;
PIHAK KEDUA wajib mematuhi segala ketentuan dan prosedur yang berlaku dan mengikat PIHAK PERTAMA selama melaksanakan Pekerjaan berdasarkan Perjanjian ini; dan
PIHAK KEDUA dilarang untuk membuat perjanjian dengan pihak ketiga, baik secara langsung maupun tidak langsung, yang merugikan atau dapat merugikan pelaksanaan Perjanjian ini.

Hak dan kewajiban PARA PIHAK yang bersifat teknis Pekerjaan akan dituangkan dalam Bagian II - Ketentuan Khusus Perjanjian ini.

PASAL 4
TENAGA AHLI

PIHAK KEDUA dalam menyelenggarakan Pekerjaan ini wajib menggunakan Tenaga Ahli kompeten yang dibuktikan dengan kepemilikan atas seluruh sertifikasi dan izin yang dibutuhkan;
cakap hukum dan memiliki pengalaman dalam melaksanakan pekerjaan yang sejenis;
tidak sedang terlibat sebagai pihak yang berperkara dalam sengketa pengadilan baik perkara perdata maupun pidana yang dapat mengakibatkan terganggunya pelaksanaan Pekerjaan;
tidak mengidap penyakit berbahaya/menular yang dapat mengakibatkan terganggunya pelaksanaan Pekerjaan;
terdaftar sebagai karyawan dari PIHAK KEDUA; dan
Tenaga Ahli yang dipekerjakan harus sesuai dengan identitas, kualifikasi dan pengalaman dalam Proposal.

PIHAK KEDUA wajib mengganti Tenaga Ahli yang digunakan apabila atas dasar pertimbangan PIHAK PERTAMA semata, Tenaga Ahli tersebut tidak sesuai dengan ketentuan.
Dalam melaksanakan Pekerjaan, PIHAK KEDUA wajib memperhatikan ketentuan jam kerja bagi karyawan, Tenaga Ahli dan pihak lain yang bekerja untuk dan atas nama PIHAK KEDUA sesuai dengan peraturan perundang-undangan yang berlaku.
Setiap karyawan termasuk Tenaga Ahli, personel, dan/atau pihak lain yang bekerja untuk dan atas nama PIHAK KEDUA tidak terikat dalam hubungan kerja dengan PIHAK PERTAMA, sehingga PIHAK PERTAMA dibebaskan sepenuhnya dari segala tanggung jawab ketenagakerjaan.

PASAL 5
JANGKA WAKTU

Perjanjian ini berlaku sesuai dengan jangka waktu yang diatur pada Bagian II - Ketentuan Khusus Perjanjian ini (Jangka Waktu Perjanjian).
PARA PIHAK dapat memperpanjang Jangka Waktu Perjanjian berdasarkan kesepakatan PARA PIHAK yang dituangkan secara tertulis dalam suatu adendum sebagaimana dimaksud pada Pasal 19 Perjanjian ini.

PASAL 6
NILAI KONTRAK

Tunduk pada ketentuan lain dalam Perjanjian ini, PARA PIHAK sepakat biaya pelaksanaan Ruang Lingkup Pekerjaan yang akan dibayarkan PIHAK PERTAMA kepada PIHAK KEDUA adalah sebagaimana dimaksud dalam Bagian II - Ketentuan Khusus Perjanjian ini.
Nilai Kontrak sebagaimana tersebut pada ayat (1) Pasal ini yang tertuang dalam Bagian II - Ketentuan Khusus Perjanjian ini adalah tetap dan tidak dapat diubah, kecuali terdapat tambahan Pekerjaan maupun pengurangan Pekerjaan yang wajib disepakati terlebih dahulu oleh PARA PIHAK secara tertulis.

PASAL 7
CARA PEMBAYARAN

Pembayaran atas Nilai Kontrak sebagaimana dimaksud dalam Pasal 6 Perjanjian ini dengan rincian sebagaimana dimaksud dalam Bagian II - Ketentuan Khusus Perjanjian ini.
Pembayaran akan dilakukan PIHAK PERTAMA kepada PIHAK KEDUA melalui rekening bank milik PIHAK KEDUA dengan rincian sebagaimana dimaksud dalam Bagian II - Ketentuan Khusus Perjanjian ini.
PIHAK PERTAMA hanya akan melakukan pembayaran Nilai Kontrak apabila PIHAK KEDUA telah mengajukan tagihan kepada PIHAK PERTAMA disertai dengan kelengkapan dokumen tagihan sebagai berikut:
surat permohonan pembayaran (tagihan);
faktur pajak;
Berita Acara Serah Terima Pekerjaan; dan
dokumen lain yang dapat diminta oleh PIHAK PERTAMA dari waktu ke waktu;
(untuk selanjutnya secara kolektif disebut sebagai Dokumen Tagihan).
Setelah PIHAK PERTAMA menerima Dokumen Tagihan dari PIHAK KEDUA secara lengkap maka PIHAK PERTAMA akan melakukan pembayaran sesuai ketentuan yang diatur dalam Bagian II - Ketentuan Khusus Perjanjian ini.
PIHAK PERTAMA berhak untuk melakukan penundaan pembayaran terhadap tagihan Nilai Kontrak apabila tagihan tersebut tidak disertai Dokumen Tagihan atau terdapat kekurangan pada kelengkapan Dokumen Tagihan.
PIHAK PERTAMA berhak untuk menolak melakukan pembayaran terhadap tagihan Nilai Kontrak apabila terdapat pelanggaran yang dilakukan oleh PIHAK KEDUA terhadap syarat dan ketentuan yang diatur dalam Perjanjian ini yang mengakibatkan tidak terselesaikannya Pekerjaan.

PASAL 8
PERNYATAAN DAN JAMINAN

Masing-masing PIHAK menyatakan dan menjamin kepada PIHAK lainnya bahwa pada tanggal penandatanganan Perjanjian ini:
masing-masing PIHAK merupakan badan hukum yang didirikan secara sah berdasarkan peraturan perundang-undangan yang berlaku di Negara Republik Indonesia;
masing-masing PIHAK adalah pihak-pihak yang berhak, berwenang, memiliki izin-izin yang diperlukan dan/atau mempunyai kemampuan untuk melaksanakan kegiatan usahanya dan/atau kewajibannya berdasarkan Perjanjian ini; dan
perwakilan masing-masing PIHAK yang menandatangani Perjanjian ini adalah pihak yang berwenang bertindak untuk dan atas nama masing-masing PIHAK berdasarkan anggaran dasar perseroan dan telah memperoleh izin-izin dan/atau kuasa yang diperlukan.
PIHAK KEDUA menyatakan dan menjamin kepada PIHAK PERTAMA bahwa:
penandatanganan Perjanjian ini oleh PIHAK KEDUA tidak bertentangan atau melanggar Hak Kekayaan Intelektual pihak mana pun, perjanjian-perjanjian lain maupun peraturan perundang-undangan yang berlaku dan mengikat PIHAK KEDUA;
PIHAK KEDUA tidak terlibat dalam suatu perkara perdata maupun pidana yang sedang berlangsung; dan
PIHAK KEDUA memiliki atau berhak untuk menggunakan semua Hak Kekayaan Intelektual yang diperlukan untuk melaksanakan kewajibannya berdasarkan Perjanjian ini.
Apabila pernyataan-pernyataan pada Pasal ini ternyata di kemudian hari terbukti tidak benar dan/atau menyesatkan serta menimbulkan kerugian bagi PIHAK PERTAMA, maka PIHAK KEDUA bersedia untuk menanggung dan mengganti segala kerugian yang diderita oleh PIHAK PERTAMA.

PASAL 9
KERAHASIAAN

Seluruh materi yang berkaitan dengan Perjanjian ini dan seluruh informasi yang dikandung di dalamnya (secara kolektif disebut Informasi Rahasia), termasuk Data Pribadi Pihak Pertama wajib dijaga kerahasiaannya oleh PIHAK KEDUA.
Informasi Rahasia di atas termasuk namun tidak terbatas pada informasi yang secara umum tidak diketahui publik, setiap Data Pribadi yang disampaikan oleh PIHAK PERTAMA kepada PIHAK KEDUA, dan dokumen apapun yang ditandai dengan rahasia atau hak milik.
PIHAK KEDUA tidak diperbolehkan baik selama Jangka Waktu Perjanjian ini, maupun setelah berakhirnya Perjanjian ini, secara langsung atau tidak langsung menggunakan, mengungkapkan, meneruskan, atau mereproduksi Informasi Rahasia maupun Data Pribadi Pihak Pertama tanpa persetujuan tertulis terlebih dahulu dari PIHAK PERTAMA.
Kewajiban merahasiakan Informasi Rahasia maupun Data Pribadi Pihak Pertama dalam Pasal ini juga berlaku terhadap karyawan dan/atau afiliasi dari PIHAK KEDUA.
Ketentuan tentang kerahasiaan dalam Pasal ini tetap berlaku dan mengikat PARA PIHAK meskipun Perjanjian ini telah berakhir dan/atau dibatalkan.

PASAL 10
WANPRESTASI

PIHAK KEDUA dinyatakan telah melakukan wanprestasi dalam melaksanakan Perjanjian ini dalam hal terpenuhinya salah satu atau lebih kondisi berikut:
PIHAK KEDUA tidak melaksanakan salah satu atau lebih kewajibannya dalam Perjanjian ini;
PIHAK KEDUA melaksanakan Pekerjaan tidak sesuai dengan syarat dan ketentuan yang diatur dalam Perjanjian ini;
PIHAK KEDUA terlambat dalam memenuhi atau melaksanakan Pekerjaan sesuai ketentuan Perjanjian ini;
salah satu pernyataan dan jaminan atau keterangan yang dibuat oleh PIHAK KEDUA ternyata terbukti tidak benar dan/atau menyesatkan;
PIHAK KEDUA secara nyata terbukti melakukan perubahan dalam pelaksanaan Pekerjaan tanpa terlebih dahulu mendapat persetujuan tertulis dari PIHAK PERTAMA; dan/atau
PIHAK KEDUA tidak membayar suatu jumlah penggantian ganti rugi, denda atau sanksi sesuai dengan ketentuan Perjanjian ini.
Atas wanprestasi yang dilaksanakan oleh PIHAK KEDUA, PIHAK PERTAMA berhak memutuskan Perjanjian ini secara sepihak tanpa mengesampingkan hak PIHAK PERTAMA untuk melakukan gugatan dan/atau menempuh proses hukum lainnya.

PASAL 11
PENGAKHIRAN PERJANJIAN

Perjanjian berakhir dalam hal:
Jangka Waktu Perjanjian telah berakhir;
PIHAK KEDUA menjadi atau dinyatakan pailit;
pengakhiran sepihak oleh PIHAK PERTAMA karena adanya wanprestasi dari PIHAK KEDUA;
pengakhiran sebagai akibat pelaksanaan hak yang terdapat di dalam Pasal 20 ayat (4) dari Perjanjian ini; atau
kesepakatan secara tertulis oleh PARA PIHAK untuk mengakhiri Perjanjian ini.
Dalam hal terjadi pengakhiran Perjanjian ini karena alasan apapun, Pasal yang Tetap Berlaku harus tetap berlaku dan PIHAK KEDUA tetap wajib menyelesaikan segala Pekerjaan atau kewajiban-kewajiban lainnya yang belum diselesaikan.
PARA PIHAK sepakat untuk mengesampingkan ketentuan Pasal 1266 Kitab Undang-Undang Hukum Perdata sepanjang pasal-pasal tersebut memerlukan putusan atau penetapan pengadilan atas pengakhiran Perjanjian ini.

PASAL 12
SANKSI DAN DENDA

Atas pengakhiran Perjanjian ini dalam hal terjadinya wanprestasi oleh PIHAK KEDUA, PIHAK KEDUA wajib (i) mengembalikan seluruh pembayaran Nilai Kontrak yang telah dilakukan oleh PIHAK PERTAMA kepada PIHAK PERTAMA dan (ii) membayar denda sebesar 10% (sepuluh persen) dari Nilai Kontrak kepada PIHAK PERTAMA.
Apabila PIHAK KEDUA terlambat menyerahkan hasil Pekerjaan sesuai dengan Bagian II - Ketentuan Khusus Perjanjian ini, maka PIHAK KEDUA dikenakan denda sebesar 0,1% (nol koma satu persen) dari Nilai Kontrak untuk setiap Hari Kalender sejak hari keterlambatan. Denda maksimal yang dapat dikenakan adalah sebesar 5% (lima persen) dari Nilai Kontrak.
PIHAK KEDUA wajib untuk melepaskan PIHAK PERTAMA dari segala tuntutan pihak lain yang disebabkan oleh tindakan-tindakan PIHAK KEDUA.

PASAL 13
PEMBEBASAN SANKSI DAN DENDA

PIHAK KEDUA dapat dibebaskan dari pembayaran sanksi dan denda sebagaimana dimaksud dalam Pasal 12 Perjanjian ini apabila memenuhi salah satu atau lebih kondisi berikut:
apabila PIHAK KEDUA dapat membuktikan secara sah bahwa keterlambatan dimaksud terjadi akibat Keadaan Memaksa;
apabila keterlambatan tersebut disebabkan oleh adanya perintah tertulis dari PIHAK PERTAMA untuk menunda sementara waktu pelaksanaan Pekerjaan; dan
apabila permintaan perpanjangan waktu penyelesaian Pekerjaan dari PIHAK KEDUA telah disetujui secara tertulis oleh PIHAK PERTAMA.

PASAL 14
KEADAAN MEMAKSA (FORCE MAJEURE)

Yang dimaksud dengan keadaan memaksa dalam Perjanjian ini adalah setiap hal yang tidak dapat diperkirakan dan berada di luar kehendak serta kuasa PIHAK yang mengalami keadaan memaksa sehingga menyebabkan PIHAK tersebut tidak dapat melaksanakan Pekerjaan. Keadaan memaksa termasuk namun tidak terbatas pada bencana alam, kebakaran, perang, huru-hara, pemogokan, pemberontakan, kerusuhan dan epidemi atau gangguan industri lainnya yang terjadi sehingga mempengaruhi PIHAK KEDUA dalam penyelesaian Pekerjaan.
Apabila terjadi Keadaan Memaksa, PIHAK yang mengalami Keadaan Memaksa wajib memberitahukan kepada PIHAK lainnya secara tertulis selambat-lambatnya dalam waktu 30 (tiga puluh) Hari Kalender sejak terjadinya Keadaan Memaksa dengan disertai bukti-bukti sah.
Atas pemberitahuan tersebut, PIHAK lainnya dapat menyetujui atau menolak secara tertulis Keadaan Memaksa dalam jangka waktu 3 (tiga) Hari Kalender sejak pemberitahuan diterima. Apabila tidak memberikan jawaban maka PIHAK tersebut dianggap menyetujui adanya Keadaan Memaksa.
Apabila terjadi Keadaan Memaksa, maka PARA PIHAK atau salah satu PIHAK tidak dapat dituntut untuk melaksanakan kewajibannya, dan Perjanjian ini serta jadwal pelaksanaan Pekerjaan dapat ditinjau kembali sesuai dengan kesepakatan PARA PIHAK.

PASAL 15
KEKAYAAN INTELEKTUAL

PIHAK PERTAMA adalah pemilik Hak Kekayaan Intelektual atas semua hasil Pekerjaan yang disampaikan PIHAK KEDUA kepada PIHAK PERTAMA berdasarkan Perjanjian ini. Sedangkan PIHAK KEDUA tetap merupakan pemilik Hak Kekayaan Intelektual atas metodologi, keterampilan, dokumen-dokumen penugasan, konsep, analisa, know-how, alat-alat, kerangka, model, dan perspektif industri yang digunakan PIHAK KEDUA.

PASAL 16
PENYERAHAN HAK DAN KEWAJIBAN

PARA PIHAK tidak dapat menyerahkan sebagian atau seluruh hak dan kewajibannya dalam Perjanjian ini kepada pihak lain, tanpa persetujuan tertulis terlebih dahulu dari PIHAK lainnya.
Dalam hal PIHAK PERTAMA memberikan persetujuan tertulis kepada PIHAK KEDUA untuk menyerahkan kewajibannya kepada subkontraktor, maka PIHAK KEDUA tetap berkedudukan sebagai pemegang kewajiban dan penanggung jawab utama kepada PIHAK PERTAMA.

PASAL 17
PEMBERITAHUAN DAN KOMUNIKASI

Pemberitahuan, permintaan atau komunikasi lainnya antara PARA PIHAK wajib dilakukan secara tertulis dan akan disampaikan langsung atau melalui jasa kurir ekspres yang dialamatkan kepada pihak terkait pada alamat yang tercantum di Bagian II - Ketentuan Khusus atau pada alamat lainnya yang diberitahukan oleh pihak tersebut.
Apabila pemberitahuan dilakukan melalui telepon, maka harus dikonfirmasikan secara tertulis dalam jangka waktu selambat-lambatnya 2 (dua) Hari Kerja.
Dalam hal terjadi perubahan informasi, maka perubahan tersebut harus diberitahukan secara tertulis kepada PIHAK lain dalam jangka waktu selambat-lambatnya 5 (lima) Hari Kerja setelah terjadinya perubahan informasi dimaksud.

PASAL 18
HUKUM YANG BERLAKU DAN PENYELESAIAN PERSELISIHAN

Perjanjian ini tunduk pada hukum yang berlaku di Negara Republik Indonesia.
Apabila terjadi perselisihan antara PARA PIHAK, maka PARA PIHAK sepakat untuk menyelesaikan perselisihan dengan cara musyawarah.
PARA PIHAK sepakat bahwa segala perselisihan yang timbul dari atau berhubungan dengan Perjanjian ini yang tidak bisa diselesaikan secara musyawarah akan diselesaikan melalui Pengadilan Negeri Jakarta Pusat.

PASAL 19
ADENDUM

PARA PIHAK dapat mengajukan perubahan Perjanjian ini terhadap hal-hal yang belum diatur dan/atau diperlukan perubahan atas ketentuan-ketentuan dalam Perjanjian ini.
Apabila salah satu PIHAK bermaksud mengadakan perubahan terhadap ketentuan dalam Perjanjian ini, maka PARA PIHAK wajib menegosiasikan hal-hal atau klausul-klausul yang akan diubah. Hasil negosiasi tersebut dituangkan dalam berita acara dan digunakan sebagai dasar penyusunan adendum perjanjian. Adendum dibuat secara tertulis dan berlaku sebagai bagian yang tidak terpisahkan dari Perjanjian ini jika disetujui oleh PARA PIHAK sepanjang masih dalam Jangka Waktu Perjanjian.

PASAL 20
ANTI PENYUAPAN DAN ANTI KORUPSI

Dalam menjalankan kewajiban dan tanggung jawab berdasarkan Perjanjian ini, masing-masing PIHAK wajib mematuhi seluruh peraturan perundang-undangan yang berlaku.
Masing-masing PIHAK menyatakan bahwa pegawai atau konsultannya tidak akan membayar, menawarkan atau memberikan janji untuk membayar kepada pejabat atau pegawai pemerintah, atau kepada partai politik atau kandidat untuk kantor pemerintah dengan tujuan untuk mempengaruhi tindakan atau keputusan dari kantor pemerintah atau pemerintah.
PIHAK yang melanggar ketentuan sebagaimana dimaksud dalam Pasal ini akan membebaskan PIHAK yang dirugikan dari segala tuntutan atau gugatan atau kerugian yang timbul karena adanya pelanggaran tersebut.

PASAL 21
PEKERJAAN TAMBAH / KURANG

Yang dimaksud dengan pekerjaan tambah / kurang adalah pekerjaan yang diperintahkan PIHAK PERTAMA kepada PIHAK KEDUA untuk dilaksanakan, yang sebelumnya tidak / telah tercantum.
Apabila terdapat pekerjaan tambah / kurang, maka oleh PARA PIHAK akan diperhitungkan penyesuaian harga atas penambahan atau pengurangan Pekerjaan tersebut.
Pembayaran Pekerjaan tambahan akan dilaksanakan setelah disepakatinya Pekerjaan tambahan tersebut oleh PARA PIHAK dengan formalitas yang sama sebagaimana Perjanjian ini.

PASAL 22
EVALUASI, MONITORING, dan AUDIT

PIHAK PERTAMA mempunyai hak untuk melakukan evaluasi, monitoring, dan audit kepada PIHAK KEDUA terhadap pelaksanaan Pekerjaan sesuai dengan Perjanjian ini.
PIHAK KEDUA wajib untuk memberikan data sesuai dengan kebutuhan atas evaluasi, monitoring, dan audit kepada PIHAK PERTAMA terhadap pelaksanaan Pekerjaan berdasarkan Perjanjian ini, termasuk namun tidak terbatas pada permintaan data, informasi, dokumen, dan waktu.
Pelaksanaan waktu evaluasi, monitoring dan audit akan disepakati kembali oleh PARA PIHAK.

PASAL 23
LAIN-LAIN

Surat-surat, dokumen-dokumen maupun lampiran-lampiran yang berhubungan dengan Perjanjian ini merupakan satu kesatuan dan bagian yang tidak terpisahkan dari Perjanjian ini.
Risiko inflasi, kenaikan upah dan harga bahan selama pelaksanaan Pekerjaan menjadi beban dan tanggung jawab PIHAK KEDUA sepenuhnya.
Apabila suatu ketentuan dalam Perjanjian ini dianggap tidak sah, tidak dapat dilaksanakan atau melanggar hukum untuk alasan apa pun, maka Perjanjian ini akan tetap berlaku terlepas dari ketentuan tersebut.

BAGIAN II
KETENTUAN KHUSUS

JENIS PEKERJAAN
Pekerjaan yang akan diserahkan kepada PIHAK KEDUA, yaitu {{JENIS_PEKERJAAN}}.

RUANG LINGKUP PEKERJAAN
Ruang lingkup Pekerjaan yang diserahkan kepada PIHAK KEDUA, yaitu:
{{RUANG_LINGKUP}}

JANGKA WAKTU
Jangka waktu pelaksanaan Perjanjian ini yaitu berlaku sejak tanggal {{JANGKA_MULAI}} sampai dengan tanggal {{JANGKA_SELESAI}}.

HARGA DAN TATA CARA PEMBAYARAN
Tunduk pada ketentuan lain dalam Perjanjian ini, PARA PIHAK sepakat bahwa biaya pelaksanaan Ruang Lingkup Pekerjaan yang akan dibayarkan PIHAK PERTAMA kepada PIHAK KEDUA adalah sebesar {{NILAI_KONTRAK}} ({{NILAI_TERBILANG}}) termasuk pajak (untuk selanjutnya disebut Nilai Kontrak).
Pembayaran dari PIHAK PERTAMA kepada PIHAK KEDUA dibagi menjadi 2 (dua) termin, yaitu:
Termin pertama: Pembayaran sebesar {{TERMIN1_PERSEN}}% ({{TERMIN1_NILAI}} persen) dari Nilai Kontrak atau sebesar {{TERMIN1_NILAI}} ({{TERMIN1_TERBILANG}}) sudah termasuk pajak yang berlaku, yang akan dibayarkan setelah Perjanjian ini ditandatangani oleh PARA PIHAK.
Termin kedua: Pembayaran sebesar {{TERMIN2_PERSEN}}% ({{TERMIN2_NILAI}} persen) dari Nilai Kontrak atau sebesar {{TERMIN2_NILAI}} ({{TERMIN2_TERBILANG}}) sudah termasuk pajak yang berlaku, yang akan dibayarkan setelah penyampaian Hasil Laporan Pekerjaan dan Berita Acara Serah Terima Pekerjaan disetujui oleh PIHAK PERTAMA dan ditandatangani oleh PARA PIHAK.
Pembayaran dilakukan PIHAK PERTAMA kepada PIHAK KEDUA melalui rekening bank milik PIHAK KEDUA dengan rincian sebagai berikut:
Bank: {{BANK}}
Nomor Rekening: {{NO_REKENING}}
Atas Nama: {{ATAS_NAMA}}
PIHAK PERTAMA akan melakukan pembayaran kepada PIHAK KEDUA paling lambat 14 (empat belas) Hari Kerja sejak diterimanya Dokumen Tagihan oleh PIHAK PERTAMA, dan tidak ada koreksi dari Dokumen Tagihan dari PIHAK PERTAMA.

PEMBERITAHUAN DAN KOMUNIKASI
PARA PIHAK menyertakan alamat dan nomor telepon kepada pihak lainnya sebagai sarana pemberitahuan dan komunikasi sebagai berikut:

PIHAK PERTAMA
PT Reasuransi Indonesia Utama (Persero)
Alamat: Jl. Salemba Raya No. 30 Kenari Selatan, Jakarta Pusat
Telepon: 021 3920101
e-mail: info@indonesiare.co.id
PIC:

PIHAK KEDUA
{{PIHAK_KEDUA_NAMA}}
Alamat: {{PK_ALAMAT}}
Telepon: {{PK_TELEPON}}
e-mail: {{PK_EMAIL}}
PIC: {{PK_PIC}}

BAGIAN III
LAMPIRAN-LAMPIRAN
`
