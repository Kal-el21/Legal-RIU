package router

// Master data yang dikelola ADMIN (divisi, perusahaan, tujuan pembuatan, jenis
// kasus, kategori kasus, jenis dokumen) dan materi hukum.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// Induk dikenali dari domain surelnya (entity.ParentCompanyDomain). Tanda
// internal dan domainnya tidak bisa diubah dan ia tidak bisa dihapus; kalau
// bisa, user induk kehilangan pandangan ke perkara anak usaha, atau sistem
// berjalan tanpa induk. Perusahaan lain tetap bebas diatur.
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
	if kode := ubah(induk.ID, induk.Name, "induk-pindah.example", true); kode != http.StatusBadRequest {
		t.Errorf("mengubah domain induk: dapat %d, harap 400", kode)
	}
	if kode := panggil(t, admin, http.MethodDelete, alamat+"/"+induk.ID, nil).Code; kode != http.StatusBadRequest {
		t.Errorf("menghapus induk: dapat %d, harap 400", kode)
	}
	// Yang dikunci hanya tanda internal dan domainnya; nama tetap boleh diubah.
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
