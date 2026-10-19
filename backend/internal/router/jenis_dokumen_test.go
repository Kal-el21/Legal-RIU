package router

// Jenis dokumen perjanjian dan definisi field-nya (PR 9). Field disimpan di
// setiap versi template, dan pengajuan membaca field dari versi template yang
// dikuncinya, bukan dari kode.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"

	"github.com/google/uuid"
)

func hitungBaris(t *testing.T, sql string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := dbTes.Raw(sql, args...).Scan(&n).Error; err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return n
}

type bagianField struct {
	Title  string `json:"title"`
	Fields []struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Required bool   `json:"required"`
	} `json:"fields"`
}

func fieldTemplate(t *testing.T, id string) string {
	t.Helper()
	var raw *string
	if err := dbTes.Raw("SELECT fields::text FROM agreement_templates WHERE id = ?", id).Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	if raw == nil {
		return ""
	}
	return *raw
}

// Migrasi PR 9: jenis PKS terdaftar, dan setiap versi template PKS yang belum
// punya field (termasuk versi lama yang masih dikunci pengajuan) mendapat field
// dari form-schema.json. Dijalankan dua kali, seperti server yang dinyalakan ulang.
func TestMigrasiJenisDokumen(t *testing.T) {
	// Versi lama tanpa field, seperti template sebelum PR 9.
	lama := uuid.NewString()
	eksekusi(t, `INSERT INTO agreement_templates
		(id, created_at, updated_at, code, version, name, file_name, file_path, checksum, status, is_legacy, note)
		VALUES (?, now(), now(), 'PKS', (SELECT COALESCE(MAX(version), 0) + 1 FROM agreement_templates WHERE code = 'PKS'),
			'uji', 'uji.docx', 'uji/tidak-ada.docx', 'sebelum-pr9-uji', 'ARCHIVED', false, 'uji')`, lama)
	t.Cleanup(func() { dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", lama) })
	eksekusi(t, "DELETE FROM agreement_types WHERE code = 'PKS'")
	eksekusi(t, "UPDATE agreement_templates SET fields = NULL WHERE code = 'PKS'")

	var jenis [2]string
	var field [2]string
	for i := range 2 {
		if err := seed.SeedAgreementTypes(dbTes); err != nil {
			t.Fatalf("seed jenis dokumen putaran %d: %v", i+1, err)
		}
		dbTes.Raw("SELECT id FROM agreement_types WHERE code = 'PKS'").Scan(&jenis[i])
		field[i] = fieldTemplate(t, lama)
	}
	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_types WHERE code = 'PKS'"); n != 1 || jenis[0] != jenis[1] {
		t.Errorf("jenis PKS terdaftar %d kali (id %s lalu %s), harus sekali", n, jenis[0], jenis[1])
	}
	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_templates WHERE code = 'PKS' AND fields IS NULL"); n != 0 {
		t.Errorf("%d versi template PKS belum punya field", n)
	}
	if field[0] == "" || field[0] != field[1] {
		t.Fatalf("field versi lama setelah putaran 1 dan 2 berbeda atau kosong:\n%s\n%s", field[0], field[1])
	}

	var bagian []bagianField
	if err := json.Unmarshal([]byte(field[0]), &bagian); err != nil {
		t.Fatal(err)
	}
	isi := map[string]struct {
		tipe  string
		wajib bool
	}{}
	for _, b := range bagian {
		for _, f := range b.Fields {
			isi[f.Name] = struct {
				tipe  string
				wajib bool
			}{f.Type, f.Required}
		}
	}
	if len(bagian) != 4 || len(isi) != 27 {
		t.Errorf("field PKS: %d bagian, %d field; harus 4 dan 27", len(bagian), len(isi))
	}
	if _, ada := isi["tempat_ttd"]; ada {
		t.Error("tempat_ttd masih field Admin; seharusnya isian sistem (keputusan B1)")
	}
	for _, nama := range []string{"termin_1_persen", "termin_2_persen"} {
		if isi[nama].tipe != "percent" {
			t.Errorf("%s bertipe %q, harus percent (keputusan B2)", nama, isi[nama].tipe)
		}
	}
	for _, surat := range []string{"penawaran", "penunjukan"} {
		for _, kolom := range []string{"nomor", "perihal", "tanggal"} {
			if nama := "surat_" + surat + "_" + kolom; !isi[nama].wajib {
				t.Errorf("%s tidak wajib (K7)", nama)
			}
		}
	}

	// Kode jenis lama dipindah ke PKS di data, jadi kode tidak lagi menerima alias.
	t.Run("KodeJenisLama", func(t *testing.T) {
		var pemilik string
		dbTes.Raw("SELECT id FROM users WHERE deleted_at IS NULL LIMIT 1").Scan(&pemilik)
		for _, kode := range []string{"PERJANJIAN KERJA SAMA", "PERJANJIAN_KERJA_SAMA", "lain-lain", "LAIN LAIN", "OTHER", ""} {
			id := uuid.NewString()
			eksekusi(t, `INSERT INTO agreement_documents (id, created_at, updated_at, ticket_number, user_id, document_type_code, form_data, status)
				VALUES (?, now(), now(), ?, ?, ?, '{}', 'SUBMITTED')`, id, "UJI-KODE-"+id[:8], pemilik, kode)
			t.Cleanup(func() { dbTes.Exec("DELETE FROM agreement_documents WHERE id = ?", id) })
			if err := seed.BackfillAgreementDocumentDocumentTypeCode(dbTes); err != nil {
				t.Fatal(err)
			}
			var hasil string
			dbTes.Raw("SELECT document_type_code FROM agreement_documents WHERE id = ?", id).Scan(&hasil)
			if hasil != "PKS" {
				t.Errorf("kode lama %q menjadi %q, harus PKS", kode, hasil)
			}
		}
	})
}

// Pengajuan yang dibuat sebelum PR 9 menunjuk versi template tanpa field. Setelah
// migrasi pengajuan itu tetap bisa dibuka, diubah, dan dipratinjau, juga kalau
// versinya sudah bukan versi aktif.
func TestPengajuanLamaTetapJalan(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	aktif := templateAktif(t)
	p := buatPerjanjian(t, user, formPerjanjian())

	// Versi baru diaktifkan, jadi pengajuan itu kini memakai versi lama.
	baru := uuid.NewString()
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED' WHERE id = ?", aktif.ID)
	salinTemplate(t, aktif.ID, baru, "versi-baru-uji", string(entity.TemplateStatusActive), false, nil)
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", baru)
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
	})
	eksekusi(t, "UPDATE agreement_templates SET fields = NULL WHERE code = 'PKS'")
	if err := seed.SeedAgreementTypes(dbTes); err != nil {
		t.Fatal(err)
	}

	ambilPerjanjian(t, user, "", p.ID)
	form := formPerjanjian()
	form["tempat_ttd"] = "Bandung"
	harapKode(t, panggil(t, user, http.MethodPut, urlPerjanjian("", "/", p.ID), map[string]any{"form_data": form}), http.StatusOK, "mengubah pengajuan lama")
	if v := ambilPerjanjian(t, user, "", p.ID).FormData["tempat_ttd"]; v != "Bandung" {
		t.Errorf("usulan tempat tanda tangan setelah diubah: %v", v)
	}

	t.Run("Pratinjau", func(t *testing.T) {
		butuhLibreOffice(t)
		w := panggil(t, user, http.MethodGet, urlPerjanjian("", "/", p.ID, "/preview"), nil)
		harapKode(t, w, http.StatusOK, "pratinjau pengajuan lama")
		if !strings.HasPrefix(w.Body.String(), "%PDF") {
			t.Error("pratinjau bukan PDF")
		}
	})
}

// Pengajuan dikirim bersama versi template yang dipakai form. Versi yang sudah
// bukan versi aktif ditolak 409 sebelum apa pun tersimpan, termasuk lampiran,
// supaya pengaju tidak mengirim isian untuk field yang sudah berubah.
func TestPengajuanVersiUsang(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)

	usang := uuid.NewString()
	salinTemplate(t, templateAktif(t).ID, usang, "versi-usang-uji", string(entity.TemplateStatusArchived), false, nil)
	t.Cleanup(func() { dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", usang) })

	dokumen := hitungBaris(t, "SELECT COUNT(*) FROM agreement_documents")
	lampiran := hitungBaris(t, "SELECT COUNT(*) FROM agreement_attachments")

	isi := isiPengajuan(t, user, "PKS", formPerjanjian())
	isi["template_id"] = usang
	data, _ := json.Marshal(isi)
	w := panggilForm(t, user, http.MethodPost, urlPerjanjian(""), map[string]string{"data": string(data)}, berkas{"attachments", "penawaran.pdf", []byte("penawaran")})
	harapKode(t, w, http.StatusConflict, "mengirim dengan versi template usang")
	if !strings.Contains(w.Body.String(), "Template sudah diperbarui") {
		t.Errorf("pesan versi usang: %s", w.Body)
	}

	isi["template_id"] = ""
	w = panggil(t, user, http.MethodPost, urlPerjanjian(""), isi)
	harapKode(t, w, http.StatusBadRequest, "mengirim tanpa versi template")
	if !strings.Contains(w.Body.String(), "versi template") {
		t.Errorf("pesan tanpa versi template: %s", w.Body)
	}

	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_documents"); n != dokumen {
		t.Errorf("pengajuan bertambah %d walau ditolak", n-dokumen)
	}
	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_attachments"); n != lampiran {
		t.Errorf("lampiran bertambah %d walau ditolak", n-lampiran)
	}
}

// Aturan termin dan jangka waktu hanya berlaku untuk jenis yang punya field-nya.
// Sebelumnya setiap pengajuan diperiksa "jumlah persentase termin harus 100",
// sehingga jenis tanpa termin tidak pernah bisa dikirim.
func TestJenisTanpaTermin(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	const kode = "UJI-NDA"

	jenis, tpl := uuid.NewString(), uuid.NewString()
	eksekusi(t, "INSERT INTO agreement_types (id, created_at, updated_at, code, name, is_active) VALUES (?, now(), now(), ?, 'Perjanjian Kerahasiaan Uji', true)", jenis, kode)
	eksekusi(t, `INSERT INTO agreement_templates
		(id, created_at, updated_at, code, version, name, file_name, file_path, checksum, status, is_legacy, note, fields)
		SELECT ?, now(), now(), ?, 1, 'uji', file_name, file_path, 'uji-nda', 'ACTIVE', false, 'uji', ?
		FROM agreement_templates WHERE id = ?`,
		tpl, kode, `[{"title":"Pihak Kedua","fields":[{"name":"pihak_kedua_nama","label":"Nama Perusahaan","type":"text","required":true}]}]`, templateAktif(t).ID)
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", tpl)
		dbTes.Exec("DELETE FROM agreement_types WHERE id = ?", jenis)
	})

	w := panggil(t, user, http.MethodGet, "/api/v1/agreement-document-types/"+kode+"/schema", nil)
	harapKode(t, w, http.StatusOK, "skema jenis tanpa termin")
	var skema struct {
		Schema struct {
			TemplateID string `json:"template_id"`
			Sections   []struct {
				Fields []struct {
					Name string `json:"name"`
				} `json:"fields"`
			} `json:"sections"`
		} `json:"schema"`
	}
	isiData(t, w, &skema)
	if skema.Schema.TemplateID != tpl {
		t.Errorf("skema menyebut template %q, harus %q", skema.Schema.TemplateID, tpl)
	}
	// Tempat tanda tangan adalah isian sistem: ada di form setiap jenis, di bagian
	// pertama, walau tidak didefinisikan Admin (keputusan B1).
	if b := skema.Schema.Sections; len(b) != 1 || len(b[0].Fields) != 2 || b[0].Fields[1].Name != "tempat_ttd" {
		t.Errorf("isian form jenis tanpa termin: %+v", b)
	}

	harapKode(t, panggil(t, user, http.MethodPost, urlPerjanjian(""), isiPengajuan(t, user, kode, map[string]any{})), http.StatusBadRequest, "field wajib jenis baru kosong")

	w = panggil(t, user, http.MethodPost, urlPerjanjian(""), isiPengajuan(t, user, kode, map[string]any{"pihak_kedua_nama": "PT Rahasia"}))
	harapKode(t, w, http.StatusCreated, "mengirim jenis tanpa termin")
	var p perjanjian
	isiData(t, w, &p)

	t.Run("Persetujuan", func(t *testing.T) {
		butuhLibreOffice(t)
		statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
		harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_date": "2026-10-20"}), http.StatusOK, "mengisi tanggal tanda tangan")
		statusPerjanjian(t, legal, p.ID, "COMPLETED", "", http.StatusOK)
	})
}

// Surat Penawaran dan Surat Penunjukan wajib (K7), juga saat persetujuan:
// pengajuan yang dibuat sebelum aturan ini dan mengosongkannya tidak bisa
// disetujui, dan pesannya mengarahkan approver ke revisi.
func TestSuratWajibSaatPersetujuan(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	tanggal := map[string]string{"signing_date": "2026-10-20"}

	tanpaSurat := formPerjanjian()
	delete(tanpaSurat, "surat_penawaran_nomor")
	w := panggil(t, pemilik, http.MethodPost, urlPerjanjian(""), isiPengajuan(t, pemilik, "PKS", tanpaSurat))
	harapKode(t, w, http.StatusBadRequest, "mengirim tanpa nomor Surat Penawaran")
	if !strings.Contains(w.Body.String(), "Nomor Surat Penawaran wajib diisi") {
		t.Errorf("pesan tanpa Surat Penawaran: %s", w.Body)
	}

	// Pengajuan lama: dibuat lengkap, lalu nomor suratnya dikosongkan langsung.
	p := buatPerjanjian(t, pemilik, formPerjanjian())
	eksekusi(t, "UPDATE agreement_documents SET form_data = form_data - 'surat_penawaran_nomor' WHERE id = ?", p.ID)
	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), tanggal), http.StatusOK, "mengisi tanggal tanda tangan")

	w = panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/status"), map[string]string{"status": "COMPLETED"})
	harapKode(t, w, http.StatusBadRequest, "menyetujui pengajuan tanpa Surat Penawaran")
	if !strings.Contains(w.Body.String(), "Nomor Surat Penawaran wajib diisi") || !strings.Contains(w.Body.String(), "revisi") {
		t.Errorf("pesan persetujuan tanpa Surat Penawaran: %s", w.Body)
	}
	if s := ambilPerjanjian(t, pemilik, "", p.ID).Status; s != "UNDER_REVIEW" {
		t.Fatalf("status setelah persetujuan ditolak: %s", s)
	}

	t.Run("SetelahRevisi", func(t *testing.T) {
		butuhLibreOffice(t)
		statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "Lengkapi Surat Penawaran", http.StatusOK)
		harapKode(t, panggil(t, pemilik, http.MethodPut, urlPerjanjian("", "/", p.ID), map[string]any{"form_data": formPerjanjian()}), http.StatusOK, "merevisi isian")
		harapKode(t, panggilForm(t, pemilik, http.MethodPost, urlPerjanjian("", "/", p.ID, "/resubmit"), nil), http.StatusOK, "mengajukan ulang")
		statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
		harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), tanggal), http.StatusOK, "mengisi ulang tanggal tanda tangan")
		statusPerjanjian(t, legal, p.ID, "COMPLETED", "", http.StatusOK)
	})
}
