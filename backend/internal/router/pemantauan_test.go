package router

// Dashboard, pengingat, pengaturan notifikasi, laporan, audit log, dan
// repositori dokumen.
//
// Database tes dipakai bersama semua tes, jadi angka diperiksa lewat selisih
// sebelum dan sesudah sebuah aksi, bukan lewat nilai mutlak.

import (
	"bytes"
	"net/http"
	"testing"

	"legal-riu-portal/internal/entity"
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

type pengingat struct {
	Red []struct {
		ID             string `json:"id"`
		SubmissionType string `json:"submission_type"`
		IsRead         bool   `json:"is_read"`
	} `json:"red"`
	Total       int `json:"total"`
	UnreadTotal int `json:"unread_total"`
}

// Pengajuan yang lama tidak ditangani muncul sebagai pengingat merah, dan bisa
// ditandai sudah dibaca.
func TestPengingat(t *testing.T) {
	const surel = "tes-pengingat@example.com"
	userKhusus(t, surel)
	pemilik := masuk(t, surel)
	const alamat = "/api/v1/notifications"

	ambil := func() pengingat {
		t.Helper()
		w := panggil(t, pemilik, http.MethodGet, alamat+"?limit=100", nil)
		harapKode(t, w, http.StatusOK, "mengambil pengingat")
		var p pengingat
		isiData(t, w, &p)
		return p
	}

	baru := semuaModulPengajuan[0].buat(t, pemilik)
	if n := len(ambil().Red); n != 0 {
		t.Fatalf("pengajuan yang baru dibuat sudah menjadi pengingat merah (%d)", n)
	}

	// Dimundurkan 30 hari, melewati ambang merah bawaan (14 hari). Tanggalnya
	// relatif terhadap hari ini supaya tes tidak kedaluwarsa.
	if err := dbTes.Exec("UPDATE legal_opinions SET created_at = now() - interval '30 days', updated_at = now() - interval '30 days', status_updated_at = NULL WHERE id = ?", baru.ID).Error; err != nil {
		t.Fatal(err)
	}
	p := ambil()
	if len(p.Red) != 1 || p.Red[0].ID != baru.ID || p.Red[0].IsRead || p.UnreadTotal != 1 {
		t.Fatalf("pengingat setelah pengajuan berumur 30 hari: %+v", p)
	}

	w := panggil(t, pemilik, http.MethodPatch, alamat+"/read", map[string]string{"submission_type": p.Red[0].SubmissionType, "submission_id": baru.ID})
	harapKode(t, w, http.StatusOK, "menandai pengingat dibaca")
	if p = ambil(); len(p.Red) != 1 || !p.Red[0].IsRead || p.UnreadTotal != 0 {
		t.Errorf("pengingat setelah ditandai dibaca: %+v", p)
	}
	harapKode(t, panggil(t, pemilik, http.MethodPatch, alamat+"/read", map[string]string{}), http.StatusBadRequest, "tandai dibaca tanpa isi")
	harapKode(t, panggil(t, pemilik, http.MethodPatch, alamat+"/read-all", nil), http.StatusOK, "menandai semua dibaca")

	// ADMIN dan LEGAL memakai alamat yang sama dan melihat pengingat semua pengajuan.
	for _, peran := range []entity.UserRole{entity.RoleAdmin, entity.RoleLegal} {
		harapKode(t, panggil(t, masukPeran(t, peran), http.MethodGet, alamat, nil), http.StatusOK, "pengingat "+string(peran))
	}
}

// Lonceng dibuka oleh permission lihat pengajuan, bukan oleh peran: EXTERNAL dan
// LEGAL_AU bawaannya ditolak, dan diterima setelah diberi override.
func TestPengingatMengikutiPermission(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/notifications"

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

	t.Run("USERTidakMelihatDokumenPerkara", func(t *testing.T) {
		bugDiketahui(t, "repositori dokumen menampilkan semua dokumen ke semua peran: USER, yang tidak punya akses ke Case Management, bisa melihat dan mengunduh dokumen perkara")
		w := panggilForm(t, legal, http.MethodPost, perkaraLegal+"/"+p.ID+"/upload-document", nil, berkas{"document", "rahasia.pdf", isi})
		harapKode(t, w, http.StatusOK, "unggah dokumen perkara")
		user := masukPeran(t, entity.RoleUser)
		if id := cari(user); id != "" {
			w := panggil(t, user, http.MethodGet, alamat+"/"+id+"/download", nil)
			t.Errorf("USER, yang tidak punya akses ke Case Management, melihat dokumen perkara di repositori (unduh: %d)", w.Code)
		}
	})
}
