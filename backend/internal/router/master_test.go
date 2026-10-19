package router

// Master data yang dikelola ADMIN (divisi, perusahaan, tujuan pembuatan, jenis
// kasus, kategori kasus, jenis dokumen) dan materi hukum.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"

	"github.com/google/uuid"
)

// masterData menjelaskan satu jenis master. Keenamnya punya bentuk endpoint
// yang sama, jadi diuji dengan satu alur.
type masterData struct {
	nama        string
	alamat      string         // di bawah /api/v1/admin
	buat, ubah  map[string]any // isi yang sah untuk membuat dan mengubah
	kolomWajib  string         // kolom yang dihapus untuk menguji validasi
	csvImpor    string         // satu baris sah dan satu baris tanpa kolom wajib
	kolomDiubah string         // kolom yang nilainya dicek setelah diubah
}

var semuaMaster = []masterData{
	{
		nama: "Divisi", alamat: "/divisions", kolomWajib: "name", kolomDiubah: "name",
		buat:     map[string]any{"name": "Divisi Uji", "description": "awal"},
		ubah:     map[string]any{"name": "Divisi Uji Diubah", "description": "akhir"},
		csvImpor: "name,description\nDivisi Impor,dari impor\n,tanpa nama\n",
	},
	{
		nama: "Perusahaan", alamat: "/companies", kolomWajib: "email_domain", kolomDiubah: "name",
		buat:     map[string]any{"name": "PT Uji", "email_domain": "uji.example", "is_internal": true},
		ubah:     map[string]any{"name": "PT Uji Diubah", "email_domain": "uji.example", "is_internal": true},
		csvImpor: "name,email_domain\nPT Impor,impor.example\nPT Tanpa Domain,\n",
	},
	{
		nama: "TujuanPembuatan", alamat: "/purpose-types", kolomWajib: "name", kolomDiubah: "name",
		buat:     map[string]any{"name": "Tujuan Uji", "description": "awal"},
		ubah:     map[string]any{"name": "Tujuan Uji Diubah", "description": "akhir", "is_active": true},
		csvImpor: "name,description\nTujuan Impor,dari impor\n,tanpa nama\n",
	},
	{
		nama: "JenisKasus", alamat: "/case-types", kolomWajib: "label", kolomDiubah: "label",
		buat:     map[string]any{"code": "UJI", "label": "Jenis Uji", "is_active": true},
		ubah:     map[string]any{"code": "UJI", "label": "Jenis Uji Diubah", "is_active": true},
		csvImpor: "code,label\nIMPOR,Jenis Impor\nTANPA-LABEL,\n",
	},
	{
		nama: "KategoriKasus", alamat: "/case-categories", kolomWajib: "label", kolomDiubah: "label",
		buat:     map[string]any{"code": "UJI", "label": "Kategori Uji", "is_active": true},
		ubah:     map[string]any{"code": "UJI", "label": "Kategori Uji Diubah", "is_active": true},
		csvImpor: "code,label\nIMPOR,Kategori Impor\nTANPA-LABEL,\n",
	},
	{
		nama: "JenisDokumen", alamat: "/document-types", kolomWajib: "label", kolomDiubah: "label",
		buat:     map[string]any{"name": "jenis_uji", "label": "Jenis Dokumen Uji"},
		ubah:     map[string]any{"name": "jenis_uji", "label": "Jenis Dokumen Uji Diubah", "is_active": true},
		csvImpor: "name,label\njenis_impor,Jenis Dokumen Impor\ntanpa_label,\n",
	},
}

// idDariDaftar membaca ID dari jawaban daftar, baik yang berupa array langsung
// maupun yang dibungkus {"items": [...]}.
func idDariDaftar(t *testing.T, w *httptest.ResponseRecorder) []string {
	t.Helper()
	var mentah json.RawMessage
	isiData(t, w, &mentah)
	var baris []struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(mentah, &baris) != nil {
		var bungkus struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		if err := json.Unmarshal(mentah, &bungkus); err != nil {
			t.Fatalf("bentuk daftar tidak dikenali: %s", w.Body)
		}
		baris = bungkus.Items
	}
	ids := make([]string, 0, len(baris))
	for _, b := range baris {
		ids = append(ids, b.ID)
	}
	return ids
}

func ada(daftar []string, cari string) bool {
	for _, x := range daftar {
		if x == cari {
			return true
		}
	}
	return false
}

// Buat, lihat, ubah, impor, dan hapus untuk keenam master data.
func TestMasterData(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)

	for _, m := range semuaMaster {
		t.Run(m.nama, func(t *testing.T) {
			alamat := "/api/v1/admin" + m.alamat

			w := panggil(t, admin, http.MethodPost, alamat, m.buat)
			harapKode(t, w, http.StatusCreated, "membuat")
			var baru map[string]any
			isiData(t, w, &baru)
			id, _ := baru["id"].(string)
			if id == "" {
				t.Fatalf("jawaban membuat tidak memuat id: %s", w.Body)
			}

			kurang := map[string]any{}
			for k, v := range m.buat {
				kurang[k] = v
			}
			delete(kurang, m.kolomWajib)
			harapKode(t, panggil(t, admin, http.MethodPost, alamat, kurang), http.StatusBadRequest, "membuat tanpa "+m.kolomWajib)
			harapKode(t, panggil(t, admin, http.MethodPost, alamat, m.buat), http.StatusBadRequest, "membuat data kembar")

			w = panggil(t, admin, http.MethodGet, alamat, nil)
			harapKode(t, w, http.StatusOK, "daftar")
			if !ada(idDariDaftar(t, w), id) {
				t.Error("data baru tidak muncul di daftar")
			}
			harapKode(t, panggil(t, admin, http.MethodGet, alamat+"/"+id, nil), http.StatusOK, "detail")
			harapKode(t, panggil(t, admin, http.MethodGet, alamat+"/"+uuid.NewString(), nil), http.StatusNotFound, "detail yang tidak ada")

			w = panggil(t, admin, http.MethodPut, alamat+"/"+id, m.ubah)
			harapKode(t, w, http.StatusOK, "mengubah")
			var sesudah map[string]any
			isiData(t, panggil(t, admin, http.MethodGet, alamat+"/"+id, nil), &sesudah)
			if sesudah[m.kolomDiubah] != m.ubah[m.kolomDiubah] {
				t.Errorf("%s setelah diubah: %v", m.kolomDiubah, sesudah[m.kolomDiubah])
			}

			// Impor: baris yang sah masuk, baris tanpa kolom wajib dilaporkan.
			w = panggilForm(t, admin, http.MethodPost, alamat+"/import", nil, berkas{"file", "impor.csv", []byte(m.csvImpor)})
			harapKode(t, w, http.StatusOK, "impor")
			var hasil hasilImpor
			isiData(t, w, &hasil)
			if hasil.Imported != 1 || len(hasil.Errors) != 1 || hasil.Errors[0].Row != 3 {
				t.Errorf("hasil impor: %+v", hasil)
			}
			w = panggil(t, admin, http.MethodGet, alamat+"/import/template", nil)
			harapKode(t, w, http.StatusOK, "unduh template impor")
			if !bytes.HasPrefix(w.Body.Bytes(), []byte("PK")) {
				t.Error("template impor bukan file Excel")
			}

			harapKode(t, panggil(t, admin, http.MethodDelete, alamat+"/"+id, nil), http.StatusOK, "menghapus")
			harapKode(t, panggil(t, admin, http.MethodGet, alamat+"/"+id, nil), http.StatusNotFound, "detail setelah dihapus")
		})
	}
}

// Tanda internal sebuah perusahaan tersimpan sesuai yang diminta, lewat tambah,
// impor, maupun ubah. Tanda ini menentukan apakah usernya dibatasi ke perkara
// perusahaannya sendiri.
func TestPerusahaanNonInternal(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/admin/companies"

	internal := func(id string) bool {
		t.Helper()
		var p struct {
			IsInternal bool `json:"is_internal"`
		}
		w := panggil(t, admin, http.MethodGet, alamat+"/"+id, nil)
		harapKode(t, w, http.StatusOK, "mengambil perusahaan")
		isiData(t, w, &p)
		return p.IsInternal
	}

	// Dibuat non-internal: tersimpan non-internal, juga setelah dibaca ulang.
	w0 := panggil(t, admin, http.MethodPost, alamat, map[string]any{"name": "PT Luar Uji", "email_domain": "luar-uji.example", "is_internal": false})
	harapKode(t, w0, http.StatusCreated, "membuat perusahaan non-internal")
	var luar struct {
		ID         string `json:"id"`
		IsInternal bool   `json:"is_internal"`
	}
	isiData(t, w0, &luar)
	if luar.IsInternal || internal(luar.ID) {
		t.Error("perusahaan dibuat dengan is_internal=false, tetapi tersimpan sebagai internal")
	}

	// Perusahaan hasil impor selalu non-internal.
	w0 = panggilForm(t, admin, http.MethodPost, alamat+"/import", nil, berkas{"file", "perusahaan.csv", []byte("name,email_domain\nPT Impor Luar,impor-luar.example\n")})
	harapKode(t, w0, http.StatusOK, "impor perusahaan")
	var idImpor string
	dbTes.Raw("SELECT id FROM companies WHERE email_domain = ?", "impor-luar.example").Scan(&idImpor)
	if idImpor == "" || internal(idImpor) {
		t.Errorf("perusahaan hasil impor tersimpan sebagai internal (id %q)", idImpor)
	}

	w := panggil(t, admin, http.MethodPost, alamat, map[string]any{"name": "PT Anak Usaha Uji", "email_domain": "anak-uji.example", "is_internal": true})
	harapKode(t, w, http.StatusCreated, "membuat perusahaan")
	var p struct {
		ID         string `json:"id"`
		IsInternal bool   `json:"is_internal"`
	}
	isiData(t, w, &p)

	w = panggil(t, admin, http.MethodPut, alamat+"/"+p.ID, map[string]any{"name": "PT Anak Usaha Uji", "email_domain": "anak-uji.example", "is_internal": false})
	harapKode(t, w, http.StatusOK, "mengubah perusahaan menjadi non-internal")
	isiData(t, panggil(t, admin, http.MethodGet, alamat+"/"+p.ID, nil), &p)
	if p.IsInternal {
		t.Error("perusahaan tetap internal setelah diubah menjadi non-internal")
	}
}

// Materi hukum: dikelola LEGAL, dibaca USER.
func TestMateriHukum(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)
	user := masukPeran(t, entity.RoleUser)
	const kelola, baca = "/api/v1/legal/materials", "/api/v1/materials"

	w := panggil(t, legal, http.MethodPost, kelola, map[string]string{"title": "Materi Uji", "excerpt": "ringkas", "content": "<p>isi materi</p>"})
	harapKode(t, w, http.StatusCreated, "membuat materi")
	var m struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	isiData(t, w, &m)
	harapKode(t, panggil(t, legal, http.MethodPost, kelola, map[string]string{"title": "Tanpa isi"}), http.StatusBadRequest, "materi tanpa isi")

	// USER membaca lewat alamat umum.
	w = panggil(t, user, http.MethodGet, baca, nil)
	harapKode(t, w, http.StatusOK, "USER melihat daftar materi")
	if !ada(idDariDaftar(t, w), m.ID) {
		t.Error("materi baru tidak muncul di daftar untuk USER")
	}
	harapKode(t, panggil(t, user, http.MethodGet, baca+"/"+m.ID, nil), http.StatusOK, "USER membaca materi")

	harapKode(t, panggil(t, legal, http.MethodPut, kelola+"/"+m.ID, map[string]string{"title": "Materi Uji Diubah", "content": "<p>isi baru</p>"}), http.StatusOK, "mengubah materi")
	isiData(t, panggil(t, user, http.MethodGet, baca+"/"+m.ID, nil), &m)
	if m.Title != "Materi Uji Diubah" {
		t.Errorf("judul setelah diubah: %q", m.Title)
	}

	w = panggilForm(t, admin, http.MethodPost, "/api/v1/admin/materials/import", nil,
		berkas{"file", "materi.csv", []byte("title,excerpt,content\nMateri Impor,ringkas,isi impor\n,tanpa judul,isi\n")})
	harapKode(t, w, http.StatusOK, "impor materi")
	var hasil hasilImpor
	isiData(t, w, &hasil)
	if hasil.Imported != 1 || len(hasil.Errors) != 1 {
		t.Errorf("hasil impor materi: %+v", hasil)
	}

	harapKode(t, panggil(t, legal, http.MethodDelete, kelola+"/"+m.ID, nil), http.StatusOK, "menghapus materi")
	harapKode(t, panggil(t, user, http.MethodGet, baca+"/"+m.ID, nil), http.StatusNotFound, "membaca materi yang dihapus")
}

// Di database baru, dari tiga perusahaan bawaan hanya induk yang internal.
// Setelah itu tandanya milik ADMIN: seed hanya membuat yang belum ada.
func TestSeedPerusahaanBawaan(t *testing.T) {
	harap := map[string]bool{"indonesiare.co.id": true, "asei.co.id": false, "resyariah.co.id": false}
	cek := func(konteks string) {
		t.Helper()
		var baris []struct {
			EmailDomain string
			IsInternal  bool
		}
		dbTes.Raw("SELECT email_domain, is_internal FROM companies WHERE email_domain IN ? AND deleted_at IS NULL", []string{"indonesiare.co.id", "asei.co.id", "resyariah.co.id"}).Scan(&baris)
		if len(baris) != len(harap) {
			t.Fatalf("%s: %d perusahaan bawaan, harap %d", konteks, len(baris), len(harap))
		}
		for _, b := range baris {
			if b.IsInternal != harap[b.EmailDomain] {
				t.Errorf("%s: %s internal=%v, harap %v", konteks, b.EmailDomain, b.IsInternal, harap[b.EmailDomain])
			}
		}
	}
	cek("setelah seed di database baru")

	// Tanda internal anak usaha diatur ADMIN lewat menu. Seed berikutnya (server
	// dinyalakan ulang) tidak boleh mengembalikannya.
	const diatur = "asei.co.id"
	if err := dbTes.Exec("UPDATE companies SET is_internal = true WHERE email_domain = ?", diatur).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { dbTes.Exec("UPDATE companies SET is_internal = false WHERE email_domain = ?", diatur) })
	for putaran := 1; putaran <= 2; putaran++ {
		if err := seed.SeedCompanies(dbTes); err != nil {
			t.Fatalf("seed perusahaan putaran %d: %v", putaran, err)
		}
	}
	harap[diatur] = true
	cek("setelah seed berikutnya")

	// Perusahaan bawaan yang dihapus admin tidak membuat server gagal start, dan
	// tidak dibuat lagi.
	const dihapus = "resyariah.co.id"
	if err := dbTes.Exec("UPDATE companies SET deleted_at = now() WHERE email_domain = ?", dihapus).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { dbTes.Exec("UPDATE companies SET deleted_at = NULL WHERE email_domain = ?", dihapus) })
	if err := seed.SeedCompanies(dbTes); err != nil {
		t.Fatalf("seed setelah perusahaan bawaan dihapus: %v", err)
	}
	var jumlah int64
	dbTes.Raw("SELECT count(*) FROM companies WHERE email_domain = ?", dihapus).Scan(&jumlah)
	if jumlah != 1 {
		t.Errorf("perusahaan bawaan yang dihapus: %d baris, harap tetap 1", jumlah)
	}
}

// Tanda internal induk tidak bisa dicabut dan ia tidak bisa dihapus; kalau
// bisa, user induk kehilangan pandangan ke perkara anak usaha, atau sistem
// berjalan tanpa induk. Domainnya boleh diubah karena induk dikenali dari
// tandanya. Perusahaan lain tetap bebas diatur.
func TestPerusahaanIndukDikunci(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/admin/companies"

	type perusahaan struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		EmailDomain string `json:"email_domain"`
		IsInternal  bool   `json:"is_internal"`
		IsParent    bool   `json:"is_parent"`
	}
	baca := func(id string) perusahaan {
		t.Helper()
		var p perusahaan
		w := panggil(t, admin, http.MethodGet, alamat+"/"+id, nil)
		harapKode(t, w, http.StatusOK, "mengambil perusahaan")
		isiData(t, w, &p)
		return p
	}
	ubah := func(id, nama, domain string, internal bool) int {
		t.Helper()
		return panggil(t, admin, http.MethodPut, alamat+"/"+id, map[string]any{"name": nama, "email_domain": domain, "is_internal": internal}).Code
	}

	var semua, bertandaInduk []perusahaan
	w := panggil(t, admin, http.MethodGet, alamat, nil)
	harapKode(t, w, http.StatusOK, "daftar perusahaan")
	isiData(t, w, &semua)
	for _, p := range semua {
		if p.IsParent {
			bertandaInduk = append(bertandaInduk, p)
		}
	}
	if len(bertandaInduk) != 1 || bertandaInduk[0].EmailDomain != entity.ParentCompanyDomain {
		t.Fatalf("yang bertanda induk di daftar perusahaan: %+v, harap tepat satu dengan domain %s", bertandaInduk, entity.ParentCompanyDomain)
	}
	induk := bertandaInduk[0]

	if kode := ubah(induk.ID, induk.Name, induk.EmailDomain, false); kode != http.StatusBadRequest {
		t.Errorf("mencabut tanda internal induk: dapat %d, harap 400", kode)
	}
	if kode := ubah(induk.ID, induk.Name, "induk-pindah.example", true); kode != http.StatusOK {
		t.Errorf("mengubah domain induk: dapat %d, harap 200", kode)
	}
	if p := baca(induk.ID); !p.IsParent || p.EmailDomain != "induk-pindah.example" {
		t.Errorf("induk setelah domainnya diubah: %+v, harap tetap induk", p)
	}
	if kode := ubah(induk.ID, induk.Name, induk.EmailDomain, true); kode != http.StatusOK {
		t.Errorf("mengembalikan domain induk: dapat %d, harap 200", kode)
	}
	if kode := panggil(t, admin, http.MethodDelete, alamat+"/"+induk.ID, nil).Code; kode != http.StatusBadRequest {
		t.Errorf("menghapus induk: dapat %d, harap 400", kode)
	}
	// Yang dikunci hanya tanda internal; nama tetap boleh diubah.
	if kode := ubah(induk.ID, induk.Name+" (Persero)", induk.EmailDomain, true); kode != http.StatusOK {
		t.Errorf("mengubah nama induk: dapat %d, harap 200", kode)
	}
	if kode := ubah(induk.ID, induk.Name, induk.EmailDomain, true); kode != http.StatusOK {
		t.Errorf("mengembalikan nama induk: dapat %d, harap 200", kode)
	}
	// Dibaca langsung dari database, supaya induk yang terhapus pun terlihat.
	var kini struct {
		EmailDomain string
		IsInternal  bool
		Dihapus     bool
	}
	dbTes.Raw("SELECT email_domain, is_internal, deleted_at IS NOT NULL AS dihapus FROM companies WHERE id = ?", induk.ID).Scan(&kini)
	if kini.EmailDomain != induk.EmailDomain || !kini.IsInternal || kini.Dihapus {
		t.Errorf("induk setelah semua percobaan: %+v, harap domain %s, internal, tidak dihapus", kini, induk.EmailDomain)
	}

	w = panggil(t, admin, http.MethodPost, alamat, map[string]any{"name": "PT Bebas Uji", "email_domain": "bebas-uji.example", "is_internal": true})
	harapKode(t, w, http.StatusCreated, "membuat perusahaan lain")
	var lain perusahaan
	isiData(t, w, &lain)
	if lain.IsParent {
		t.Error("perusahaan lain bertanda induk")
	}
	if kode := ubah(lain.ID, lain.Name, "bebas-uji-2.example", false); kode != http.StatusOK {
		t.Errorf("mengubah domain dan tanda internal perusahaan lain: dapat %d, harap 200", kode)
	}
	if p := baca(lain.ID); p.IsInternal || p.EmailDomain != "bebas-uji-2.example" {
		t.Errorf("perusahaan lain setelah diubah: %+v", p)
	}
	harapKode(t, panggil(t, admin, http.MethodDelete, alamat+"/"+lain.ID, nil), http.StatusOK, "menghapus perusahaan lain")
}

// Induk dikenali dari tanda is_parent, bukan dari domain email, supaya bisa
// diganti. Isian tanda untuk database lama hanya berjalan kalau belum ada induk;
// kalau tidak, induk yang sudah diganti kembali ke Indonesia Re setiap server
// dinyalakan ulang. Database menolak dua induk.
func TestIndukDariTanda(t *testing.T) {
	jadikanInduk := func(domain string) {
		t.Helper()
		// Dua perintah, karena indeks unik menolak dua induk walau hanya sesaat.
		if err := dbTes.Exec("UPDATE companies SET is_parent = false WHERE is_parent").Error; err != nil {
			t.Fatal(err)
		}
		if domain == "" {
			return
		}
		if err := dbTes.Exec("UPDATE companies SET is_parent = true WHERE email_domain = ?", domain).Error; err != nil {
			t.Fatal(err)
		}
	}
	domainInduk := func(konteks string) string {
		t.Helper()
		var domain []string
		dbTes.Raw("SELECT email_domain FROM companies WHERE is_parent AND deleted_at IS NULL").Scan(&domain)
		if len(domain) != 1 {
			t.Fatalf("%s: perusahaan bertanda induk %v, harap tepat satu", konteks, domain)
		}
		return domain[0]
	}
	t.Cleanup(func() { jadikanInduk(entity.ParentCompanyDomain) })

	// Database lama: kolomnya baru ada dan belum ada perusahaan bertanda induk.
	jadikanInduk("")
	if err := seed.BackfillParentCompany(dbTes); err != nil {
		t.Fatal(err)
	}
	if got := domainInduk("database lama"); got != entity.ParentCompanyDomain {
		t.Errorf("database lama: induk %s, harap %s", got, entity.ParentCompanyDomain)
	}

	// Induk sudah diganti: server yang dinyalakan ulang tidak mengembalikannya.
	const pengganti = "asei.co.id"
	jadikanInduk(pengganti)
	for putaran := 1; putaran <= 2; putaran++ {
		if err := seed.SeedCompanies(dbTes); err != nil {
			t.Fatalf("seed perusahaan putaran %d: %v", putaran, err)
		}
		if err := seed.BackfillParentCompany(dbTes); err != nil {
			t.Fatalf("isian induk putaran %d: %v", putaran, err)
		}
	}
	if got := domainInduk("setelah seed berikutnya"); got != pengganti {
		t.Errorf("setelah seed berikutnya: induk %s, harap tetap %s", got, pengganti)
	}

	// Aturan yang bergantung pada induk mengikuti tandanya: LEGAL baru
	// terdaftar di induk pengganti.
	admin := masukPeran(t, entity.RoleAdmin)
	var divisi, idPengganti string
	dbTes.Raw("SELECT id FROM divisions WHERE deleted_at IS NULL LIMIT 1").Scan(&divisi)
	dbTes.Raw("SELECT id FROM companies WHERE email_domain = ?", pengganti).Scan(&idPengganti)
	w := panggil(t, admin, http.MethodPost, "/api/v1/admin/users", map[string]string{"full_name": "Legal Induk Pengganti", "email": "tes-induk-pengganti@example.com", "password": sandiTes, "position": "Staff", "division": divisi, "role": "LEGAL"})
	harapKode(t, w, http.StatusCreated, "membuat LEGAL")
	var u akun
	isiData(t, w, &u)
	t.Cleanup(func() { dbTes.Exec("UPDATE users SET deleted_at = now() WHERE id = ?", u.ID) })
	var terdaftar string
	dbTes.Raw("SELECT COALESCE(company_id::text, '') FROM users WHERE id = ?", u.ID).Scan(&terdaftar)
	if terdaftar != idPengganti {
		t.Errorf("LEGAL baru terdaftar di %q, harap induk pengganti %q", terdaftar, idPengganti)
	}

	if err := dbTes.Exec("UPDATE companies SET is_parent = true WHERE email_domain = ?", "resyariah.co.id").Error; err == nil {
		t.Error("database menerima perusahaan induk kedua")
	}
}

// Ganti induk lewat endpoint (docs/case_company_access.md): pratinjau
// menunjukkan dampaknya tanpa mengubah apa pun; pergantian ditolak selama ada
// Legal AU di tujuan, karena Legal AU tidak boleh terdaftar di induk; LEGAL ikut
// pindah, USER dan ADMIN tidak; induk lama tetap internal; pergantian tercatat
// di audit log dan bisa dibalik.
func TestGantiInduk(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/admin/companies"
	var indukAwal string
	dbTes.Raw("SELECT id FROM companies WHERE is_parent AND deleted_at IS NULL").Scan(&indukAwal)
	if indukAwal == "" {
		t.Fatal("tidak ada perusahaan induk")
	}

	var perusahaanUji, userUji []string
	t.Cleanup(func() {
		dbTes.Exec("UPDATE companies SET is_parent = false WHERE is_parent")
		dbTes.Exec("UPDATE companies SET is_parent = true WHERE id = ?", indukAwal)
		dbTes.Exec("UPDATE users SET company_id = ? WHERE role = ? AND company_id IN ?", indukAwal, entity.RoleLegal, perusahaanUji)
		dbTes.Exec("UPDATE users SET deleted_at = now() WHERE id IN ?", userUji)
		dbTes.Exec("UPDATE companies SET deleted_at = now() WHERE id IN ?", perusahaanUji)
	})
	buatPerusahaan := func(tanda string) string {
		t.Helper()
		w := panggil(t, admin, http.MethodPost, alamat, map[string]any{"name": "PT Induk Uji " + tanda, "email_domain": "induk-uji-" + tanda + ".example", "is_internal": false})
		harapKode(t, w, http.StatusCreated, "membuat perusahaan uji "+tanda)
		var p struct {
			ID string `json:"id"`
		}
		isiData(t, w, &p)
		perusahaanUji = append(perusahaanUji, p.ID)
		return p.ID
	}
	buatUser := func(surel string, peran entity.UserRole, perusahaan string) string {
		t.Helper()
		if err := tambahUser(surel, peran, uuid.MustParse(perusahaan)); err != nil {
			t.Fatal(err)
		}
		var id string
		dbTes.Raw("SELECT id FROM users WHERE email = ?", surel).Scan(&id)
		userUji = append(userUji, id)
		return id
	}
	perusahaanUser := func(id string) string {
		t.Helper()
		var perusahaan string
		dbTes.Raw("SELECT COALESCE(company_id::text, '') FROM users WHERE id = ?", id).Scan(&perusahaan)
		return perusahaan
	}
	type tanda struct {
		ID         string
		IsParent   bool
		IsInternal bool
	}
	tandaPerusahaan := func(id string) tanda {
		t.Helper()
		var p tanda
		dbTes.Raw("SELECT id, is_parent, is_internal FROM companies WHERE id = ?", id).Scan(&p)
		return p
	}
	jadikanInduk := func(id string) *httptest.ResponseRecorder {
		t.Helper()
		return panggil(t, admin, http.MethodPost, alamat+"/"+id+"/make-parent", nil)
	}

	tujuan := buatPerusahaan("tujuan")
	legal := buatUser("tes-ganti-induk-legal@example.com", entity.RoleLegal, indukAwal)
	user := buatUser("tes-ganti-induk-user@example.com", entity.RoleUser, indukAwal)
	adminLain := buatUser("tes-ganti-induk-admin@example.com", entity.RoleAdmin, indukAwal)
	au := buatUser("tes-ganti-induk-au@example.com", entity.RoleLegalAU, tujuan)
	var legalAktif int64
	dbTes.Raw("SELECT count(*) FROM users WHERE role = ? AND deleted_at IS NULL", entity.RoleLegal).Scan(&legalAktif)

	t.Run("Pratinjau", func(t *testing.T) {
		w := panggil(t, admin, http.MethodGet, alamat+"/"+tujuan+"/make-parent/preview", nil)
		harapKode(t, w, http.StatusOK, "pratinjau ganti induk")
		var p struct {
			CurrentParent *struct {
				ID string `json:"id"`
			} `json:"current_parent"`
			LegalToMove   int64 `json:"legal_to_move"`
			BlockingUsers []struct {
				ID string `json:"id"`
			} `json:"blocking_users"`
		}
		isiData(t, w, &p)
		if p.CurrentParent == nil || p.CurrentParent.ID != indukAwal {
			t.Errorf("induk sekarang di pratinjau: %+v, harap %s", p.CurrentParent, indukAwal)
		}
		if p.LegalToMove != legalAktif {
			t.Errorf("LEGAL yang akan pindah: %d, harap %d", p.LegalToMove, legalAktif)
		}
		if len(p.BlockingUsers) != 1 || p.BlockingUsers[0].ID != au {
			t.Errorf("Legal AU yang menghalangi: %+v, harap hanya %s", p.BlockingUsers, au)
		}
		if got := tandaPerusahaan(tujuan); got.IsParent || got.IsInternal {
			t.Errorf("pratinjau mengubah perusahaan tujuan: %+v", got)
		}
	})

	// Selama ada Legal AU di tujuan, ditolak dan tidak ada yang berubah.
	harapKode(t, jadikanInduk(tujuan), http.StatusBadRequest, "ganti induk selama ada Legal AU di tujuan")
	if got := tandaPerusahaan(indukAwal); !got.IsParent {
		t.Fatalf("induk berpindah walau ditolak: %+v", got)
	}
	if got := perusahaanUser(legal); got != indukAwal {
		t.Fatalf("LEGAL pindah walau pergantian ditolak: %s", got)
	}

	// Legal AU dipindah dulu, seperti yang harus dilakukan admin.
	dbTes.Exec("UPDATE users SET deleted_at = now() WHERE id = ?", au)
	w := jadikanInduk(tujuan)
	harapKode(t, w, http.StatusOK, "ganti induk")
	var hasil struct {
		PreviousParent *struct {
			ID string `json:"id"`
		} `json:"previous_parent"`
		Parent struct {
			ID string `json:"id"`
		} `json:"parent"`
		MovedLegalIDs []string `json:"moved_legal_ids"`
	}
	isiData(t, w, &hasil)
	if hasil.PreviousParent == nil || hasil.PreviousParent.ID != indukAwal || hasil.Parent.ID != tujuan {
		t.Errorf("jawaban ganti induk: lama %+v, baru %s", hasil.PreviousParent, hasil.Parent.ID)
	}
	if int64(len(hasil.MovedLegalIDs)) != legalAktif || !slices.Contains(hasil.MovedLegalIDs, legal) {
		t.Errorf("LEGAL yang dipindah: %d (%v), harap %d termasuk %s", len(hasil.MovedLegalIDs), hasil.MovedLegalIDs, legalAktif, legal)
	}

	if got := tandaPerusahaan(tujuan); !got.IsParent || !got.IsInternal {
		t.Errorf("induk baru: %+v, harap induk dan internal", got)
	}
	if got := tandaPerusahaan(indukAwal); got.IsParent || !got.IsInternal {
		t.Errorf("induk lama: %+v, harap bukan induk dan tetap internal", got)
	}
	var legalDiLuar int64
	dbTes.Raw("SELECT count(*) FROM users WHERE role = ? AND deleted_at IS NULL AND company_id IS DISTINCT FROM ?", entity.RoleLegal, tujuan).Scan(&legalDiLuar)
	if legalDiLuar != 0 {
		t.Errorf("%d LEGAL tidak ikut pindah ke induk baru", legalDiLuar)
	}
	if got := perusahaanUser(user); got != indukAwal {
		t.Errorf("USER ikut pindah ke %s, harap tetap di induk lama", got)
	}
	if got := perusahaanUser(adminLain); got != indukAwal {
		t.Errorf("ADMIN ikut pindah ke %s, harap tetap di induk lama", got)
	}

	var catatan struct {
		OldValue    string
		NewValue    string
		Description string
	}
	dbTes.Raw("SELECT old_value, new_value, description FROM audit_logs WHERE action = ? AND entity_id = ? ORDER BY created_at DESC LIMIT 1", entity.ActionParentChange, tujuan).Scan(&catatan)
	if catatan.OldValue != indukAwal || catatan.NewValue != tujuan || !strings.Contains(catatan.Description, legal) {
		t.Errorf("audit log ganti induk: %+v, harap induk lama %s, induk baru %s, dan ID LEGAL %s", catatan, indukAwal, tujuan, legal)
	}

	harapKode(t, jadikanInduk(tujuan), http.StatusBadRequest, "menjadikan induk perusahaan yang sudah induk")
	terhapus := buatPerusahaan("terhapus")
	harapKode(t, panggil(t, admin, http.MethodDelete, alamat+"/"+terhapus, nil), http.StatusOK, "menghapus perusahaan uji")
	harapKode(t, jadikanInduk(terhapus), http.StatusBadRequest, "menjadikan induk perusahaan yang dihapus")

	t.Run("Dibalik", func(t *testing.T) {
		// Legal AU yang terdaftar di induk lama dipindah dulu.
		var auDiIndukLama []string
		dbTes.Raw("SELECT id FROM users WHERE role = ? AND company_id = ? AND deleted_at IS NULL", entity.RoleLegalAU, indukAwal).Scan(&auDiIndukLama)
		parkir := buatPerusahaan("parkir")
		dbTes.Exec("UPDATE users SET company_id = ? WHERE id IN ?", parkir, auDiIndukLama)
		t.Cleanup(func() { dbTes.Exec("UPDATE users SET company_id = ? WHERE id IN ?", indukAwal, auDiIndukLama) })

		harapKode(t, jadikanInduk(indukAwal), http.StatusOK, "mengembalikan induk lama")
		if got := tandaPerusahaan(indukAwal); !got.IsParent {
			t.Errorf("induk lama setelah dibalik: %+v", got)
		}
		if got := tandaPerusahaan(tujuan); got.IsParent || !got.IsInternal {
			t.Errorf("induk sementara setelah dibalik: %+v, harap bukan induk dan tetap internal", got)
		}
		if got := perusahaanUser(legal); got != indukAwal {
			t.Errorf("LEGAL setelah dibalik terdaftar di %s, harap %s", got, indukAwal)
		}
	})

	// Dua admin menekan bersamaan untuk tujuan berbeda: hasilnya tetap satu induk.
	t.Run("Bersamaan", func(t *testing.T) {
		calon := []string{buatPerusahaan("serentak-a"), buatPerusahaan("serentak-b")}
		kode := make([]int, len(calon))
		var wg sync.WaitGroup
		for i, id := range calon {
			wg.Add(1)
			go func() {
				defer wg.Done()
				kode[i] = panggil(t, admin, http.MethodPost, alamat+"/"+id+"/make-parent", nil).Code
			}()
		}
		wg.Wait()
		var induk []string
		dbTes.Raw("SELECT id FROM companies WHERE is_parent").Scan(&induk)
		if len(induk) != 1 || !slices.Contains(calon, induk[0]) {
			t.Errorf("induk setelah dua permintaan bersamaan: %v (kode %v), harap tepat satu dari %v", induk, kode, calon)
		}
		if !slices.Contains(kode, http.StatusOK) {
			t.Errorf("tidak ada permintaan yang berhasil: kode %v", kode)
		}
	})
}
