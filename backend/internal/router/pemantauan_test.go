package router

// Dashboard, pengingat, pengaturan notifikasi, laporan, audit log, dan
// repositori dokumen.
//
// Database tes dipakai bersama semua tes, jadi angka diperiksa lewat selisih
// sebelum dan sesudah sebuah aksi, bukan lewat nilai mutlak.

import (
	"bytes"
	"net/http"
	"slices"
	"testing"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
)

func ambilAngka(t *testing.T, token, alamat, kolom string) float64 {
	t.Helper()
	w := panggil(t, token, http.MethodGet, alamat, nil)
	harapKode(t, w, http.StatusOK, "mengambil "+alamat)
	var data map[string]any
	isiData(t, w, &data)
	angka, ok := data[kolom].(float64)
	if !ok {
		t.Fatalf("%s tidak memuat angka %q: %s", alamat, kolom, w.Body)
	}
	return angka
}

// Angka dashboard mengikuti data: satu Legal Opinion baru menambah hitungan
// pemiliknya, ADMIN, dan LEGAL, tetapi tidak hitungan USER lain.
func TestDashboard(t *testing.T) {
	const surel = "tes-dashboard@example.com"
	userKhusus(t, surel)
	pemilik := masuk(t, surel)
	admin, legal, lain := masukPeran(t, entity.RoleAdmin), masukPeran(t, entity.RoleLegal), masuk(t, surelUserLain)

	const kolom = "total_legal_opinions"
	sumber := map[string]struct{ token, alamat string }{
		"pemilik": {pemilik, "/api/v1/dashboard/stats"},
		"admin":   {admin, "/api/v1/admin/dashboard/stats"},
		"legal":   {legal, "/api/v1/legal/dashboard/stats"},
		"lain":    {lain, "/api/v1/dashboard/stats"},
	}
	sebelum := map[string]float64{}
	for nama, s := range sumber {
		sebelum[nama] = ambilAngka(t, s.token, s.alamat, kolom)
	}

	semuaModulPengajuan[0].buat(t, pemilik)

	for nama, s := range sumber {
		harap := sebelum[nama] + 1
		if nama == "lain" {
			harap = sebelum[nama]
		}
		if got := ambilAngka(t, s.token, s.alamat, kolom); got != harap {
			t.Errorf("%s untuk %s: dapat %v, harap %v", kolom, nama, got, harap)
		}
	}

	for _, s := range []struct{ token, alamat string }{
		{pemilik, "/api/v1/dashboard/recent"}, {admin, "/api/v1/admin/dashboard/recent"}, {legal, "/api/v1/legal/dashboard/recent"},
	} {
		harapKode(t, panggil(t, s.token, http.MethodGet, s.alamat, nil), http.StatusOK, "aktivitas terbaru "+s.alamat)
	}
}

// barisPengingat adalah satu baris lonceng notifikasi.
type barisPengingat struct {
	ID             string `json:"id"`
	SubmissionType string `json:"submission_type"`
	Title          string `json:"title"`
	Status         string `json:"status"`
	Note           string `json:"note"`
	WarningLevel   string `json:"warning_level"`
	IsRead         bool   `json:"is_read"`
}

type pengingat struct {
	Items       []barisPengingat `json:"items"`
	Total       int              `json:"total"`
	UnreadTotal int              `json:"unread_total"`
}

const alamatPengingat = "/api/v1/notifications"

// ambilPengingat mengambil lonceng milik token. saringan ditempel ke alamat,
// mis. "&warning_only=true".
func ambilPengingat(t *testing.T, token, saringan string) pengingat {
	t.Helper()
	w := panggil(t, token, http.MethodGet, alamatPengingat+"?limit=100"+saringan, nil)
	harapKode(t, w, http.StatusOK, "mengambil pengingat")
	var p pengingat
	isiData(t, w, &p)
	return p
}

// cari mengembalikan baris lonceng untuk satu pengajuan, atau nil kalau tidak ada.
func (p pengingat) cari(id string) *barisPengingat {
	for i := range p.Items {
		if p.Items[i].ID == id {
			return &p.Items[i]
		}
	}
	return nil
}

// Lonceng memberi kabar ke pemilik setiap status pengajuannya diubah pemroses,
// dan ke pemroses saat ada pengajuan masuk atau yang terlalu lama belum selesai.
func TestPengingat(t *testing.T) {
	const surel = "tes-pengingat@example.com"
	userKhusus(t, surel)
	pemilik := masuk(t, surel)
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)

	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			// harap memeriksa lonceng milik token: status kosong berarti pengajuan
			// tidak boleh ada di sana.
			harap := func(token, siapa, id, status string) *barisPengingat {
				t.Helper()
				b := ambilPengingat(t, token, "").cari(id)
				switch {
				case status == "" && b != nil:
					t.Errorf("lonceng %s memuat pengajuan berstatus %s, harap tidak ada", siapa, b.Status)
				case status != "" && b == nil:
					t.Errorf("lonceng %s tidak memuat pengajuan, harap ada dengan status %s", siapa, status)
				case status != "" && b.Status != status:
					t.Errorf("lonceng %s: status %s, harap %s", siapa, b.Status, status)
				}
				return b
			}
			mundurkan := func(id string) {
				t.Helper()
				// Melewati ambang merah bawaan (14 hari). Relatif terhadap hari ini
				// supaya tes tidak kedaluwarsa.
				if err := dbTes.Exec("UPDATE "+m.tabel+" SET status_updated_at = now() - interval '30 days' WHERE id = ?", id).Error; err != nil {
					t.Fatal(err)
				}
			}

			// Baru masuk: kabar untuk pemroses, bukan untuk pemilik yang mengajukannya.
			p := m.buat(t, pemilik)
			harap(pemilik, "pemilik", p.ID, "")
			harap(legal, "LEGAL", p.ID, "SUBMITTED")
			harap(admin, "ADMIN", p.ID, "SUBMITTED")

			// Mulai direview: kabar untuk pemilik, dan hilang dari lonceng pemroses.
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			harap(pemilik, "pemilik", p.ID, "UNDER_REVIEW")
			harap(legal, "LEGAL", p.ID, "")
			harap(admin, "ADMIN", p.ID, "")

			// Tandai dibaca: hanya baris itu, dan hanya untuk user yang menandai.
			tandaiDibaca := func(token string, b *barisPengingat) {
				t.Helper()
				w := panggil(t, token, http.MethodPatch, alamatPengingat+"/read",
					map[string]string{"submission_type": b.SubmissionType, "submission_id": b.ID})
				harapKode(t, w, http.StatusOK, "menandai pengingat dibaca")
			}
			sebelum := ambilPengingat(t, pemilik, "")
			tandaiDibaca(pemilik, sebelum.cari(p.ID))
			sesudah := ambilPengingat(t, pemilik, "")
			if b := sesudah.cari(p.ID); b == nil || !b.IsRead || sesudah.UnreadTotal != sebelum.UnreadTotal-1 {
				t.Errorf("setelah ditandai dibaca: baris %+v, belum dibaca %d (sebelumnya %d)", b, sesudah.UnreadTotal, sebelum.UnreadTotal)
			}
			if ambilPengingat(t, pemilik, "&unread_only=true").cari(p.ID) != nil {
				t.Error("saringan belum dibaca memuat baris yang sudah dibaca")
			}

			// Terlalu lama direview: muncul lagi di pemroses sebagai peringatan
			// merah. Peringatan bukan urusan pemilik.
			mundurkan(p.ID)
			if b := ambilPengingat(t, legal, "&warning_only=true").cari(p.ID); b == nil || b.WarningLevel != "RED" {
				t.Errorf("pengajuan 30 hari direview di lonceng LEGAL: %+v, harap peringatan RED", b)
			}
			if b := harap(pemilik, "pemilik", p.ID, "UNDER_REVIEW"); b != nil && b.WarningLevel != "NONE" {
				t.Errorf("pemilik mendapat peringatan %s atas pengajuannya sendiri", b.WarningLevel)
			}
			if ambilPengingat(t, pemilik, "&warning_only=true").cari(p.ID) != nil {
				t.Error("saringan peringatan milik pemilik memuat pengajuannya sendiri")
			}

			// Peringatan yang dibaca selagi kuning kembali belum dibaca setelah naik
			// ke merah. Pengajuan ini masuk merah 16 hari lalu, jadi tanda terbacanya
			// dimundurkan ke 20 hari lalu, saat levelnya masih kuning.
			tandaiDibaca(legal, ambilPengingat(t, legal, "&warning_only=true").cari(p.ID))
			if b := ambilPengingat(t, legal, "&warning_only=true").cari(p.ID); b == nil || !b.IsRead {
				t.Errorf("peringatan setelah ditandai dibaca: %+v", b)
			}
			if err := dbTes.Exec(`UPDATE notification_reads SET read_at = now() - interval '20 days'
				WHERE submission_id = ? AND user_id = (SELECT id FROM users WHERE email = ?)`, p.ID, surelPeran(entity.RoleLegal)).Error; err != nil {
				t.Fatal(err)
			}
			if b := ambilPengingat(t, legal, "&warning_only=true").cari(p.ID); b == nil || b.IsRead {
				t.Errorf("peringatan yang dibaca sebelum naik ke merah masih dianggap terbaca: %+v", b)
			}

			// Minta revisi: catatan sampai ke lonceng pemilik, pemroses tidak lagi menunggu.
			m.ubahStatus(t, legal, "/legal", p.ID, "NEED_REVISION", "lengkapi data", http.StatusOK)
			if b := harap(pemilik, "pemilik", p.ID, "NEED_REVISION"); b != nil && (b.Note != "lengkapi data" || b.IsRead) {
				t.Errorf("kabar revisi di lonceng pemilik: catatan %q, terbaca %v; harap belum dibaca karena statusnya baru berganti", b.Note, b.IsRead)
			}
			harap(legal, "LEGAL", p.ID, "")

			// Ajukan ulang: tindakan pemilik sendiri, jadi kabarnya untuk pemroses.
			harapKode(t, panggil(t, pemilik, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil), http.StatusOK, "ajukan ulang")
			harap(pemilik, "pemilik", p.ID, "")
			harap(legal, "LEGAL", p.ID, "RESUBMITTED")

			// Selesai: kabar untuk pemilik, tanpa catatan revisi lama.
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			m.ubahStatus(t, legal, "/legal", p.ID, "COMPLETED", "", http.StatusOK)
			if b := harap(pemilik, "pemilik", p.ID, "COMPLETED"); b != nil && b.Note != "" {
				t.Errorf("catatan revisi lama ikut terkirim setelah selesai: %q", b.Note)
			}
			harap(legal, "LEGAL", p.ID, "")

			// Pengajuan selesai tidak pernah menjadi peringatan, selama apa pun
			// umurnya, dan kabar selesai yang lebih tua dari 30 hari hilang dari
			// lonceng pemiliknya.
			mundurkan(p.ID)
			if ambilPengingat(t, legal, "&warning_only=true").cari(p.ID) != nil {
				t.Error("pengajuan yang sudah selesai masih menjadi peringatan di lonceng LEGAL")
			}
			harap(pemilik, "pemilik", p.ID, "")
		})
	}

	harapKode(t, panggil(t, pemilik, http.MethodPatch, alamatPengingat+"/read", map[string]string{}), http.StatusBadRequest, "tandai dibaca tanpa isi")

	// Tandai semua dibaca.
	baru := semuaModulPengajuan[0].buat(t, pemilik)
	semuaModulPengajuan[0].ubahStatus(t, legal, "/legal", baru.ID, "UNDER_REVIEW", "", http.StatusOK)
	if n := ambilPengingat(t, pemilik, "").UnreadTotal; n == 0 {
		t.Fatal("kabar baru tidak terhitung belum dibaca")
	}
	harapKode(t, panggil(t, pemilik, http.MethodPatch, alamatPengingat+"/read-all", nil), http.StatusOK, "menandai semua dibaca")
	if n := ambilPengingat(t, pemilik, "").UnreadTotal; n != 0 {
		t.Errorf("belum dibaca setelah tandai semua: %d, harap 0", n)
	}
}

// Dokumen Perjanjian masuk lonceng dengan aturan yang sama, tetapi tanpa
// peringatan karena tidak punya ambang hari. Persetujuan akhir tidak diuji di
// sini: butuh LibreOffice.
func TestPengingatPerjanjian(t *testing.T) {
	butuhMinIO(t)
	const surel = "tes-pengingat-perjanjian@example.com"
	userKhusus(t, surel)
	pemilik := masuk(t, surel)
	legal := masukPeran(t, entity.RoleLegal)

	// harap memeriksa lonceng milik token: status kosong berarti perjanjian tidak
	// boleh ada di sana.
	harap := func(token, siapa, id, status string) *barisPengingat {
		t.Helper()
		b := ambilPengingat(t, token, "").cari(id)
		switch {
		case status == "" && b != nil:
			t.Errorf("lonceng %s memuat perjanjian berstatus %s, harap tidak ada", siapa, b.Status)
		case status != "" && b == nil:
			t.Errorf("lonceng %s tidak memuat perjanjian, harap ada dengan status %s", siapa, status)
		case status != "" && (b.Status != status || b.SubmissionType != "agreement_document"):
			t.Errorf("lonceng %s: %s berstatus %s, harap agreement_document berstatus %s", siapa, b.SubmissionType, b.Status, status)
		}
		return b
	}

	p := buatPerjanjian(t, pemilik, formPerjanjian())
	harap(pemilik, "pemilik", p.ID, "")
	if b := harap(legal, "LEGAL", p.ID, "SUBMITTED"); b != nil && b.Title != "PKS - PT Mitra Tes" {
		t.Errorf("judul baris perjanjian: %q", b.Title)
	}

	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	harap(pemilik, "pemilik", p.ID, "UNDER_REVIEW")
	harap(legal, "LEGAL", p.ID, "")

	// Selama apa pun direview, perjanjian tidak menjadi peringatan.
	if err := dbTes.Exec("UPDATE agreement_documents SET status_updated_at = now() - interval '30 days' WHERE id = ?", p.ID).Error; err != nil {
		t.Fatal(err)
	}
	if ambilPengingat(t, legal, "&warning_only=true").cari(p.ID) != nil {
		t.Error("perjanjian yang lama direview menjadi peringatan, padahal tidak punya ambang")
	}

	statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "lengkapi alamat", http.StatusOK)
	b := harap(pemilik, "pemilik", p.ID, "NEED_REVISION")
	if b == nil || b.Note != "lengkapi alamat" {
		t.Fatalf("kabar revisi perjanjian di lonceng pemilik: %+v", b)
	}
	harap(legal, "LEGAL", p.ID, "")

	// Tandai dibaca menerima jenis agreement_document.
	w := panggil(t, pemilik, http.MethodPatch, alamatPengingat+"/read",
		map[string]string{"submission_type": b.SubmissionType, "submission_id": b.ID})
	harapKode(t, w, http.StatusOK, "menandai kabar perjanjian dibaca")
	if b := ambilPengingat(t, pemilik, "").cari(p.ID); b == nil || !b.IsRead {
		t.Errorf("kabar perjanjian setelah ditandai dibaca: %+v", b)
	}

	harapKode(t, panggil(t, pemilik, http.MethodPost, urlPerjanjian("", "/", p.ID, "/resubmit"), nil), http.StatusOK, "ajukan ulang perjanjian")
	harap(pemilik, "pemilik", p.ID, "")
	harap(legal, "LEGAL", p.ID, "RESUBMITTED")

	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	statusPerjanjian(t, legal, p.ID, "REJECTED", "tidak sesuai kebijakan", http.StatusOK)
	if b := harap(pemilik, "pemilik", p.ID, "REJECTED"); b != nil && (b.Note != "tidak sesuai kebijakan" || b.IsRead) {
		t.Errorf("kabar ditolak di lonceng pemilik: catatan %q, terbaca %v", b.Note, b.IsRead)
	}
	harap(legal, "LEGAL", p.ID, "")
}

// Lonceng dibuka oleh permission lihat pengajuan, bukan oleh peran: EXTERNAL dan
// LEGAL_AU bawaannya ditolak, dan diterima setelah diberi override.
func TestPengingatMengikutiPermission(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = alamatPengingat

	for _, peran := range []entity.UserRole{entity.RoleExternal, entity.RoleLegalAU} {
		token := masukPeran(t, peran)
		harapKode(t, panggil(t, token, http.MethodGet, alamat, nil), http.StatusForbidden, "pengingat "+string(peran)+" bawaan")

		var idUser string
		if err := dbTes.Raw("SELECT id FROM users WHERE email = ?", surelPeran(peran)).Scan(&idUser).Error; err != nil {
			t.Fatal(err)
		}
		atur := func(overrides []map[string]string) {
			t.Helper()
			w := panggil(t, admin, http.MethodPut, "/api/v1/admin/users/"+idUser+"/permissions", map[string]any{"overrides": overrides})
			harapKode(t, w, http.StatusOK, "mengatur override "+string(peran))
		}
		t.Cleanup(func() { atur([]map[string]string{}) })

		atur([]map[string]string{{"code": "legal_opinion.view.own", "effect": "ALLOW"}})
		harapKode(t, panggil(t, token, http.MethodGet, alamat, nil), http.StatusOK, "pengingat "+string(peran)+" dengan override")
	}
}

// Ambang hari pengingat diatur ADMIN.
func TestPengaturanNotifikasi(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/admin/notification-settings"

	w := panggil(t, admin, http.MethodGet, alamat, nil)
	harapKode(t, w, http.StatusOK, "daftar pengaturan notifikasi")
	var daftar []struct {
		ID            string `json:"id"`
		DaysThreshold int    `json:"days_threshold"`
	}
	isiData(t, w, &daftar)
	if len(daftar) != 4 {
		t.Fatalf("pengaturan notifikasi hasil seed: %d baris, harap 4", len(daftar))
	}
	awal := daftar[0]
	// Dikembalikan setelah tes, karena tes pengingat bergantung pada ambang bawaan.
	t.Cleanup(func() {
		panggil(t, admin, http.MethodPut, alamat+"/"+awal.ID, map[string]int{"days_threshold": awal.DaysThreshold})
	})

	harapKode(t, panggil(t, admin, http.MethodPut, alamat+"/"+awal.ID, map[string]int{"days_threshold": awal.DaysThreshold + 5}), http.StatusOK, "mengubah ambang")
	isiData(t, panggil(t, admin, http.MethodGet, alamat, nil), &daftar)
	for _, d := range daftar {
		if d.ID == awal.ID && d.DaysThreshold != awal.DaysThreshold+5 {
			t.Errorf("ambang setelah diubah: %d", d.DaysThreshold)
		}
	}
	harapKode(t, panggil(t, admin, http.MethodPut, alamat+"/"+awal.ID, map[string]int{"days_threshold": 0}), http.StatusBadRequest, "ambang nol")
}

type grafik struct {
	Labels []string `json:"labels"`
	Series []struct {
		Data []float64 `json:"data"`
	} `json:"series"`
}

func (g grafik) jumlah() (total float64) {
	for _, s := range g.Series {
		for _, v := range s.Data {
			total += v
		}
	}
	return total
}

// Tiga laporan: isinya mengikuti data, dan saringan yang salah ditolak.
func TestLaporan(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)
	user := masukPeran(t, entity.RoleUser)

	ambil := func(token, alamat string) grafik {
		t.Helper()
		w := panggil(t, token, http.MethodGet, alamat, nil)
		harapKode(t, w, http.StatusOK, "laporan "+alamat)
		var g grafik
		isiData(t, w, &g)
		return g
	}

	const lo = "/api/v1/legal/reports/legal-opinions?group_by=status"
	sebelum := ambil(legal, lo).jumlah()
	semuaModulPengajuan[0].buat(t, user)
	if got := ambil(legal, lo).jumlah(); got != sebelum+1 {
		t.Errorf("jumlah di laporan Legal Opinion setelah satu pengajuan baru: dapat %v, harap %v", got, sebelum+1)
	}

	// Setiap laporan dengan setiap pengelompokan yang didukungnya.
	kelompok := map[string][]string{
		"legal-opinions":   {"company", "legal_type", "status", "division"},
		"document-reviews": {"company", "document_type", "status", "division"},
		"legal-cases":      {"company", "status", "level", "location", "pic"},
	}
	for laporan, daftar := range kelompok {
		for _, k := range daftar {
			ambil(admin, "/api/v1/admin/reports/"+laporan+"?group_by="+k)
		}
	}
	t.Run("PerkaraPerJenisDanKategori", func(t *testing.T) {
		bugDiketahui(t, "laporan perkara dengan group_by=case_type atau category gagal dengan error SQL (kolom label milik tabel rujukan bentrok dengan alias label di GROUP BY)")
		for _, k := range []string{"case_type", "category"} {
			ambil(admin, "/api/v1/admin/reports/legal-cases?group_by="+k)
		}
	})

	// Saringan tanggal membatasi isi laporan.
	if got := ambil(legal, lo+"&date_from=2000-01-01&date_to=2000-12-31").jumlah(); got != 0 {
		t.Errorf("laporan untuk tahun 2000 seharusnya kosong, dapat %v", got)
	}
	for nama, sisa := range map[string]string{
		"tanpa group_by":         "",
		"group_by tidak dikenal": "?group_by=zodiak",
		"tanggal tidak terbaca":  "?group_by=status&date_from=kemarin",
	} {
		harapKode(t, panggil(t, legal, http.MethodGet, "/api/v1/legal/reports/legal-opinions"+sisa, nil), http.StatusBadRequest, nama)
	}
}

type catatanAudit struct {
	Action     string  `json:"action"`
	EntityType string  `json:"entity_type"`
	EntityID   string  `json:"entity_id"`
	NewValue   *string `json:"new_value"`
}

func cariAudit(t *testing.T, token, saringan string, cocok func(catatanAudit) bool) bool {
	t.Helper()
	w := panggil(t, token, http.MethodGet, "/api/v1/admin/audit-logs?limit=100&"+saringan, nil)
	harapKode(t, w, http.StatusOK, "mengambil audit log")
	var hasil struct {
		Items []catatanAudit `json:"items"`
	}
	isiData(t, w, &hasil)
	for _, c := range hasil.Items {
		if cocok(c) {
			return true
		}
	}
	return false
}

// Aksi sensitif masuk audit log: login, perubahan status, perubahan permission.
func TestAuditLog(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	legal := masukPeran(t, entity.RoleLegal)
	const surel = "tes-audit@example.com"
	userKhusus(t, surel)
	pemilik := masuk(t, surel)
	var idUser string
	dbTes.Raw("SELECT id FROM users WHERE email = ?", surel).Scan(&idUser)

	if !cariAudit(t, admin, "action=LOGIN", func(c catatanAudit) bool { return c.EntityID == idUser }) {
		t.Error("login tidak tercatat di audit log")
	}

	m := semuaModulPengajuan[0]
	p := m.buat(t, pemilik)
	m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
	if !cariAudit(t, admin, "action=STATUS_CHANGE&entity_type=legal_opinion", func(c catatanAudit) bool {
		return c.EntityID == p.ID && c.NewValue != nil && *c.NewValue == "UNDER_REVIEW"
	}) {
		t.Error("perubahan status Legal Opinion tidak tercatat di audit log")
	}

	w := panggil(t, admin, http.MethodPut, "/api/v1/admin/users/"+idUser+"/permissions",
		map[string]any{"overrides": []map[string]string{{"code": "master_data.view", "effect": "ALLOW"}}})
	harapKode(t, w, http.StatusOK, "mengatur override permission")
	if !cariAudit(t, admin, "action=PERMISSION_UPDATE", func(c catatanAudit) bool { return c.EntityID == idUser }) {
		t.Error("perubahan permission tidak tercatat di audit log")
	}

	// LEGAL membaca audit log lewat alamatnya sendiri; USER tidak punya alamat itu.
	harapKode(t, panggil(t, legal, http.MethodGet, "/api/v1/legal/audit-logs", nil), http.StatusOK, "LEGAL membaca audit log")
}

// Repositori dokumen mengumpulkan file dari modul lain.
func TestRepositoriDokumen(t *testing.T) {
	butuhMinIO(t)
	legal := masukPeran(t, entity.RoleLegal)
	const alamat = "/api/v1/document-repositories"
	isi := []byte("dokumen perkara untuk repositori")

	p := buatPerkara(t, legal)
	w := panggilForm(t, legal, http.MethodPost, perkaraLegal+"/"+p.ID+"/upload-document", nil, berkas{"document", "putusan.pdf", isi})
	harapKode(t, w, http.StatusOK, "unggah dokumen perkara")

	cari := func(token string) (id string) {
		t.Helper()
		w := panggil(t, token, http.MethodGet, alamat, nil)
		harapKode(t, w, http.StatusOK, "daftar repositori")
		var daftar []struct {
			ID       string `json:"id"`
			SourceID string `json:"source_id"`
		}
		isiData(t, w, &daftar)
		for _, d := range daftar {
			if d.SourceID == p.ID {
				return d.ID
			}
		}
		return ""
	}

	id := cari(legal)
	if id == "" {
		t.Fatal("dokumen perkara yang baru diunggah tidak muncul di repositori")
	}
	harapKode(t, panggil(t, legal, http.MethodGet, alamat+"/"+id, nil), http.StatusOK, "detail dokumen repositori")
	w = panggil(t, legal, http.MethodGet, alamat+"/"+id+"/download", nil)
	harapKode(t, w, http.StatusOK, "unduh dari repositori")
	if !bytes.Equal(w.Body.Bytes(), isi) {
		t.Errorf("isi yang diunduh dari repositori berbeda: %q", w.Body)
	}

	// Menghapus dokumen perkara ikut mencabutnya dari repositori.
	harapKode(t, panggil(t, legal, http.MethodDelete, perkaraLegal+"/"+p.ID+"/document", nil), http.StatusOK, "hapus dokumen perkara")
	if cari(legal) != "" {
		t.Error("dokumen perkara yang sudah dihapus masih ada di repositori")
	}

	// USER tidak punya akses ke Case Management, jadi dokumen perkara tidak
	// boleh sampai kepadanya lewat repositori: tidak di daftar, tidak lewat ID.
	t.Run("USERTidakMelihatDokumenPerkara", func(t *testing.T) {
		w := panggilForm(t, legal, http.MethodPost, perkaraLegal+"/"+p.ID+"/upload-document", nil, berkas{"document", "rahasia.pdf", isi})
		harapKode(t, w, http.StatusOK, "unggah dokumen perkara")
		id := cari(legal)
		if id == "" {
			t.Fatal("dokumen perkara yang baru diunggah tidak muncul di repositori")
		}
		user := masukPeran(t, entity.RoleUser)
		if cari(user) != "" {
			t.Error("USER melihat dokumen perkara di daftar repositori")
		}
		harapKode(t, panggil(t, user, http.MethodGet, alamat+"/"+id, nil), http.StatusNotFound, "USER membuka detail dokumen perkara")
		harapKode(t, panggil(t, user, http.MethodGet, alamat+"/"+id+"/download", nil), http.StatusNotFound, "USER mengunduh dokumen perkara")
	})
}

// Repositori dokumen mengikuti hak akses sumbernya: sebuah dokumen hanya
// terlihat oleh yang boleh membukanya di modul asalnya. Baris repositori dibuat
// langsung di tabelnya supaya tes ini tidak butuh MinIO; unduhan sungguhan
// diuji TestRepositoriDokumen.
func TestRepositoriDokumenMengikutiSumber(t *testing.T) {
	const alamat = "/api/v1/document-repositories"
	const pdf = "application/pdf"
	const docx = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

	idPerusahaan := func(domain string) string {
		t.Helper()
		var id string
		if err := dbTes.Raw("SELECT id FROM companies WHERE email_domain = ? AND deleted_at IS NULL", domain).Scan(&id).Error; err != nil || id == "" {
			t.Fatalf("perusahaan %s tidak ada: %v", domain, err)
		}
		return id
	}
	asei, syariah := idPerusahaan("asei.co.id"), idPerusahaan("resyariah.co.id")
	pengguna := func(surel string, peran entity.UserRole, perusahaan string) string {
		t.Helper()
		if err := tambahUser(surel, peran, uuid.MustParse(perusahaan)); err != nil {
			t.Fatal(err)
		}
		return masuk(t, surel)
	}

	const (
		orangLegal    = "LEGAL"
		orangAdmin    = "ADMIN"
		orangPemilik  = "USER pemilik pengajuan"
		orangUserLain = "USER lain"
		orangAUAsei   = "Legal AU ASEI"
		orangAUSya    = "Legal AU Reasuransi Syariah"
		orangExternal = "EXTERNAL"
	)
	legal, pemilik := masukPeran(t, entity.RoleLegal), masukPeran(t, entity.RoleUser)
	token := map[string]string{
		orangLegal:    legal,
		orangAdmin:    masukPeran(t, entity.RoleAdmin),
		orangPemilik:  pemilik,
		orangUserLain: pengguna("tes-repo-user-lain@example.com", entity.RoleUser, idPerusahaan(entity.ParentCompanyDomain)),
		orangAUAsei:   pengguna("tes-repo-au-asei@example.com", entity.RoleLegalAU, asei),
		orangAUSya:    pengguna("tes-repo-au-syariah@example.com", entity.RoleLegalAU, syariah),
		orangExternal: masukPeran(t, entity.RoleExternal),
	}

	// Sumber dokumennya: dua perkara anak usaha, dan tiga pengajuan milik satu USER.
	perkaraUntuk := func(perusahaan string) perkara {
		t.Helper()
		form := formPerkara(t)
		form["company_id"] = perusahaan
		w := panggil(t, legal, http.MethodPost, perkaraLegal, form)
		harapKode(t, w, http.StatusCreated, "membuat perkara untuk anak usaha")
		var p perkara
		isiData(t, w, &p)
		return p
	}
	perkaraAsei, perkaraSyariah := perkaraUntuk(asei), perkaraUntuk(syariah)
	harapKode(t, panggil(t, legal, http.MethodPost, perkaraLegal+"/"+perkaraAsei.ID+"/chronology", map[string]string{"agenda_date": "2026-10-05", "agenda": "Sidang"}), http.StatusCreated, "menambah kronologi")
	var kronologi string
	dbTes.Raw("SELECT id FROM case_chronologies WHERE case_id = ?", perkaraAsei.ID).Scan(&kronologi)
	if kronologi == "" {
		t.Fatal("kronologi yang baru dibuat tidak ditemukan")
	}
	opini := semuaModulPengajuan[0].buat(t, pemilik)
	review := semuaModulPengajuan[1].buat(t, pemilik)
	// Perjanjian dibuat langsung di database: lewat API butuh template aktif,
	// dan template disimpan di MinIO.
	var idPemilik string
	dbTes.Raw("SELECT id FROM users WHERE email = ?", surelPeran(entity.RoleUser)).Scan(&idPemilik)
	pks := entity.AgreementDocument{TicketNumber: "UJI-REPO-" + uuid.NewString(), UserID: uuid.MustParse(idPemilik), DocumentTypeCode: "PKS", FormData: []byte("{}")}
	if err := dbTes.Create(&pks).Error; err != nil {
		t.Fatal(err)
	}

	perkaraAseiDilihat := []string{orangLegal, orangAdmin, orangAUAsei, orangExternal}
	pengajuanDilihat := []string{orangLegal, orangAdmin, orangPemilik}
	dokumen := []struct {
		judul, fitur, jenis, sumber, mime string
		dilihat                           []string
		id                                string
	}{
		{judul: "uji-repo-perkara-asei.pdf", fitur: "case_management", jenis: "legal_cases", sumber: perkaraAsei.ID, mime: pdf, dilihat: perkaraAseiDilihat},
		{judul: "uji-repo-foto-asei.png", fitur: "case_management", jenis: "legal_cases_photo", sumber: perkaraAsei.ID, mime: "image/png", dilihat: perkaraAseiDilihat},
		{judul: "uji-repo-kronologi-asei.pdf", fitur: "case_management", jenis: "case_chronologies", sumber: kronologi, mime: pdf, dilihat: perkaraAseiDilihat},
		{judul: "uji-repo-perkara-syariah.pdf", fitur: "case_management", jenis: "legal_cases", sumber: perkaraSyariah.ID, mime: pdf, dilihat: []string{orangLegal, orangAdmin, orangAUSya, orangExternal}},
		{judul: "uji-repo-lampiran-opini.pdf", fitur: "legal_opinion", jenis: "legal_opinion_attachments", sumber: opini.ID, mime: pdf, dilihat: pengajuanDilihat},
		{judul: "uji-repo-hasil-opini.pdf", fitur: "legal_opinion", jenis: "legal_opinion_results", sumber: opini.ID, mime: pdf, dilihat: pengajuanDilihat},
		{judul: "uji-repo-lampiran-review.pdf", fitur: "document_review", jenis: "document_review_attachments", sumber: review.ID, mime: pdf, dilihat: pengajuanDilihat},
		{judul: "uji-repo-lampiran-perjanjian.pdf", fitur: "agreement_document", jenis: "agreement_attachments", sumber: pks.ID.String(), mime: pdf, dilihat: pengajuanDilihat},
		{judul: "uji-repo-perjanjian-final.pdf", fitur: "agreement_document", jenis: "agreement_documents", sumber: pks.ID.String(), mime: pdf, dilihat: pengajuanDilihat},
		// DOCX final hanya untuk yang punya agreement_document.download_docx.all.
		{judul: "uji-repo-perjanjian-final.docx", fitur: "agreement_document", jenis: "agreement_documents", sumber: pks.ID.String(), mime: docx, dilihat: []string{orangLegal, orangAdmin}},
		// Sumber yang tidak dikenal tidak ditampilkan ke siapa pun.
		{judul: "uji-repo-fitur-asing.pdf", fitur: "fitur_asing", jenis: "tabel_asing", sumber: uuid.NewString(), mime: pdf},
	}
	for i, d := range dokumen {
		baris := entity.RepositoryDocument{Title: d.judul, FeatureCode: d.fitur, SourceType: d.jenis, SourceID: d.sumber, UploaderType: "requester", FilePath: "uji-repo/" + d.judul, FileName: d.judul, MIMEType: d.mime, IsActive: true}
		if err := dbTes.Create(&baris).Error; err != nil {
			t.Fatal(err)
		}
		dokumen[i].id = baris.ID.String()
	}
	t.Cleanup(func() { dbTes.Exec("DELETE FROM repository_documents WHERE title LIKE 'uji-repo-%'") })

	for nama, tkn := range token {
		w := panggil(t, tkn, http.MethodGet, alamat+"?search=uji-repo-", nil)
		harapKode(t, w, http.StatusOK, nama+": daftar repositori")
		var daftar []struct {
			ID string `json:"id"`
		}
		isiData(t, w, &daftar)
		diDaftar := make(map[string]bool, len(daftar))
		for _, b := range daftar {
			diDaftar[b.ID] = true
		}
		for _, d := range dokumen {
			harap := slices.Contains(d.dilihat, nama)
			if diDaftar[d.id] != harap {
				t.Errorf("%s, %s di daftar: dapat %v, harap %v", nama, d.judul, diDaftar[d.id], harap)
			}
			if kode := panggil(t, tkn, http.MethodGet, alamat+"/"+d.id, nil).Code; (kode == http.StatusOK) != harap {
				t.Errorf("%s, detail %s: dapat %d, harap terlihat=%v", nama, d.judul, kode, harap)
			}
			if !harap {
				if kode := panggil(t, tkn, http.MethodGet, alamat+"/"+d.id+"/download", nil).Code; kode != http.StatusNotFound {
					t.Errorf("%s, unduh %s: dapat %d, harap 404", nama, d.judul, kode)
				}
			}
		}
	}
}
