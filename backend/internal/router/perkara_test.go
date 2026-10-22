package router

// Alur Case Management: perkara, kronologi, dokumen dan foto, data rujukan
// (cedant dan kabupaten/kota), dan pembatasan per perusahaan.

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// Awalan alamat perkara untuk tiap peran. LEGAL_AU memakai "/legal-au/cases",
// peran lain memakai "<awalan>/legal-cases".
const (
	perkaraLegal    = "/api/v1/legal/legal-cases"
	perkaraAdmin    = "/api/v1/admin/legal-cases"
	perkaraExternal = "/api/v1/external/legal-cases"
	perkaraLegalAU  = "/api/v1/legal-au/cases"
)

type perkara struct {
	ID            string `json:"id"`
	TicketNumber  string `json:"ticket_number"`
	CaseName      string `json:"case_name"`
	CurrentStatus string `json:"current_status"`
	CompanyID     string `json:"company_id"`
	DocumentLink  string `json:"document_link"`
	Photo         string `json:"photo"`
	Chronologies  []struct {
		ID string `json:"id"`
	} `json:"chronologies"`
}

type kronologi struct {
	ID        string   `json:"id"`
	Agenda    string   `json:"agenda"`
	Documents []string `json:"documents"`
}

type hasilImpor struct {
	Imported int `json:"imported"`
	Errors   []struct {
		Row   int    `json:"row"`
		Field string `json:"field"`
	} `json:"errors"`
}

// formPerkara mengembalikan isi form perkara yang sah. Data rujukannya diambil
// dari hasil seed, kecuali cedant yang dibuat lewat API.
func formPerkara(t *testing.T) map[string]any {
	t.Helper()
	idPertama := func(tabel string) string {
		var id string
		if err := dbTes.Raw("SELECT id FROM " + tabel + " WHERE deleted_at IS NULL ORDER BY created_at LIMIT 1").Scan(&id).Error; err != nil || id == "" {
			t.Fatalf("tidak ada data di tabel %s: %v", tabel, err)
		}
		return id
	}
	return map[string]any{
		"case_name": "Perkara tes", "related_party_id": cedantTes(t),
		"category_id": idPertama("case_categories"), "case_type_id": idPertama("case_types"),
		"pic": idPertama("divisions"), "location_regency_id": idPertama("regencies"),
		"company_id": idPertama("companies"),
		"case_date":  "2026-10-01", "level": "Pengadilan Negeri", "current_status": "Berjalan", "case_value": 1000000,
	}
}

var idCedantTes string

func cedantTes(t *testing.T) string {
	t.Helper()
	if idCedantTes == "" {
		w := panggil(t, masukPeran(t, entity.RoleLegal), http.MethodPost, perkaraLegal+"/cedants", map[string]string{"name": "Cedant Tes"})
		harapKode(t, w, http.StatusCreated, "membuat cedant")
		var c struct {
			ID string `json:"id"`
		}
		isiData(t, w, &c)
		idCedantTes = c.ID
	}
	return idCedantTes
}

func buatPerkara(t *testing.T, token string) perkara {
	t.Helper()
	w := panggil(t, token, http.MethodPost, perkaraLegal, formPerkara(t))
	harapKode(t, w, http.StatusCreated, "membuat perkara")
	var p perkara
	isiData(t, w, &p)
	return p
}

func ambilPerkara(t *testing.T, token, alamat, id string) perkara {
	t.Helper()
	w := panggil(t, token, http.MethodGet, alamat+"/"+id, nil)
	harapKode(t, w, http.StatusOK, "mengambil perkara")
	var p perkara
	isiData(t, w, &p)
	return p
}

func perkaraDiDaftar(t *testing.T, token, alamat, id string) bool {
	t.Helper()
	w := panggil(t, token, http.MethodGet, alamat+"?limit=100", nil)
	harapKode(t, w, http.StatusOK, "mengambil daftar perkara")
	var daftar struct {
		Items []perkara `json:"items"`
	}
	isiData(t, w, &daftar)
	for _, p := range daftar.Items {
		if p.ID == id {
			return true
		}
	}
	return false
}

// xlsx menyusun file Excel satu sheet dari baris-baris teks.
func xlsx(t *testing.T, baris ...[]string) []byte {
	t.Helper()
	wb := excelize.NewFile()
	defer wb.Close()
	for i, b := range baris {
		for j, nilai := range b {
			sel, _ := excelize.CoordinatesToCellName(j+1, i+1)
			if err := wb.SetCellValue("Sheet1", sel, nilai); err != nil {
				t.Fatal(err)
			}
		}
	}
	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// Alur tanpa file: buat, validasi, ubah, status, kronologi, impor, PDF, hapus.
func TestAlurPerkara(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)
	legalAU := masukPeran(t, entity.RoleLegalAU)
	external := masukPeran(t, entity.RoleExternal)

	p := buatPerkara(t, legal)
	if !strings.HasPrefix(p.TicketNumber, "LC-") || p.CurrentStatus != "Berjalan" {
		t.Fatalf("perkara baru: tiket %q, status %q", p.TicketNumber, p.CurrentStatus)
	}

	t.Run("Validasi", func(t *testing.T) {
		kasus := map[string]func(map[string]any){
			"nama kosong":           func(f map[string]any) { delete(f, "case_name") },
			"cedant tidak ada":      func(f map[string]any) { f["related_party_id"] = uuid.NewString() },
			"kabupaten bukan UUID":  func(f map[string]any) { f["location_regency_id"] = "bukan-uuid" },
			"tanggal tidak terbaca": func(f map[string]any) { f["case_date"] = "tiga puluh februari" },
		}
		for nama, rusak := range kasus {
			form := formPerkara(t)
			rusak(form)
			harapKode(t, panggil(t, legal, http.MethodPost, perkaraLegal, form), http.StatusBadRequest, nama)
		}
	})

	t.Run("LihatDanUbah", func(t *testing.T) {
		// Semua peran yang mengelola perkara melihat perkara yang sama.
		for _, k := range []struct{ token, alamat string }{{legal, perkaraLegal}, {admin, perkaraAdmin}, {legalAU, perkaraLegalAU}, {external, perkaraExternal}} {
			ambilPerkara(t, k.token, k.alamat, p.ID)
			if !perkaraDiDaftar(t, k.token, k.alamat, p.ID) {
				t.Errorf("perkara tidak muncul di daftar %s", k.alamat)
			}
		}
		harapKode(t, panggil(t, legal, http.MethodGet, perkaraLegal+"/"+uuid.NewString(), nil), http.StatusNotFound, "perkara yang tidak ada")
		harapKode(t, panggil(t, legal, http.MethodGet, perkaraLegal+"/latest", nil), http.StatusOK, "perkara terbaru")

		form := formPerkara(t)
		form["case_name"] = "Perkara diubah"
		harapKode(t, panggil(t, legal, http.MethodPut, perkaraLegal+"/"+p.ID, form), http.StatusOK, "mengubah perkara")
		if got := ambilPerkara(t, legal, perkaraLegal, p.ID).CaseName; got != "Perkara diubah" {
			t.Errorf("nama setelah diubah: %q", got)
		}

		// Pencarian dan saringan daftar.
		for cari, harap := range map[string]bool{"?search=diubah": true, "?search=tidak-ada-yang-cocok": false} {
			w := panggil(t, legal, http.MethodGet, perkaraLegal+cari, nil)
			harapKode(t, w, http.StatusOK, "mencari perkara")
			var daftar struct {
				Items []perkara `json:"items"`
			}
			isiData(t, w, &daftar)
			if (len(daftar.Items) > 0) != harap {
				t.Errorf("pencarian %s: %d hasil", cari, len(daftar.Items))
			}
		}
	})

	t.Run("Status", func(t *testing.T) {
		w := panggil(t, legalAU, http.MethodPatch, perkaraLegalAU+"/"+p.ID+"/status", map[string]string{"current_status": "Selesai"})
		harapKode(t, w, http.StatusOK, "LEGAL_AU mengubah status")
		if got := ambilPerkara(t, legal, perkaraLegal, p.ID).CurrentStatus; got != "Selesai" {
			t.Errorf("status setelah diubah: %q", got)
		}
		harapKode(t, panggil(t, legalAU, http.MethodPatch, perkaraLegalAU+"/"+p.ID+"/status", map[string]string{}), http.StatusBadRequest, "status kosong")
		harapKode(t, panggil(t, legalAU, http.MethodPatch, perkaraLegalAU+"/"+uuid.NewString()+"/status", map[string]string{"current_status": "Selesai"}), http.StatusNotFound, "status perkara yang tidak ada")
	})

	t.Run("Kronologi", func(t *testing.T) {
		alamat := perkaraLegal + "/" + p.ID + "/chronology"
		w := panggil(t, legal, http.MethodPost, alamat, map[string]string{"agenda_date": "2026-10-05", "agenda": "Sidang pertama", "description": "Pembacaan gugatan"})
		harapKode(t, w, http.StatusCreated, "menambah kronologi")
		var k kronologi
		isiData(t, w, &k)

		harapKode(t, panggil(t, legal, http.MethodPost, alamat, map[string]string{"agenda_date": "bukan tanggal", "agenda": "x"}), http.StatusBadRequest, "tanggal agenda salah")
		harapKode(t, panggil(t, legal, http.MethodPost, alamat, map[string]string{"agenda_date": "2026-10-05", "agenda": "   "}), http.StatusBadRequest, "agenda kosong")
		harapKode(t, panggil(t, legal, http.MethodPost, perkaraLegal+"/"+uuid.NewString()+"/chronology", map[string]string{"agenda_date": "2026-10-05", "agenda": "x"}), http.StatusBadRequest, "kronologi untuk perkara yang tidak ada")

		harapKode(t, panggil(t, legal, http.MethodPut, alamat+"/"+k.ID, map[string]string{"agenda_date": "2026-10-06", "agenda": "Sidang ditunda"}), http.StatusOK, "mengubah kronologi")

		// LEGAL_AU dan EXTERNAL ikut mengelola kronologi lewat alamat masing-masing.
		harapKode(t, panggil(t, legalAU, http.MethodPost, perkaraLegalAU+"/"+p.ID+"/chronology", map[string]string{"agenda_date": "2026-10-07", "agenda": "Mediasi"}), http.StatusCreated, "LEGAL_AU menambah kronologi")
		harapKode(t, panggil(t, external, http.MethodPost, perkaraExternal+"/"+p.ID+"/chronology", map[string]string{"agenda_date": "2026-10-08", "agenda": "Kesimpulan"}), http.StatusCreated, "EXTERNAL menambah kronologi")

		w = panggil(t, legal, http.MethodGet, alamat, nil)
		harapKode(t, w, http.StatusOK, "daftar kronologi")
		var daftar []kronologi
		isiData(t, w, &daftar)
		agenda := map[string]bool{}
		for _, x := range daftar {
			agenda[x.Agenda] = true
		}
		if len(daftar) != 3 || !agenda["Sidang ditunda"] || !agenda["Mediasi"] || !agenda["Kesimpulan"] {
			t.Errorf("kronologi setelah ditambah dan diubah: %+v", daftar)
		}

		// Kronologi tidak bisa diubah atau dihapus lewat perkara lain.
		lain := buatPerkara(t, legal)
		harapKode(t, panggil(t, legal, http.MethodPut, perkaraLegal+"/"+lain.ID+"/chronology/"+k.ID, map[string]string{"agenda_date": "2026-10-06", "agenda": "x"}), http.StatusBadRequest, "ubah kronologi lewat perkara lain")
		harapKode(t, panggil(t, legal, http.MethodDelete, perkaraLegal+"/"+lain.ID+"/chronology/"+k.ID, nil), http.StatusBadRequest, "hapus kronologi lewat perkara lain")

		harapKode(t, panggil(t, legal, http.MethodDelete, alamat+"/"+k.ID, nil), http.StatusOK, "menghapus kronologi")
		if n := len(ambilPerkara(t, legal, perkaraLegal, p.ID).Chronologies); n != 2 {
			t.Errorf("jumlah kronologi setelah satu dihapus: %d", n)
		}
	})

	t.Run("ImporKronologi", func(t *testing.T) {
		q := buatPerkara(t, legal)
		alamat := perkaraLegal + "/" + q.ID + "/chronology/import"

		// Template yang diunduh bisa langsung diimpor: tiga baris contohnya masuk.
		w := panggil(t, legal, http.MethodGet, perkaraLegal+"/chronology/template", nil)
		harapKode(t, w, http.StatusOK, "unduh template kronologi")
		w = panggilForm(t, legal, http.MethodPost, alamat, nil, berkas{"file", "template.xlsx", w.Body.Bytes()})
		harapKode(t, w, http.StatusOK, "impor template kronologi")
		var hasil hasilImpor
		isiData(t, w, &hasil)
		if hasil.Imported != 3 || len(hasil.Errors) != 0 {
			t.Errorf("impor template: %+v", hasil)
		}

		// Baris yang salah dilaporkan beserta nomor barisnya, baris yang benar tetap masuk.
		file := xlsx(t,
			[]string{"agenda_date", "agenda", "description"},
			[]string{"2026-11-01", "Sidang lanjutan", ""},
			[]string{"2026-11-02", "", "agenda kosong"},
			[]string{"bukan tanggal", "Putusan", ""},
		)
		w = panggilForm(t, legal, http.MethodPost, alamat, nil, berkas{"file", "kronologi.xlsx", file})
		harapKode(t, w, http.StatusOK, "impor kronologi campuran")
		hasil = hasilImpor{}
		isiData(t, w, &hasil)
		if hasil.Imported != 1 || len(hasil.Errors) != 2 || hasil.Errors[0].Row != 3 || hasil.Errors[1].Row != 4 {
			t.Errorf("impor campuran: %+v", hasil)
		}
		if n := len(ambilPerkara(t, legal, perkaraLegal, q.ID).Chronologies); n != 4 {
			t.Errorf("jumlah kronologi setelah dua kali impor: %d", n)
		}

		harapKode(t, panggilForm(t, legal, http.MethodPost, alamat, nil), http.StatusBadRequest, "impor tanpa file")
		harapKode(t, panggilForm(t, legal, http.MethodPost, alamat, nil, berkas{"file", "rusak.xlsx", []byte("bukan excel")}), http.StatusBadRequest, "impor file rusak")
	})

	t.Run("PDF", func(t *testing.T) {
		for _, k := range []struct{ token, alamat string }{{legal, perkaraLegal}, {legalAU, perkaraLegalAU}} {
			w := panggil(t, k.token, http.MethodGet, k.alamat+"/"+p.ID+"/pdf", nil)
			harapKode(t, w, http.StatusOK, "membuat PDF perkara")
			if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
				t.Errorf("jawaban %s bukan PDF", k.alamat)
			}
			harapNamaUnduhan(t, w, p.TicketNumber+"_Perkara.pdf", "PDF perkara "+k.alamat)
		}
	})

	t.Run("Hapus", func(t *testing.T) {
		q := buatPerkara(t, legal)
		harapKode(t, panggil(t, legal, http.MethodDelete, perkaraLegal+"/"+q.ID, nil), http.StatusOK, "menghapus perkara")
		harapKode(t, panggil(t, legal, http.MethodGet, perkaraLegal+"/"+q.ID, nil), http.StatusNotFound, "melihat perkara yang dihapus")
		if perkaraDiDaftar(t, legal, perkaraLegal, q.ID) {
			t.Error("perkara yang dihapus masih muncul di daftar")
		}
		// Nomor tiket perkara yang dihapus tidak boleh dipakai ulang.
		if lagi := buatPerkara(t, legal); lagi.TicketNumber == q.TicketNumber {
			t.Errorf("nomor tiket %s dipakai ulang setelah perkaranya dihapus", lagi.TicketNumber)
		}
	})
}

// Data rujukan perkara: cedant dan kabupaten/kota.
func TestRujukanPerkara(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)

	t.Run("Cedant", func(t *testing.T) {
		alamat := perkaraLegal + "/cedants"
		w := panggil(t, legal, http.MethodPost, alamat, map[string]string{"name": "Cedant Rujukan", "description": "awal"})
		harapKode(t, w, http.StatusCreated, "membuat cedant")
		var c struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		isiData(t, w, &c)

		harapKode(t, panggil(t, legal, http.MethodPost, alamat, map[string]string{"name": ""}), http.StatusBadRequest, "cedant tanpa nama")
		harapKode(t, panggil(t, legal, http.MethodPut, alamat+"/"+c.ID, map[string]string{"name": "Cedant Rujukan Baru"}), http.StatusOK, "mengubah cedant")
		harapKode(t, panggil(t, legal, http.MethodPut, alamat+"/"+uuid.NewString(), map[string]string{"name": "x"}), http.StatusBadRequest, "mengubah cedant yang tidak ada")

		w = panggil(t, legal, http.MethodGet, alamat+"?search="+url.QueryEscape("Rujukan Baru"), nil)
		harapKode(t, w, http.StatusOK, "mencari cedant")
		var daftar []struct {
			ID string `json:"id"`
		}
		isiData(t, w, &daftar)
		if len(daftar) != 1 || daftar[0].ID != c.ID {
			t.Errorf("pencarian cedant: %+v", daftar)
		}

		// Impor: nama yang sudah ada dan nama kosong dilaporkan, sisanya masuk.
		csv := "name,description\nCedant Impor A,satu\nCedant Rujukan Baru,sudah ada\n,tanpa nama\n"
		harapSatuAuditBaru(t, "cedant", entity.ActionFileUpload, uuid.Nil.String(), "impor cedant", func() {
			w = panggilForm(t, legal, http.MethodPost, alamat+"/import", nil, berkas{"file", "cedant.csv", []byte(csv)})
		})
		harapKode(t, w, http.StatusOK, "impor cedant")
		var hasil hasilImpor
		isiData(t, w, &hasil)
		if hasil.Imported != 1 || len(hasil.Errors) != 2 {
			t.Errorf("impor cedant: %+v", hasil)
		}
		w = panggil(t, legal, http.MethodGet, alamat+"/import/template", nil)
		harapKode(t, w, http.StatusOK, "unduh template cedant")
		if !bytes.HasPrefix(w.Body.Bytes(), []byte("PK")) {
			t.Error("template cedant bukan file Excel")
		}

		harapKode(t, panggil(t, legal, http.MethodDelete, alamat+"/"+c.ID, nil), http.StatusOK, "menghapus cedant")
		harapKode(t, panggil(t, legal, http.MethodDelete, alamat+"/"+c.ID, nil), http.StatusBadRequest, "menghapus cedant dua kali")
	})

	t.Run("Kabupaten", func(t *testing.T) {
		w := panggil(t, legal, http.MethodGet, perkaraLegal+"/regencies?limit=5", nil)
		harapKode(t, w, http.StatusOK, "daftar kabupaten")
		var daftar []struct {
			ID string `json:"id"`
		}
		isiData(t, w, &daftar)
		if len(daftar) != 5 {
			t.Errorf("daftar kabupaten dengan limit 5: %d baris", len(daftar))
		}

		const alamat = "/api/v1/admin/regencies"
		baru := map[string]string{"name": "Kota Tes", "province": "Provinsi Tes", "type": "Kota"}
		w = panggil(t, admin, http.MethodPost, alamat, baru)
		harapKode(t, w, http.StatusCreated, "membuat kabupaten")
		var r struct {
			ID string `json:"id"`
		}
		isiData(t, w, &r)
		harapKode(t, panggil(t, admin, http.MethodPost, alamat, baru), http.StatusBadRequest, "kabupaten kembar di provinsi yang sama")
		harapKode(t, panggil(t, admin, http.MethodPost, alamat, map[string]string{"name": "Kota Tes"}), http.StatusBadRequest, "kabupaten tanpa provinsi")

		baru["name"] = "Kota Tes Baru"
		harapKode(t, panggil(t, admin, http.MethodPut, alamat+"/"+r.ID, baru), http.StatusOK, "mengubah kabupaten")
		harapKode(t, panggil(t, admin, http.MethodDelete, alamat+"/"+r.ID, nil), http.StatusOK, "menghapus kabupaten")

		// Kabupaten yang dipakai sebuah perkara tidak boleh dihapus.
		form := formPerkara(t)
		harapKode(t, panggil(t, legal, http.MethodPost, perkaraLegal, form), http.StatusCreated, "membuat perkara")
		harapKode(t, panggil(t, admin, http.MethodDelete, alamat+"/"+form["location_regency_id"].(string), nil), http.StatusBadRequest, "menghapus kabupaten yang dipakai")
	})
}

// Dokumen, foto, dan lampiran kronologi perkara.
func TestBerkasPerkara(t *testing.T) {
	butuhMinIO(t)
	legal := masukPeran(t, entity.RoleLegal)
	p := buatPerkara(t, legal)
	alamat := perkaraLegal + "/" + p.ID
	isiDok, isiFoto, isiLampiran := []byte("dokumen perkara"), []byte("foto perkara"), []byte("lampiran kronologi")

	unduh := func(jalur, path, nama string) *bytes.Buffer {
		t.Helper()
		w := panggil(t, legal, http.MethodGet, perkaraLegal+jalur+"?path="+url.QueryEscape(path), nil)
		harapKode(t, w, http.StatusOK, "mengambil "+path)
		harapNamaUnduhan(t, w, nama, "mengambil "+path)
		return w.Body
	}

	// Dokumen: unggah, unduh, hapus.
	w := panggilForm(t, legal, http.MethodPost, alamat+"/upload-document", nil, berkas{"document", "gugatan.pdf", isiDok})
	harapKode(t, w, http.StatusOK, "unggah dokumen")
	isiData(t, w, &p)
	if !strings.HasPrefix(p.DocumentLink, "legal-cases/documents/") {
		t.Fatalf("alamat dokumen: %q", p.DocumentLink)
	}
	if got := unduh("/download", p.DocumentLink, p.TicketNumber+"_Dokumen_gugatan.pdf"); !bytes.Equal(got.Bytes(), isiDok) {
		t.Errorf("isi dokumen yang diunduh berbeda: %q", got)
	}
	harapKode(t, panggilForm(t, legal, http.MethodPost, alamat+"/upload-document", nil, berkas{"document", "program.exe", []byte("x")}), http.StatusBadRequest, "unggah dokumen .exe")
	harapKode(t, panggilForm(t, legal, http.MethodPost, alamat+"/upload-document", nil), http.StatusBadRequest, "unggah dokumen tanpa file")
	harapKode(t, panggil(t, legal, http.MethodDelete, alamat+"/document", nil), http.StatusOK, "hapus dokumen")
	if got := ambilPerkara(t, legal, perkaraLegal, p.ID).DocumentLink; got != "" {
		t.Errorf("alamat dokumen setelah dihapus: %q", got)
	}

	// Foto: unggah, lihat, hapus.
	w = panggilForm(t, legal, http.MethodPost, alamat+"/photo", nil, berkas{"photo", "sidang.png", isiFoto})
	harapKode(t, w, http.StatusOK, "unggah foto")
	isiData(t, w, &p)
	if got := unduh("/file", p.Photo, p.TicketNumber+"_Foto_sidang.png"); !bytes.Equal(got.Bytes(), isiFoto) {
		t.Errorf("isi foto yang dilihat berbeda: %q", got)
	}
	harapKode(t, panggil(t, legal, http.MethodDelete, alamat+"/photo", nil), http.StatusOK, "hapus foto")
	if got := ambilPerkara(t, legal, perkaraLegal, p.ID).Photo; got != "" {
		t.Errorf("alamat foto setelah dihapus: %q", got)
	}

	// Form perkara menawarkan foto jpg, png, dan webp; ketiganya harus diterima.
	for _, nama := range []string{"sidang.jpg", "sidang.webp"} {
		harapKode(t, panggilForm(t, legal, http.MethodPost, alamat+"/photo", nil, berkas{"photo", nama, isiFoto}), http.StatusOK, "unggah foto "+nama)
	}
	harapKode(t, panggil(t, legal, http.MethodDelete, alamat+"/photo", nil), http.StatusOK, "hapus foto")

	// Kronologi dengan lampiran.
	w = panggilForm(t, legal, http.MethodPost, alamat+"/chronology",
		map[string]string{"agenda_date": "2026-10-05", "agenda": "Sidang bukti"}, berkas{"files", "bukti.pdf", isiLampiran})
	harapKode(t, w, http.StatusCreated, "kronologi dengan lampiran")
	var k kronologi
	isiData(t, w, &k)
	if len(k.Documents) != 1 {
		t.Fatalf("lampiran kronologi: %+v", k.Documents)
	}
	if got := unduh("/download", k.Documents[0], p.TicketNumber+"_Kronologi_2026-10-05.pdf"); !bytes.Equal(got.Bytes(), isiLampiran) {
		t.Errorf("isi lampiran kronologi berbeda: %q", got)
	}

	// Alamat file di luar folder perkara ditolak.
	for _, jalur := range []string{"/download", "/file"} {
		w := panggil(t, legal, http.MethodGet, perkaraLegal+jalur+"?path="+url.QueryEscape("legal-opinions/attachments/milik-orang.pdf"), nil)
		harapKode(t, w, http.StatusNotFound, "mengambil file di luar folder perkara lewat "+jalur)
	}
}

// File perkara hanya dilayani kalau path-nya tercatat di sebuah perkara, untuk
// semua pengguna. Sebelumnya pengguna yang tidak dibatasi perusahaan (induk,
// ADMIN, EXTERNAL) hanya diperiksa awalan foldernya, sehingga path berisi "../"
// lolos dan hanya tertahan karena MinIO tidak menerjemahkannya.
func TestFilePerkaraHarusTerdaftar(t *testing.T) {
	butuhMinIO(t)
	legal := masukPeran(t, entity.RoleLegal)
	isi := []byte("dokumen perkara terdaftar")
	p := buatPerkara(t, legal)
	w := panggilForm(t, legal, http.MethodPost, perkaraLegal+"/"+p.ID+"/upload-document", nil, berkas{"document", "putusan.pdf", isi})
	harapKode(t, w, http.StatusOK, "mengunggah dokumen perkara")
	isiData(t, w, &p)
	lampiranLO := semuaModulPengajuan[0].buat(t, masukPeran(t, entity.RoleUser), berkas{"attachments", "rahasia.pdf", []byte("rahasia")}).Attachments[0].FilePath

	for _, k := range []struct{ nama, token, alamat string }{
		{"LEGAL", legal, perkaraLegal},
		{"ADMIN", masukPeran(t, entity.RoleAdmin), perkaraAdmin},
		{"EXTERNAL", masukPeran(t, entity.RoleExternal), perkaraExternal},
	} {
		for _, jalur := range []string{"/download", "/file"} {
			ambil := func(path string) (int, []byte) {
				t.Helper()
				w := panggil(t, k.token, http.MethodGet, k.alamat+jalur+"?path="+url.QueryEscape(path), nil)
				return w.Code, w.Body.Bytes()
			}
			if kode, got := ambil(p.DocumentLink); kode != http.StatusOK || !bytes.Equal(got, isi) {
				t.Errorf("%s %s dokumen perkara yang sah: dapat %d %q", k.nama, jalur, kode, got)
			}
			for _, path := range []string{
				"legal-cases/documents/../../" + lampiranLO,
				"legal-cases/documents/tidak-terdaftar.pdf",
				lampiranLO,
			} {
				if kode, _ := ambil(path); kode != http.StatusNotFound {
					t.Errorf("%s %s %q: dapat %d, harap 404", k.nama, jalur, path, kode)
				}
			}
		}
	}
}

// Perkara dibatasi per perusahaan untuk user dari perusahaan non-internal:
// mereka hanya melihat dan mengelola perkara perusahaannya sendiri.
func TestPerkaraPerPerusahaan(t *testing.T) {
	// Memakai perusahaan non-internal buatan sendiri, supaya tes ini tidak
	// bergantung pada perusahaan bawaan mana yang bertanda internal. Aturan untuk
	// perusahaan bawaan diuji di TestAnakUsahaBawaanDibatasi.
	w := panggil(t, masukPeran(t, entity.RoleAdmin), http.MethodPost, "/api/v1/admin/companies",
		map[string]any{"name": "PT Luar Tes", "email_domain": "luar-tes.example", "is_internal": false})
	harapKode(t, w, http.StatusCreated, "membuat perusahaan non-internal")
	var luar struct {
		ID         uuid.UUID `json:"id"`
		IsInternal bool      `json:"is_internal"`
	}
	isiData(t, w, &luar)
	if luar.IsInternal {
		t.Fatal("perusahaan dibuat dengan is_internal=false, tetapi tersimpan sebagai internal")
	}
	const surel = "tes-legal-au-luar@example.com"
	if err := tambahUser(surel, entity.RoleLegalAU, luar.ID); err != nil {
		t.Fatal(err)
	}
	orangLuar := masuk(t, surel)

	// Perkara ini milik perusahaan internal (perusahaan si pembuat).
	p := buatPerkara(t, masukPeran(t, entity.RoleLegal))

	if perkaraDiDaftar(t, orangLuar, perkaraLegalAU, p.ID) {
		t.Error("perkara perusahaan lain muncul di daftar")
	}
	w = panggil(t, orangLuar, http.MethodPatch, perkaraLegalAU+"/"+p.ID+"/status", map[string]string{"current_status": "Diubah orang luar"})
	if w.Code == http.StatusOK {
		t.Errorf("user perusahaan lain berhasil mengubah status perkara")
	}

	t.Run("DetailDitolak", func(t *testing.T) {
		harapKode(t, panggil(t, orangLuar, http.MethodGet, perkaraLegalAU+"/"+p.ID, nil), http.StatusNotFound, "user perusahaan lain melihat detail perkara")
	})
	t.Run("PDFDitolak", func(t *testing.T) {
		w := panggil(t, orangLuar, http.MethodGet, perkaraLegalAU+"/"+p.ID+"/pdf", nil)
		if w.Code == http.StatusOK {
			t.Fatalf("user perusahaan lain berhasil mengunduh PDF perkara")
		}
	})
	t.Run("KronologiDitolak", func(t *testing.T) {
		w := panggil(t, orangLuar, http.MethodPost, perkaraLegalAU+"/"+p.ID+"/chronology", map[string]string{"agenda_date": "2026-10-09", "agenda": "Disisipkan orang luar"})
		if w.Code == http.StatusCreated {
			t.Fatalf("user perusahaan lain berhasil menambah kronologi")
		}
	})
}

// Perusahaan pemilik perkara mengikuti pilihan di form untuk user induk, saat
// membuat maupun mengubah. User dari perusahaan non-internal selalu memakai
// perusahaannya sendiri. Aturannya ada di docs/case_company_access.md.
func TestPerusahaanPemilikPerkara(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	induk := masukPeran(t, entity.RoleLegal) // perusahaannya internal

	buatPerusahaan := func(nama, domain string) string {
		t.Helper()
		w := panggil(t, admin, http.MethodPost, "/api/v1/admin/companies", map[string]any{"name": nama, "email_domain": domain, "is_internal": false})
		harapKode(t, w, http.StatusCreated, "membuat perusahaan "+nama)
		var p struct {
			ID string `json:"id"`
		}
		isiData(t, w, &p)
		return p.ID
	}
	anak := buatPerusahaan("PT Anak Pemilik Uji", "anak-pemilik.example")
	lain := buatPerusahaan("PT Anak Lain Uji", "anak-lain.example")

	t.Run("UserInduk", func(t *testing.T) {
		form := formPerkara(t)
		form["company_id"] = anak
		w := panggil(t, induk, http.MethodPost, perkaraLegal, form)
		harapKode(t, w, http.StatusCreated, "induk membuat perkara untuk anak usaha")
		var p perkara
		isiData(t, w, &p)
		if p.CompanyID != anak {
			t.Fatalf("perkara dibuat untuk %s, tersimpan atas %s", anak, p.CompanyID)
		}

		// Memindahkan perkara ke perusahaan lain lewat ubah.
		form["company_id"] = lain
		harapKode(t, panggil(t, induk, http.MethodPut, perkaraLegal+"/"+p.ID, form), http.StatusOK, "induk mengubah perusahaan perkara")
		if got := ambilPerkara(t, induk, perkaraLegal, p.ID).CompanyID; got != lain {
			t.Errorf("perusahaan setelah diubah: dapat %s, harap %s", got, lain)
		}

		form["company_id"] = uuid.NewString()
		harapKode(t, panggil(t, induk, http.MethodPost, perkaraLegal, form), http.StatusBadRequest, "membuat perkara untuk perusahaan yang tidak ada")
		harapKode(t, panggil(t, induk, http.MethodPut, perkaraLegal+"/"+p.ID, form), http.StatusBadRequest, "memindahkan perkara ke perusahaan yang tidak ada")
	})

	t.Run("UserAnakUsaha", func(t *testing.T) {
		const surel = "tes-legal-anak@example.com"
		if err := tambahUser(surel, entity.RoleLegal, uuid.MustParse(anak)); err != nil {
			t.Fatal(err)
		}
		orangAnak := masuk(t, surel)

		// Pilihan perusahaan lain di form diabaikan.
		form := formPerkara(t)
		form["company_id"] = lain
		w := panggil(t, orangAnak, http.MethodPost, perkaraLegal, form)
		harapKode(t, w, http.StatusCreated, "user anak usaha membuat perkara")
		var p perkara
		isiData(t, w, &p)
		if p.CompanyID != anak {
			t.Fatalf("perkara user anak usaha tersimpan atas %s, harap perusahaannya sendiri %s", p.CompanyID, anak)
		}

		// Ia juga tidak bisa memindahkannya lewat ubah.
		harapKode(t, panggil(t, orangAnak, http.MethodPut, perkaraLegal+"/"+p.ID, form), http.StatusOK, "user anak usaha mengubah perkaranya")
		if got := ambilPerkara(t, orangAnak, perkaraLegal, p.ID).CompanyID; got != anak {
			t.Errorf("perusahaan setelah diubah user anak usaha: dapat %s, harap tetap %s", got, anak)
		}
	})
}

// Mengubah pilihan rujukan di form edit (pihak terkait, jenis, kategori, PIC,
// kabupaten) harus benar-benar tersimpan, bukan hanya dijawab berhasil.
func TestUbahRujukanPerkara(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	p := buatPerkara(t, legal)

	idKedua := func(tabel string) string {
		t.Helper()
		var id string
		if err := dbTes.Raw("SELECT id FROM " + tabel + " WHERE deleted_at IS NULL ORDER BY created_at OFFSET 1 LIMIT 1").Scan(&id).Error; err != nil || id == "" {
			t.Fatalf("tabel %s butuh minimal dua baris: %v", tabel, err)
		}
		return id
	}
	w := panggil(t, legal, http.MethodPost, perkaraLegal+"/cedants", map[string]string{"name": "Cedant Pengganti Uji"})
	harapKode(t, w, http.StatusCreated, "membuat cedant pengganti")
	var cedant struct {
		ID string `json:"id"`
	}
	isiData(t, w, &cedant)

	form := formPerkara(t)
	baru := map[string]string{
		"related_party_id":    cedant.ID,
		"case_type_id":        idKedua("case_types"),
		"category_id":         idKedua("case_categories"),
		"pic":                 idKedua("divisions"),
		"location_regency_id": idKedua("regencies"),
	}
	for kolom, nilai := range baru {
		if form[kolom] == nilai {
			t.Fatalf("nilai pengganti untuk %s sama dengan nilai awal", kolom)
		}
		form[kolom] = nilai
	}
	harapKode(t, panggil(t, legal, http.MethodPut, perkaraLegal+"/"+p.ID, form), http.StatusOK, "mengubah rujukan perkara")

	var tersimpan map[string]any
	isiData(t, panggil(t, legal, http.MethodGet, perkaraLegal+"/"+p.ID, nil), &tersimpan)
	for kolom, nilai := range baru {
		if tersimpan[kolom] != nilai {
			t.Errorf("%s setelah diubah: dapat %v, harap %s", kolom, tersimpan[kolom], nilai)
		}
	}
}

// Semua operasi perkara dibatasi ke perusahaan penggunanya: user dari
// perusahaan non-internal tidak bisa menyentuh perkara perusahaan lain lewat
// jalur mana pun, dan tetap bisa mengelola perkara perusahaannya sendiri.
func TestOperasiPerkaraPerPerusahaan(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	induk := masukPeran(t, entity.RoleLegal)

	w := panggil(t, admin, http.MethodPost, "/api/v1/admin/companies",
		map[string]any{"name": "PT Operasi Uji", "email_domain": "operasi-uji.example", "is_internal": false})
	harapKode(t, w, http.StatusCreated, "membuat perusahaan non-internal")
	var perusahaan struct {
		ID string `json:"id"`
	}
	isiData(t, w, &perusahaan)
	// Peran LEGAL dipakai supaya semua alamat perkara tersedia; yang membatasi
	// adalah perusahaannya, bukan perannya.
	const surel = "tes-legal-operasi@example.com"
	if err := tambahUser(surel, entity.RoleLegal, uuid.MustParse(perusahaan.ID)); err != nil {
		t.Fatal(err)
	}
	orang := masuk(t, surel)

	milikLain := buatPerkara(t, induk)
	formSendiri := formPerkara(t)
	formSendiri["company_id"] = perusahaan.ID
	w = panggil(t, induk, http.MethodPost, perkaraLegal, formSendiri)
	harapKode(t, w, http.StatusCreated, "induk membuat perkara untuk perusahaan itu")
	var milikSendiri perkara
	isiData(t, w, &milikSendiri)

	kronologiBaru := map[string]string{"agenda_date": "2026-10-05", "agenda": "Sidang"}
	w = panggil(t, induk, http.MethodPost, perkaraLegal+"/"+milikLain.ID+"/chronology", kronologiBaru)
	harapKode(t, w, http.StatusCreated, "induk menambah kronologi")
	var kronLain kronologi
	isiData(t, w, &kronLain)

	t.Run("PerkaraPerusahaanLainDitolak", func(t *testing.T) {
		dasar := perkaraLegal + "/" + milikLain.ID
		kasus := []struct {
			nama, metode, alamat string
			isi                  any
		}{
			{"detail", http.MethodGet, dasar, nil},
			{"PDF", http.MethodGet, dasar + "/pdf", nil},
			{"ubah", http.MethodPut, dasar, formPerkara(t)},
			{"hapus", http.MethodDelete, dasar, nil},
			{"daftar kronologi", http.MethodGet, dasar + "/chronology", nil},
			{"tambah kronologi", http.MethodPost, dasar + "/chronology", kronologiBaru},
			{"ubah kronologi", http.MethodPut, dasar + "/chronology/" + kronLain.ID, kronologiBaru},
			{"hapus kronologi", http.MethodDelete, dasar + "/chronology/" + kronLain.ID, nil},
			{"hapus dokumen", http.MethodDelete, dasar + "/document", nil},
			{"hapus foto", http.MethodDelete, dasar + "/photo", nil},
			// Kronologi perkara lain tidak bisa dijangkau lewat perkara sendiri.
			{"ubah kronologi lewat perkara sendiri", http.MethodPut, perkaraLegal + "/" + milikSendiri.ID + "/chronology/" + kronLain.ID, kronologiBaru},
		}
		for _, k := range kasus {
			if kode := panggil(t, orang, k.metode, k.alamat, k.isi).Code; kode < 400 || kode >= 500 {
				t.Errorf("%s pada perkara perusahaan lain: dapat %d, harap ditolak (4xx)", k.nama, kode)
			}
		}
		file := xlsx(t, []string{"agenda_date", "agenda"}, []string{"2026-11-01", "Disisipkan"})
		if kode := panggilForm(t, orang, http.MethodPost, dasar+"/chronology/import", nil, berkas{"file", "k.xlsx", file}).Code; kode < 400 || kode >= 500 {
			t.Errorf("impor kronologi pada perkara perusahaan lain: dapat %d, harap ditolak (4xx)", kode)
		}

		// Tidak ada yang berubah pada perkara itu.
		setelah := ambilPerkara(t, induk, perkaraLegal, milikLain.ID)
		if setelah.CaseName != milikLain.CaseName || len(setelah.Chronologies) != 1 {
			t.Errorf("perkara perusahaan lain berubah: nama %q, %d kronologi", setelah.CaseName, len(setelah.Chronologies))
		}
		if perkaraDiDaftar(t, orang, perkaraLegal, milikLain.ID) {
			t.Error("perkara perusahaan lain muncul di daftar")
		}
	})

	t.Run("PerkaraSendiriBisaDikelola", func(t *testing.T) {
		dasar := perkaraLegal + "/" + milikSendiri.ID
		ambilPerkara(t, orang, perkaraLegal, milikSendiri.ID)
		if !perkaraDiDaftar(t, orang, perkaraLegal, milikSendiri.ID) {
			t.Error("perkara sendiri tidak muncul di daftar")
		}
		formSendiri["case_name"] = "Diubah pemiliknya"
		harapKode(t, panggil(t, orang, http.MethodPut, dasar, formSendiri), http.StatusOK, "mengubah perkara sendiri")
		harapKode(t, panggil(t, orang, http.MethodPost, dasar+"/chronology", kronologiBaru), http.StatusCreated, "menambah kronologi perkara sendiri")
		harapKode(t, panggil(t, orang, http.MethodGet, dasar+"/pdf", nil), http.StatusOK, "PDF perkara sendiri")
		// Induk tetap menjangkau perkara perusahaan itu.
		if got := ambilPerkara(t, induk, perkaraLegal, milikSendiri.ID).CaseName; got != "Diubah pemiliknya" {
			t.Errorf("nama perkara dilihat induk: %q", got)
		}
	})

	t.Run("FilePerkara", func(t *testing.T) {
		butuhMinIO(t)
		isi := []byte("dokumen perkara perusahaan lain")
		w := panggilForm(t, induk, http.MethodPost, perkaraLegal+"/"+milikLain.ID+"/upload-document", nil, berkas{"document", "rahasia.pdf", isi})
		harapKode(t, w, http.StatusOK, "induk mengunggah dokumen")
		var p perkara
		isiData(t, w, &p)

		// Unggah ke perkara perusahaan lain ditolak, begitu juga mengambil filenya lewat path.
		for alamat, kolom := range map[string]string{"/upload-document": "document", "/photo": "photo"} {
			if kode := panggilForm(t, orang, http.MethodPost, perkaraLegal+"/"+milikLain.ID+alamat, nil, berkas{kolom, "x.png", []byte("x")}).Code; kode < 400 || kode >= 500 {
				t.Errorf("unggah lewat %s ke perkara perusahaan lain: dapat %d, harap ditolak (4xx)", alamat, kode)
			}
		}
		for _, jalur := range []string{"/download", "/file"} {
			alamat := perkaraLegal + jalur + "?path=" + url.QueryEscape(p.DocumentLink)
			if w := panggil(t, orang, http.MethodGet, alamat, nil); w.Code == http.StatusOK {
				t.Errorf("user perusahaan lain mengambil file perkara lewat %s: %q", jalur, w.Body)
			}
			if w := panggil(t, induk, http.MethodGet, alamat, nil); w.Code != http.StatusOK || !bytes.Equal(w.Body.Bytes(), isi) {
				t.Errorf("induk mengambil file perkara lewat %s: status %d", jalur, w.Code)
			}
		}

		// File perkara sendiri tetap bisa diunggah dan diambil.
		isiSendiri := []byte("dokumen perkara sendiri")
		w = panggilForm(t, orang, http.MethodPost, perkaraLegal+"/"+milikSendiri.ID+"/upload-document", nil, berkas{"document", "milik.pdf", isiSendiri})
		harapKode(t, w, http.StatusOK, "mengunggah dokumen perkara sendiri")
		isiData(t, w, &p)
		w = panggil(t, orang, http.MethodGet, perkaraLegal+"/download?path="+url.QueryEscape(p.DocumentLink), nil)
		harapKode(t, w, http.StatusOK, "mengunduh dokumen perkara sendiri")
		if !bytes.Equal(w.Body.Bytes(), isiSendiri) {
			t.Errorf("isi dokumen perkara sendiri berbeda: %q", w.Body)
		}
		// Lampiran kronologi perkara sendiri, yang disimpan di kolom berbeda.
		w = panggilForm(t, orang, http.MethodPost, perkaraLegal+"/"+milikSendiri.ID+"/chronology",
			map[string]string{"agenda_date": "2026-10-06", "agenda": "Bukti"}, berkas{"files", "bukti_1.pdf", isiSendiri})
		harapKode(t, w, http.StatusCreated, "kronologi dengan lampiran di perkara sendiri")
		var k kronologi
		isiData(t, w, &k)
		harapKode(t, panggil(t, orang, http.MethodGet, perkaraLegal+"/download?path="+url.QueryEscape(k.Documents[0]), nil), http.StatusOK, "mengunduh lampiran kronologi perkara sendiri")
	})
}

// Kalau data user tidak bisa dibaca, perkara tidak boleh terbuka seolah-olah ia
// tidak dibatasi.
func TestPerkaraUserTidakTerbaca(t *testing.T) {
	const surel = "tes-perkara-terhapus@example.com"
	var companyID string
	dbTes.Raw("SELECT company_id FROM users WHERE email = ?", surelPeran(entity.RoleLegal)).Scan(&companyID)
	if err := tambahUser(surel, entity.RoleLegal, uuid.MustParse(companyID)); err != nil {
		t.Fatal(err)
	}
	token := masuk(t, surel)
	harapKode(t, panggil(t, token, http.MethodGet, perkaraLegal, nil), http.StatusOK, "daftar perkara sebelum user dihapus")

	if err := dbTes.Exec("UPDATE users SET deleted_at = now() WHERE email = ?", surel).Error; err != nil {
		t.Fatal(err)
	}
	if kode := panggil(t, token, http.MethodGet, perkaraLegal, nil).Code; kode == http.StatusOK {
		t.Errorf("daftar perkara tetap terbuka untuk user yang datanya sudah tidak ada (status %d)", kode)
	}
}

// Legal AU membuat perkara untuk perusahaannya dan melampirkan dokumen serta
// foto. Halaman form Legal AU memakai alamat di bawah /legal-au/cases, jadi
// semua yang dipanggil form itu harus tersedia di sana.
func TestLegalAUMengelolaPerkara(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)

	w := panggil(t, admin, http.MethodPost, "/api/v1/admin/companies",
		map[string]any{"name": "PT Anak Legal AU Uji", "email_domain": "anak-legal-au.example", "is_internal": false})
	harapKode(t, w, http.StatusCreated, "membuat perusahaan anak usaha")
	var anak struct {
		ID string `json:"id"`
	}
	isiData(t, w, &anak)
	const surel = "tes-legal-au-anak@example.com"
	if err := tambahUser(surel, entity.RoleLegalAU, uuid.MustParse(anak.ID)); err != nil {
		t.Fatal(err)
	}
	legalAU := masuk(t, surel)

	// Daftar pilihan yang dimuat form.
	cedantTes(t)
	for _, daftar := range []string{"/regencies?limit=5", "/cedants"} {
		w := panggil(t, legalAU, http.MethodGet, perkaraLegalAU+daftar, nil)
		harapKode(t, w, http.StatusOK, "Legal AU memuat "+daftar)
		if len(idDariDaftar(t, w)) == 0 {
			t.Errorf("daftar %s kosong untuk Legal AU", daftar)
		}
	}

	// Perkara tersimpan atas perusahaannya sendiri, apa pun pilihan di form.
	form := formPerkara(t)
	w = panggil(t, legalAU, http.MethodPost, perkaraLegalAU, form)
	harapKode(t, w, http.StatusCreated, "Legal AU membuat perkara")
	var p perkara
	isiData(t, w, &p)
	if p.CompanyID != anak.ID {
		t.Fatalf("perkara Legal AU tersimpan atas %s, harap perusahaannya sendiri %s", p.CompanyID, anak.ID)
	}
	if !perkaraDiDaftar(t, legalAU, perkaraLegalAU, p.ID) {
		t.Error("perkara baru tidak muncul di daftar Legal AU")
	}
	// Induk melihat perkara anak usaha.
	ambilPerkara(t, masukPeran(t, entity.RoleLegal), perkaraLegal, p.ID)

	// Legal AU tidak mendapat izin mengubah data pokok atau menghapus perkara.
	for _, k := range []struct{ nama, metode string }{{"mengubah", http.MethodPut}, {"menghapus", http.MethodDelete}} {
		if kode := panggil(t, legalAU, k.metode, perkaraLegalAU+"/"+p.ID, form).Code; kode < 400 {
			t.Errorf("Legal AU %s perkara: dapat %d, harap ditolak", k.nama, kode)
		}
	}

	t.Run("DokumenDanFoto", func(t *testing.T) {
		butuhMinIO(t)
		isiDok, isiFoto := []byte("dokumen dari Legal AU"), []byte("foto dari Legal AU")
		dasar := perkaraLegalAU + "/" + p.ID

		w := panggilForm(t, legalAU, http.MethodPost, dasar+"/upload-document", nil, berkas{"document", "gugatan.pdf", isiDok})
		harapKode(t, w, http.StatusOK, "Legal AU mengunggah dokumen")
		isiData(t, w, &p)
		w = panggil(t, legalAU, http.MethodGet, perkaraLegalAU+"/download?path="+url.QueryEscape(p.DocumentLink), nil)
		harapKode(t, w, http.StatusOK, "Legal AU mengunduh dokumen")
		if !bytes.Equal(w.Body.Bytes(), isiDok) {
			t.Errorf("isi dokumen yang diunduh Legal AU berbeda: %q", w.Body)
		}

		w = panggilForm(t, legalAU, http.MethodPost, dasar+"/photo", nil, berkas{"photo", "sidang.png", isiFoto})
		harapKode(t, w, http.StatusOK, "Legal AU mengunggah foto")
		isiData(t, w, &p)
		w = panggil(t, legalAU, http.MethodGet, perkaraLegalAU+"/file?path="+url.QueryEscape(p.Photo), nil)
		harapKode(t, w, http.StatusOK, "Legal AU melihat foto")
		if !bytes.Equal(w.Body.Bytes(), isiFoto) {
			t.Errorf("isi foto yang dilihat Legal AU berbeda: %q", w.Body)
		}

		harapKode(t, panggil(t, legalAU, http.MethodDelete, dasar+"/document", nil), http.StatusOK, "Legal AU menghapus dokumen")
		harapKode(t, panggil(t, legalAU, http.MethodDelete, dasar+"/photo", nil), http.StatusOK, "Legal AU menghapus foto")
		if got := ambilPerkara(t, legalAU, perkaraLegalAU, p.ID); got.DocumentLink != "" || got.Photo != "" {
			t.Errorf("setelah dihapus: dokumen %q, foto %q", got.DocumentLink, got.Photo)
		}
	})
}

// Aturan utama di docs/case_company_access.md, dengan perusahaan bawaan: Legal
// AU sebuah anak usaha hanya melihat perkara perusahaannya, sedangkan induk dan
// ADMIN melihat perkara semua perusahaan.
func TestAnakUsahaBawaanDibatasi(t *testing.T) {
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
	auAsei := pengguna("tes-au-asei@example.com", entity.RoleLegalAU, asei)
	auSyariah := pengguna("tes-au-syariah@example.com", entity.RoleLegalAU, syariah)
	// ADMIN mengelola seluruh sistem, jadi tidak dibatasi walau terdaftar di anak usaha.
	adminAsei := pengguna("tes-admin-asei@example.com", entity.RoleAdmin, asei)
	induk := masukPeran(t, entity.RoleLegal)

	buatUntuk := func(perusahaan string) perkara {
		t.Helper()
		form := formPerkara(t)
		form["company_id"] = perusahaan
		w := panggil(t, induk, http.MethodPost, perkaraLegal, form)
		harapKode(t, w, http.StatusCreated, "induk membuat perkara untuk anak usaha")
		var p perkara
		isiData(t, w, &p)
		if p.CompanyID != perusahaan {
			t.Fatalf("perkara tersimpan atas %s, harap %s; apakah user LEGAL tes bukan dari induk?", p.CompanyID, perusahaan)
		}
		return p
	}
	milikAsei, milikSyariah := buatUntuk(asei), buatUntuk(syariah)

	kasus := []struct {
		nama, token, alamat string
		lihatAsei, lihatSya bool
	}{
		{"Legal AU ASEI", auAsei, perkaraLegalAU, true, false},
		{"Legal AU Reasuransi Syariah", auSyariah, perkaraLegalAU, false, true},
		{"LEGAL induk", induk, perkaraLegal, true, true},
		{"ADMIN yang terdaftar di ASEI", adminAsei, perkaraAdmin, true, true},
	}
	for _, k := range kasus {
		for _, p := range []struct {
			milik string
			id    string
			harap bool
		}{{"ASEI", milikAsei.ID, k.lihatAsei}, {"Reasuransi Syariah", milikSyariah.ID, k.lihatSya}} {
			if got := perkaraDiDaftar(t, k.token, k.alamat, p.id); got != p.harap {
				t.Errorf("%s, perkara %s di daftar: dapat %v, harap %v", k.nama, p.milik, got, p.harap)
			}
			kode := panggil(t, k.token, http.MethodGet, k.alamat+"/"+p.id, nil).Code
			if (kode == http.StatusOK) != p.harap {
				t.Errorf("%s, detail perkara %s: dapat %d, harap terlihat=%v", k.nama, p.milik, kode, p.harap)
			}
		}
	}
}
