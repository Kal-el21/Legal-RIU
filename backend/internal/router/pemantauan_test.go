package router

// Dashboard, pengingat, pengaturan notifikasi, laporan, audit log, dan
// repositori dokumen.
//
// Database tes dipakai bersama semua tes, jadi angka diperiksa lewat selisih
// sebelum dan sesudah sebuah aksi, bukan lewat nilai mutlak.

import (
	"bytes"
	"net/http"
	"net/url"
	"slices"
	"testing"
	"time"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/service"

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
		"legal-cases":      {"company", "case_type", "category", "status", "level", "location", "pic"},
	}
	for laporan, daftar := range kelompok {
		for _, k := range daftar {
			ambil(admin, "/api/v1/admin/reports/"+laporan+"?group_by="+k)
		}
	}
	// Tabel jenis dan kategori kasus punya kolom bernama label, sama dengan nama
	// kolom bulan di laporan. Pengelompokannya harus tetap per bulan: jumlah
	// perkaranya sama dengan pengelompokan lain.
	t.Run("PerkaraPerJenisDanKategori", func(t *testing.T) {
		buatPerkara(t, legal)
		harap := ambil(admin, "/api/v1/admin/reports/legal-cases?group_by=status").jumlah()
		for _, k := range []string{"case_type", "category"} {
			if got := ambil(admin, "/api/v1/admin/reports/legal-cases?group_by="+k).jumlah(); got != harap {
				t.Errorf("laporan perkara per %s: jumlah %v, harap %v seperti per status", k, got, harap)
			}
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

// Audit log dengan page dan limit yang tidak masuk akal tetap dijawab, dengan
// nilai bawaan, bukan 500 karena pembagian dengan nol.
func TestAuditLogLimitTidakWajar(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	legal := masukPeran(t, entity.RoleLegal)
	for _, k := range []struct{ token, alamat string }{
		{admin, "/api/v1/admin/audit-logs"},
		{legal, "/api/v1/legal/audit-logs"},
	} {
		for _, q := range []string{"?limit=0", "?limit=-5", "?limit=abc", "?page=0&limit=0"} {
			harapKode(t, panggil(t, k.token, http.MethodGet, k.alamat+q, nil), http.StatusOK, k.alamat+q)
		}
	}
}

// Saringan Audit Log (PR Audit Log tahap L.1): pelaku, entitas, pencarian nama pelaku, tanggal
// dalam WIB, batas limit, bentuk pelaku yang ringkas, pelaku yang akunnya dihapus, dan daftar pelaku.
func TestAuditLogSaringan(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	legal := masukPeran(t, entity.RoleLegal)
	var induk string
	dbTes.Raw("SELECT id FROM companies WHERE is_parent").Scan(&induk)
	pelaku := func(surel, nama string, peran entity.UserRole) string {
		t.Helper()
		if err := tambahUser(surel, peran, uuid.MustParse(induk)); err != nil {
			t.Fatal(err)
		}
		var id string
		dbTes.Exec("UPDATE users SET full_name = ? WHERE email = ?", nama, surel)
		dbTes.Raw("SELECT id FROM users WHERE email = ?", surel).Scan(&id)
		return id
	}
	type barisAudit struct {
		UserID     string         `json:"user_id"`
		User       map[string]any `json:"user"`
		EntityType string         `json:"entity_type"`
		EntityID   string         `json:"entity_id"`
	}
	ambil := func(token, alamat string) ([]barisAudit, int) {
		t.Helper()
		w := panggil(t, token, http.MethodGet, alamat, nil)
		harapKode(t, w, http.StatusOK, alamat)
		var h struct {
			Items []barisAudit `json:"items"`
			Limit int          `json:"limit"`
		}
		isiData(t, w, &h)
		return h.Items, h.Limit
	}
	const dasar = "/api/v1/admin/audit-logs?limit=100"
	semuaMilik := func(items []barisAudit, cocok func(barisAudit) bool) bool {
		return len(items) > 0 && !slices.ContainsFunc(items, func(b barisAudit) bool { return !cocok(b) })
	}

	const surelA, surelB = "tes-audit-saring-a@example.com", "tes-audit-saring-b@example.com"
	idA := pelaku(surelA, "Pelaku Saringan Alfa", entity.RoleUser)
	idB := pelaku(surelB, "Pelaku Saringan Beta", entity.RoleUser)
	tokenA := masuk(t, surelA) // login tercatat atas nama A
	masuk(t, surelB)

	t.Run("Pelaku", func(t *testing.T) {
		items, _ := ambil(admin, dasar+"&user_id="+idA)
		if !semuaMilik(items, func(b barisAudit) bool { return b.UserID == idA }) {
			t.Errorf("saringan pelaku A: %+v", items)
		}
		if got := items[0].User; len(got) != 4 || got["full_name"] != "Pelaku Saringan Alfa" || got["email"] != surelA || got["role"] != "USER" {
			t.Errorf("pelaku di jawaban harus berisi tepat id, full_name, email, role: %v", got)
		}
	})

	t.Run("Entitas", func(t *testing.T) {
		m := semuaModulPengajuan[0]
		p := m.buat(t, tokenA)
		m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
		items, _ := ambil(admin, dasar+"&entity_id="+p.ID)
		if len(items) < 2 || !semuaMilik(items, func(b barisAudit) bool { return b.EntityID == p.ID }) {
			t.Errorf("saringan entitas %s, harap pembuatan dan perubahan status: %+v", p.ID, items)
		}
	})

	t.Run("CariNamaPelaku", func(t *testing.T) {
		items, _ := ambil(admin, dasar+"&search="+url.QueryEscape("Saringan Alfa"))
		if !semuaMilik(items, func(b barisAudit) bool { return b.UserID == idA }) {
			t.Errorf("cari nama pelaku A: %+v", items)
		}
	})

	t.Run("TanggalWIB", func(t *testing.T) {
		// 06.30 WIB tanggal 15 masih tanggal 14 dalam UTC.
		if err := dbTes.Exec(`INSERT INTO audit_logs (id, user_id, action, entity_type, entity_id, ip_address, user_agent, created_at, updated_at)
			VALUES (gen_random_uuid(), ?, 'USER_UPDATE', 'uji_zona_waktu', gen_random_uuid(), '', '', '2026-01-15 06:30:00+07', now())`, idA).Error; err != nil {
			t.Fatal(err)
		}
		const jenis = dasar + "&entity_type=uji_zona_waktu"
		if items, _ := ambil(admin, jenis+"&date_from=2026-01-15&date_to=2026-01-15"); len(items) != 1 {
			t.Errorf("baris 06.30 WIB tanggal 15 tidak ada di saringan tanggal 15: %+v", items)
		}
		if items, _ := ambil(admin, jenis+"&date_from=2026-01-14&date_to=2026-01-14"); len(items) != 0 {
			t.Errorf("baris 06.30 WIB tanggal 15 masuk saringan tanggal 14: %+v", items)
		}
	})

	t.Run("BatasDanIDTidakValid", func(t *testing.T) {
		if _, limit := ambil(admin, "/api/v1/admin/audit-logs?limit=1000"); limit != 100 {
			t.Errorf("limit 1000 dijawab dengan limit %d, harap dibatasi 100", limit)
		}
		for _, q := range []string{"user_id=bukan-uuid", "entity_id=bukan-uuid"} {
			harapKode(t, panggil(t, admin, http.MethodGet, "/api/v1/admin/audit-logs?"+q, nil), http.StatusBadRequest, q)
		}
	})

	t.Run("PelakuTerhapus", func(t *testing.T) {
		if err := dbTes.Exec("UPDATE users SET deleted_at = now() WHERE id = ?", idB).Error; err != nil {
			t.Fatal(err)
		}
		items, _ := ambil(admin, dasar+"&user_id="+idB)
		if len(items) == 0 || items[0].User["full_name"] != "Pelaku Saringan Beta" {
			t.Errorf("nama pelaku yang akunnya dihapus hilang: %+v", items)
		}
	})

	t.Run("DaftarPelaku", func(t *testing.T) {
		daftar := func(token, alamat string, kode int) []map[string]any {
			t.Helper()
			w := panggil(t, token, http.MethodGet, alamat, nil)
			harapKode(t, w, kode, alamat)
			var d []map[string]any
			if kode == http.StatusOK {
				isiData(t, w, &d)
			}
			return d
		}
		const cari = "/audit-logs/actors?q=" + "Saringan"
		d := daftar(admin, "/api/v1/admin"+cari, http.StatusOK)
		var nama []any
		for _, p := range d {
			if len(p) != 4 {
				t.Errorf("isi pelaku harus tepat id, full_name, email, role: %v", p)
			}
			nama = append(nama, p["full_name"])
		}
		if !slices.Contains(nama, "Pelaku Saringan Alfa") || !slices.Contains(nama, "Pelaku Saringan Beta") || len(d) > 10 {
			t.Errorf("daftar pelaku 'Saringan' (termasuk akun terhapus, maks. 10): %v", nama)
		}
		if d := daftar(legal, "/api/v1/legal"+cari, http.StatusOK); len(d) == 0 {
			t.Error("LEGAL tidak mendapat daftar pelaku")
		}
		daftar(masukPeran(t, entity.RoleUser), "/api/v1/legal"+cari, http.StatusForbidden)

		// LEGAL yang izin audit_log.view-nya dicabut tidak mendapat daftar pelaku maupun audit log.
		const surelC = "tes-audit-saring-legal@example.com"
		idC := pelaku(surelC, "Legal Tanpa Audit", entity.RoleLegal)
		harapKode(t, panggil(t, admin, http.MethodPut, "/api/v1/admin/users/"+idC+"/permissions",
			map[string]any{"overrides": []map[string]string{{"code": "audit_log.view", "effect": "DENY"}}}), http.StatusOK, "mencabut audit_log.view")
		tokenC := masuk(t, surelC)
		daftar(tokenC, "/api/v1/legal"+cari, http.StatusForbidden)
		daftar(tokenC, "/api/v1/legal/audit-logs", http.StatusForbidden)
	})
}

// Riwayat per data di halaman detail (PR Audit Log tahap L.4): yang boleh membuka detail melihat
// riwayatnya, terbaru di atas; yang tidak boleh mendapat jawaban yang sama dengan detail; IP hanya
// untuk yang punya audit_log.view (keputusan L3 dan L7).
func TestRiwayatPerData(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)
	type baris struct {
		Action    string `json:"action"`
		EntityID  string `json:"entity_id"`
		IPAddress string `json:"ip_address"`
		UserAgent string `json:"user_agent"`
	}
	riwayat := func(t *testing.T, token, alamat string, kode int) []baris {
		t.Helper()
		w := panggil(t, token, http.MethodGet, alamat, nil)
		harapKode(t, w, kode, alamat)
		var r []baris
		if kode == http.StatusOK {
			isiData(t, w, &r)
		}
		return r
	}
	// periksa: baris milik data itu, aksi terbaru di atas, dan IP terisi hanya untuk auditor.
	periksa := func(t *testing.T, r []baris, id, aksiTerbaru string, auditor bool) {
		t.Helper()
		if len(r) < 2 || r[0].Action != aksiTerbaru {
			t.Fatalf("riwayat %s: %+v, harap minimal 2 baris dengan %s paling atas", id, r, aksiTerbaru)
		}
		for _, b := range r {
			if b.EntityID != id {
				t.Errorf("riwayat %s memuat baris data lain: %+v", id, b)
			}
			if (b.IPAddress != "") != auditor {
				t.Errorf("riwayat %s: IP %q untuk auditor=%v", id, b.IPAddress, auditor)
			}
		}
	}

	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			p := m.buat(t, pemilik)
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			periksa(t, riwayat(t, pemilik, m.url("", "/", p.ID, "/history"), http.StatusOK), p.ID, "STATUS_CHANGE", false)
			periksa(t, riwayat(t, legal, m.url("/legal", "/", p.ID, "/history"), http.StatusOK), p.ID, "STATUS_CHANGE", true)
			periksa(t, riwayat(t, admin, m.url("/admin", "/", p.ID, "/history"), http.StatusOK), p.ID, "STATUS_CHANGE", true)
			riwayat(t, lain, m.url("", "/", p.ID, "/history"), http.StatusNotFound)
			for _, id := range []string{uuid.NewString(), "bukan-uuid"} {
				riwayat(t, legal, m.url("/legal", "/", id, "/history"), http.StatusNotFound)
			}

			// Draf hanya untuk pemiliknya, juga riwayatnya.
			d := m.buat(t, pemilik)
			jadikanDraf(t, m.tabel, d.ID)
			riwayat(t, pemilik, m.url("", "/", d.ID, "/history"), http.StatusOK)
			riwayat(t, legal, m.url("/legal", "/", d.ID, "/history"), http.StatusNotFound)
		})
	}

	t.Run("Perjanjian", func(t *testing.T) {
		butuhMinIO(t)
		p := buatPerjanjian(t, pemilik, formPerjanjian())
		statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
		periksa(t, riwayat(t, pemilik, urlPerjanjian("", "/", p.ID, "/history"), http.StatusOK), p.ID, "STATUS_CHANGE", false)
		periksa(t, riwayat(t, legal, urlPerjanjian("/legal", "/", p.ID, "/history"), http.StatusOK), p.ID, "STATUS_CHANGE", true)
		periksa(t, riwayat(t, admin, urlPerjanjian("/admin", "/", p.ID, "/history"), http.StatusOK), p.ID, "STATUS_CHANGE", true)
		riwayat(t, lain, urlPerjanjian("", "/", p.ID, "/history"), http.StatusNotFound)
	})

	t.Run("Perkara", func(t *testing.T) {
		perk := buatPerkara(t, legal)
		harapKode(t, panggil(t, masukPeran(t, entity.RoleLegalAU), http.MethodPatch, perkaraLegalAU+"/"+perk.ID+"/status",
			map[string]string{"current_status": "Banding"}), http.StatusOK, "Legal AU mengubah status")
		periksa(t, riwayat(t, legal, perkaraLegal+"/"+perk.ID+"/history", http.StatusOK), perk.ID, "STATUS_CHANGE", true)
		periksa(t, riwayat(t, admin, perkaraAdmin+"/"+perk.ID+"/history", http.StatusOK), perk.ID, "STATUS_CHANGE", true)
		periksa(t, riwayat(t, masukPeran(t, entity.RoleLegalAU), perkaraLegalAU+"/"+perk.ID+"/history", http.StatusOK), perk.ID, "STATUS_CHANGE", false)
		periksa(t, riwayat(t, masukPeran(t, entity.RoleExternal), perkaraExternal+"/"+perk.ID+"/history", http.StatusOK), perk.ID, "STATUS_CHANGE", false)
		// USER tidak punya izin melihat perkara.
		riwayat(t, pemilik, "/api/v1/legal-cases/"+perk.ID+"/history", http.StatusForbidden)
		riwayat(t, legal, perkaraLegal+"/"+uuid.NewString()+"/history", http.StatusNotFound)
	})
}

// File yang tercatat di database tetapi tidak ada di penyimpanan dijawab 404.
// MinIO baru melapor "tidak ada" saat isi file dibaca; sebelumnya jawaban 200
// sudah telanjur terkirim dan browser menyimpan file 0 KB.
func TestUnduhFileYangHilang(t *testing.T) {
	butuhMinIO(t)
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)
	pemilik := masukPeran(t, entity.RoleUser)
	hilang := func(folder string) string { return folder + "/hilang-" + uuid.NewString() + ".pdf" }

	// Perkara: dokumen terdaftar, filenya tidak ada.
	p := buatPerkara(t, legal)
	pathPerkara := hilang("legal-cases/documents")
	dbTes.Exec("UPDATE legal_cases SET document_link = ? WHERE id = ?", pathPerkara, p.ID)

	// Legal Opinion: lampiran terdaftar, filenya tidak ada.
	opini := semuaModulPengajuan[0].buat(t, pemilik, berkas{"attachments", "ada.pdf", []byte("isi")})
	pathOpini := hilang("legal-opinions/attachments")
	dbTes.Exec("UPDATE legal_opinion_attachments SET file_path = ? WHERE legal_opinion_id = ?", pathOpini, opini.ID)

	// Repositori: dokumen perkara yang filenya tidak ada.
	repo := entity.RepositoryDocument{Title: "hilang.pdf", FeatureCode: "case_management", SourceType: "legal_cases", SourceID: p.ID, FilePath: pathPerkara, FileName: "hilang.pdf", MIMEType: "application/pdf", IsActive: true}
	if err := dbTes.Create(&repo).Error; err != nil {
		t.Fatal(err)
	}

	// Perjanjian selesai, file final dan lampirannya tidak ada.
	var idPemilik string
	dbTes.Raw("SELECT id FROM users WHERE email = ?", surelPeran(entity.RoleUser)).Scan(&idPemilik)
	pks := entity.AgreementDocument{TicketNumber: "UJI-HILANG-" + uuid.NewString(), UserID: uuid.MustParse(idPemilik), DocumentTypeCode: "PKS", FormData: []byte("{}"), Status: entity.StatusCompleted,
		GeneratedPDFPath: hilang("agreement-documents/final"), GeneratedDOCXPath: hilang("agreement-documents/final"), GeneratedFileName: "perjanjian-hilang"}
	if err := dbTes.Create(&pks).Error; err != nil {
		t.Fatal(err)
	}
	lampiran := entity.AgreementAttachment{ID: uuid.New(), AgreementDocumentID: pks.ID, FileName: "hilang.pdf", FilePath: hilang("agreement-documents/attachments"), UploadRound: 1, UploadedBy: pks.UserID}
	if err := dbTes.Create(&lampiran).Error; err != nil {
		t.Fatal(err)
	}

	for _, k := range []struct{ nama, token, alamat string }{
		{"file perkara (unduh)", legal, perkaraLegal + "/download?path=" + url.QueryEscape(pathPerkara)},
		{"file perkara (lihat)", legal, perkaraLegal + "/file?path=" + url.QueryEscape(pathPerkara)},
		{"lampiran Legal Opinion", pemilik, semuaModulPengajuan[0].url("", "/download?path=", url.QueryEscape(pathOpini))},
		{"Repositori Dokumen", legal, "/api/v1/document-repositories/" + repo.ID.String() + "/download"},
		{"PDF final perjanjian", admin, "/api/v1/admin/agreement-documents/" + pks.ID.String() + "/pdf"},
		{"DOCX final perjanjian", admin, "/api/v1/admin/agreement-documents/" + pks.ID.String() + "/docx"},
		{"lampiran perjanjian", admin, "/api/v1/admin/agreement-documents/" + pks.ID.String() + "/attachments/" + lampiran.ID.String()},
	} {
		if w := panggil(t, k.token, http.MethodGet, k.alamat, nil); w.Code != http.StatusNotFound {
			t.Errorf("%s yang filenya hilang: dapat %d (%d byte), harap 404", k.nama, w.Code, w.Body.Len())
		}
	}
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
	w = panggil(t, legal, http.MethodGet, alamat+"/"+id, nil)
	harapKode(t, w, http.StatusOK, "metadata dokumen repositori")
	var metadata struct {
		SourceNumber string     `json:"source_number"`
		SourceTitle  string     `json:"source_title"`
		UploadedAt   *time.Time `json:"uploaded_at"`
	}
	isiData(t, w, &metadata)
	if metadata.SourceNumber != p.TicketNumber || metadata.SourceTitle != p.CaseName || metadata.UploadedAt == nil {
		t.Errorf("metadata dokumen perkara: %+v, harap nomor=%q judul=%q dan waktu unggah terisi",
			metadata, p.TicketNumber, p.CaseName)
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

func TestRepositoriMetadataSumber(t *testing.T) {
	repoDoc := service.NewRepositoryDocumentService(repository.NewRepositoryDocumentRepository(dbTes), nil)
	waktu := time.Date(2026, 8, 14, 9, 30, 0, 0, time.UTC)
	approvedAt := waktu.Add(time.Hour)
	caseEntity := &entity.LegalCase{
		Base:         entity.Base{ID: uuid.New()},
		TicketNumber: "PKR-REPO-1",
		CaseName:     "Perkara Metadata",
		DocumentLink: "uji-metadata/perkara.pdf",
		Photo:        "uji-metadata/perkara.png",
	}
	chronology := &entity.CaseChronology{Base: entity.Base{ID: uuid.New()}}
	lo := &entity.LegalOpinion{
		Base:         entity.Base{ID: uuid.New()},
		TicketNumber: "LO-REPO-1",
		Title:        "Opini Awal",
		Attachments: []entity.LegalOpinionAttachment{{
			FilePath: "uji-metadata/lo.pdf", FileName: "Nama Asli LO.pdf", CreatedAt: waktu,
		}},
		Results: []entity.LegalOpinionResult{{
			FilePath: "uji-metadata/lo-hasil.pdf", FileName: "Hasil LO.pdf", CreatedAt: waktu,
		}},
	}
	review := &entity.DocumentReview{
		Base:         entity.Base{ID: uuid.New()},
		TicketNumber: "RD-REPO-1",
		DocumentName: "Dokumen Review",
		Attachments: []entity.DocumentReviewAttachment{{
			FilePath: "uji-metadata/rd.pdf", FileName: "Lampiran RD.pdf", CreatedAt: waktu,
		}},
		Results: []entity.DocumentReviewResult{{
			FilePath: "uji-metadata/rd-hasil.pdf", FileName: "Hasil RD.pdf", CreatedAt: waktu,
		}},
	}
	agreement := &entity.AgreementDocument{
		Base:              entity.Base{ID: uuid.New()},
		TicketNumber:      "PKS-REPO-1",
		FormData:          []byte(`{"pihak_kedua_nama":"Mitra Perjanjian"}`),
		GeneratedDOCXPath: "uji-metadata/pks.docx",
		GeneratedPDFPath:  "uji-metadata/pks.pdf",
		GeneratedFileName: "PKS Final",
		ApprovedAt:        &approvedAt,
		Attachments: []entity.AgreementAttachment{{
			FilePath: "uji-metadata/pks-lampiran.pdf", FileName: "Lampiran PKS.pdf", CreatedAt: waktu,
		}},
	}

	for nama, err := range map[string]error{
		"Legal Opinion":   repoDoc.SyncLegalOpinion(lo),
		"Review Dokumen":  repoDoc.SyncDocumentReview(review),
		"Perjanjian":      repoDoc.SyncAgreement(agreement),
		"Dokumen Perkara": repoDoc.SyncCaseDocument(caseEntity, "uji-metadata/perkara.pdf", "Dokumen Perkara.pdf", 10, waktu),
		"Foto Perkara":    repoDoc.SyncCasePhoto(caseEntity, "uji-metadata/perkara.png", "Foto Perkara.png", 10, waktu),
		"Kronologi": repoDoc.SyncChronologyDocument(caseEntity, chronology,
			"uji-metadata/kronologi.pdf", "Kronologi.pdf", &waktu),
	} {
		if err != nil {
			t.Errorf("%s: gagal menyinkronkan metadata: %v", nama, err)
		}
	}
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM repository_documents WHERE file_path LIKE ?", "uji-metadata/%")
	})

	type metadata struct {
		SourceNumber string
		SourceTitle  string
		UploadedAt   time.Time
		FileName     string
	}
	periksa := func(path, nomor, judul, namaFile string, uploaded time.Time) {
		t.Helper()
		var got metadata
		if err := dbTes.Raw(`SELECT source_number, source_title, uploaded_at, file_name
			FROM repository_documents WHERE file_path = ?`, path).Scan(&got).Error; err != nil {
			t.Fatalf("membaca metadata %s: %v", path, err)
		}
		if got.SourceNumber != nomor || got.SourceTitle != judul || !got.UploadedAt.Equal(uploaded) || got.FileName != namaFile {
			t.Errorf("%s: metadata dapat %+v, harap nomor=%q judul=%q file=%q waktu=%s",
				path, got, nomor, judul, namaFile, uploaded.Format(time.RFC3339))
		}
	}
	periksa("uji-metadata/lo.pdf", lo.TicketNumber, lo.Title, "Nama Asli LO.pdf", waktu)
	periksa("uji-metadata/lo-hasil.pdf", lo.TicketNumber, lo.Title, "Hasil LO.pdf", waktu)
	periksa("uji-metadata/rd.pdf", review.TicketNumber, review.DocumentName, "Lampiran RD.pdf", waktu)
	periksa("uji-metadata/rd-hasil.pdf", review.TicketNumber, review.DocumentName, "Hasil RD.pdf", waktu)
	periksa("uji-metadata/pks-lampiran.pdf", agreement.TicketNumber, "Mitra Perjanjian", "Lampiran PKS.pdf", waktu)
	periksa("uji-metadata/pks.docx", agreement.TicketNumber, "Mitra Perjanjian", "PKS Final.docx", approvedAt)
	periksa("uji-metadata/pks.pdf", agreement.TicketNumber, "Mitra Perjanjian", "PKS Final.pdf", approvedAt)
	periksa("uji-metadata/perkara.pdf", caseEntity.TicketNumber, caseEntity.CaseName, "Dokumen Perkara.pdf", waktu)
	periksa("uji-metadata/perkara.png", caseEntity.TicketNumber, caseEntity.CaseName, "Foto Perkara.png", waktu)
	periksa("uji-metadata/kronologi.pdf", caseEntity.TicketNumber, caseEntity.CaseName, "Kronologi.pdf", waktu)

	lo.Title = "Opini Diperbarui"
	if err := repoDoc.SyncLegalOpinion(lo); err != nil {
		t.Fatal(err)
	}
	periksa("uji-metadata/lo.pdf", lo.TicketNumber, lo.Title, "Nama Asli LO.pdf", waktu)

	caseEntity.CaseName = "Perkara Diperbarui"
	if err := repoDoc.SyncCaseMetadata(caseEntity); err != nil {
		t.Fatal(err)
	}
	periksa("uji-metadata/perkara.pdf", caseEntity.TicketNumber, caseEntity.CaseName, "Dokumen Perkara.pdf", waktu)
	periksa("uji-metadata/perkara.png", caseEntity.TicketNumber, caseEntity.CaseName, "Foto Perkara.png", waktu)
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

// Laporan Legal Opinion, Review Dokumen, dan perkara tidak menghitung data yang sudah dihapus.
// Ketiganya memakai db.Table, yang melewati penyaring soft delete GORM (PR A tahap A.2).
func TestLaporanTanpaDataTerhapus(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	user := masukPeran(t, entity.RoleUser)
	jumlahLaporan := func(laporan string) float64 {
		t.Helper()
		w := panggil(t, legal, http.MethodGet, "/api/v1/legal/reports/"+laporan+"?group_by=status", nil)
		harapKode(t, w, http.StatusOK, "laporan "+laporan)
		var g grafik
		isiData(t, w, &g)
		return g.jumlah()
	}
	kasus := []struct {
		laporan string
		buat    func() (hapus string, token string)
	}{
		{"legal-opinions", func() (string, string) {
			return semuaModulPengajuan[0].url("", "/", semuaModulPengajuan[0].buat(t, user).ID), user
		}},
		{"document-reviews", func() (string, string) {
			return semuaModulPengajuan[1].url("", "/", semuaModulPengajuan[1].buat(t, user).ID), user
		}},
		{"legal-cases", func() (string, string) {
			return perkaraLegal + "/" + buatPerkara(t, legal).ID, legal
		}},
	}
	for _, k := range kasus {
		t.Run(k.laporan, func(t *testing.T) {
			sebelum := jumlahLaporan(k.laporan)
			alamat, token := k.buat()
			if got := jumlahLaporan(k.laporan); got != sebelum+1 {
				t.Fatalf("laporan sesudah satu data baru: %v, harap %v", got, sebelum+1)
			}
			harapKode(t, panggil(t, token, http.MethodDelete, alamat, nil), http.StatusOK, "menghapus")
			if got := jumlahLaporan(k.laporan); got != sebelum {
				t.Errorf("laporan sesudah datanya dihapus: %v, harap %v", got, sebelum)
			}
		})
	}
}
