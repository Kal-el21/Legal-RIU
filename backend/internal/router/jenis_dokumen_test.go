package router

// Jenis dokumen perjanjian dan definisi field-nya (PR 9). Field disimpan di
// setiap versi template, dan pengajuan membaca field dari versi template yang
// dikuncinya, bukan dari kode.

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"

	"github.com/google/uuid"
)

// docxSederhana membentuk dokumen Word dengan satu paragraf per teks; teks kosong
// menjadi paragraf kosong (Enter).
func docxSederhana(t *testing.T, paragraf ...string) []byte {
	t.Helper()
	isi := ""
	for _, teks := range paragraf {
		isi += `<w:p><w:r><w:t>` + teks + `</w:t></w:r></w:p>`
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range [][2]string{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`},
		{"word/_rels/document.xml.rels", `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`},
		{"word/document.xml", `<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:body>` + isi + `<w:sectPr/></w:body></w:document>`},
	} {
		dst, err := w.Create(f[0])
		if err != nil {
			t.Fatal(err)
		}
		dst.Write([]byte(f[1]))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

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
// punya field (versi sebelum PR 9, termasuk yang masih dikunci pengajuan) mendapat
// definisi field PR 9, karena versi itu memakai template PR 8 dengan termin 1 dan 2.
// Field kerja dimulai dari field template bawaan. Dijalankan dua kali, seperti
// server yang dinyalakan ulang.
func TestMigrasiJenisDokumen(t *testing.T) {
	// Versi lama tanpa field, seperti template sebelum PR 9.
	lama := uuid.NewString()
	eksekusi(t, `INSERT INTO agreement_templates
		(id, created_at, updated_at, code, version, name, file_name, file_path, checksum, status, is_legacy, note)
		VALUES (?, now(), now(), 'PKS', (SELECT COALESCE(MAX(version), 0) + 1 FROM agreement_templates WHERE code = 'PKS'),
			'uji', 'uji.docx', 'uji/tidak-ada.docx', 'sebelum-pr9-uji', 'ARCHIVED', false, 'uji')`, lama)
	t.Cleanup(func() { dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", lama) })
	eksekusi(t, "DELETE FROM agreement_types WHERE code = 'PKS'")

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
	// Field kerja Admin dimulai dari field template bawaan (grup termin), bukan PR 9.
	if kerja := fieldJenis(t, "PKS"); !strings.Contains(kerja, `"type": "group"`) || strings.Contains(kerja, "termin_1_persen") {
		t.Errorf("field kerja jenis PKS bukan field template bawaan:\n%s", kerja)
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
	// Semua isian bawaan wajib; Admin bisa melonggarkannya di Jenis Perjanjian (keputusan F).
	for nama, f := range isi {
		if !f.wajib {
			t.Errorf("isian bawaan %s tidak wajib (keputusan F)", nama)
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

	// Pengajuan dibuat saat versi PR 8 aktif, lalu template bawaan aktif kembali,
	// jadi pengajuan itu kini memakai versi lama yang belum punya field.
	aktif := templateAktif(t)
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED' WHERE id = ?", aktif.ID)
	lama := pasangVersiPR9(t, string(entity.TemplateStatusActive), fieldPR9(t))
	t.Cleanup(func() {
		hapusVersiUji(lama)
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
	})
	p := buatPerjanjian(t, user, formPerjanjianPR9())
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED', fields = NULL WHERE id = ?", lama)
	eksekusi(t, "UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
	if err := seed.SeedAgreementTypes(dbTes); err != nil {
		t.Fatal(err)
	}
	if f := fieldTemplate(t, lama); !strings.Contains(f, "termin_1_persen") {
		t.Fatalf("versi PR 8 tidak mendapat field PR 9: %s", f)
	}

	ambilPerjanjian(t, user, "", p.ID)
	form := formPerjanjianPR9()
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

	// Template jenis ini hanya memuat placeholder field-nya sendiri dan placeholder
	// sistem, seperti template yang disusun Admin.
	jenis, tpl := uuid.NewString(), uuid.NewString()
	path := "agreement-templates/" + kode + "/" + tpl + ".docx"
	if err := storeTes.UploadBytes(context.Background(), path, docxSederhana(t, "Perjanjian kerahasiaan dengan {{PIHAK_KEDUA_NAMA}}, ditandatangani di {{TEMPAT_TTD}} pada {{TANGGAL_TTD_LENGKAP}}."), "application/vnd.openxmlformats-officedocument.wordprocessingml.document"); err != nil {
		t.Fatal(err)
	}
	const field = `[{"title":"Pihak Kedua","fields":[{"name":"pihak_kedua_nama","label":"Nama Perusahaan","type":"text","required":true}]}]`
	eksekusi(t, "INSERT INTO agreement_types (id, created_at, updated_at, code, name, is_active, fields) VALUES (?, now(), now(), ?, 'Perjanjian Kerahasiaan Uji', true, ?)", jenis, kode, field)
	eksekusi(t, `INSERT INTO agreement_templates
		(id, created_at, updated_at, code, version, name, file_name, file_path, checksum, status, is_legacy, note, fields)
		VALUES (?, now(), now(), ?, 1, 'uji', 'nda.docx', ?, 'uji-nda', 'ACTIVE', false, 'uji', ?)`,
		tpl, kode, path, field)
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", tpl)
		dbTes.Exec("DELETE FROM agreement_types WHERE id = ?", jenis)
		storeTes.DeleteFile(context.Background(), path)
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

	// Daftar placeholder untuk Word mengikuti field jenis itu, bukan field PKS.
	w = panggil(t, masukPeran(t, entity.RoleAdmin), http.MethodGet, "/api/v1/admin/agreement-templates/placeholders?code="+kode, nil)
	harapKode(t, w, http.StatusOK, "placeholder jenis tanpa termin")
	if b := w.Body.String(); !strings.Contains(b, "{{PIHAK_KEDUA_NAMA}}") || !strings.Contains(b, "{{DAFTAR_LAMPIRAN}}") || strings.Contains(b, "{{NILAI_KONTRAK}}") {
		t.Errorf("placeholder jenis tanpa termin: %s", b)
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

const alamatJenis = "/api/v1/admin/agreement-types"

type jenisDokumen struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	Fields   []struct {
		Title  string `json:"title"`
		Fields []struct {
			Name     string   `json:"name"`
			Label    string   `json:"label"`
			Type     string   `json:"type"`
			Required bool     `json:"required"`
			Options  []string `json:"options"`
		} `json:"fields"`
	} `json:"fields"`
}

// hapusJenisUji membersihkan jenis dokumen buatan tes, juga kalau tes berhenti di tengah.
func hapusJenisUji(t *testing.T, kode string) {
	t.Cleanup(func() { dbTes.Exec("DELETE FROM agreement_types WHERE code = ?", kode) })
}

func fieldJenis(t *testing.T, kode string) string {
	t.Helper()
	var raw *string
	if err := dbTes.Raw("SELECT fields::text FROM agreement_types WHERE code = ?", kode).Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	if raw == nil {
		return ""
	}
	return *raw
}

// Jenis dokumen dan field-nya hanya dikelola ADMIN (keputusan K3). Field yang
// disimpan di sini adalah field kerja; pengajuan tetap membaca field versi
// template yang dikuncinya.
func TestJenisDokumenHanyaAdmin(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const kode = "UJI-KELOLA"
	hapusJenisUji(t, kode)

	for _, peran := range []entity.UserRole{entity.RoleLegal, entity.RoleUser, entity.RoleLegalAU, entity.RoleExternal} {
		token := masukPeran(t, peran)
		harapKode(t, panggil(t, token, http.MethodGet, alamatJenis, nil), http.StatusForbidden, string(peran)+" melihat jenis dokumen")
		harapKode(t, panggil(t, token, http.MethodPost, alamatJenis, map[string]any{"code": kode, "name": "Uji"}), http.StatusForbidden, string(peran)+" membuat jenis dokumen")
		harapKode(t, panggil(t, token, http.MethodPut, alamatJenis+"/PKS", map[string]any{"name": "Diubah"}), http.StatusForbidden, string(peran)+" mengubah jenis dokumen")
	}

	isi := map[string]any{"code": " uji-kelola ", "name": "Perjanjian Uji", "fields": []map[string]any{
		{"title": "Pihak Kedua", "fields": []map[string]any{
			{"name": "Nama_Mitra", "label": " Nama Mitra ", "type": "text", "required": true},
			{"name": "JENIS_MITRA", "label": "Jenis Mitra", "type": "select", "options": []string{"Badan usaha", "Perorangan"}},
		}},
	}}
	w := panggil(t, admin, http.MethodPost, alamatJenis, isi)
	harapKode(t, w, http.StatusCreated, "ADMIN membuat jenis dokumen")
	var j jenisDokumen
	isiData(t, w, &j)
	if j.Code != kode || !j.IsActive || len(j.Fields) != 1 || j.Fields[0].Fields[0].Name != "nama_mitra" || j.Fields[0].Fields[0].Label != "Nama Mitra" {
		t.Errorf("jenis baru (kode huruf besar, nama field huruf kecil, label dirapikan): %+v", j)
	}

	w = panggil(t, admin, http.MethodGet, alamatJenis, nil)
	harapKode(t, w, http.StatusOK, "ADMIN melihat jenis dokumen")
	var daftar []jenisDokumen
	isiData(t, w, &daftar)
	kodeAda := map[string]bool{}
	for _, d := range daftar {
		kodeAda[d.Code] = true
	}
	if !kodeAda["PKS"] || !kodeAda[kode] {
		t.Errorf("daftar jenis dokumen: %v", kodeAda)
	}

	// Mengubah tanpa mengirim field tidak mengosongkan field yang tersimpan.
	w = panggil(t, admin, http.MethodPut, alamatJenis+"/"+kode, map[string]any{"name": "Perjanjian Uji Diubah", "is_active": false})
	harapKode(t, w, http.StatusOK, "ADMIN menonaktifkan jenis dokumen")
	isiData(t, w, &j)
	if j.Name != "Perjanjian Uji Diubah" || j.IsActive || len(j.Fields) != 1 {
		t.Errorf("jenis setelah diubah tanpa field: %+v", j)
	}
	harapKode(t, panggil(t, admin, http.MethodPut, alamatJenis+"/TIDAK-ADA", map[string]any{"name": "x"}), http.StatusNotFound, "mengubah jenis yang tidak ada")
}

// Field yang tidak bisa dipakai ditolak dengan pesan yang menyebut field-nya.
func TestValidasiField(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const kode = "UJI-VALIDASI"
	hapusJenisUji(t, kode)
	field := func(f ...map[string]any) []map[string]any {
		return []map[string]any{{"title": "Data", "fields": f}}
	}
	teks := func(nama string) map[string]any {
		return map[string]any{"name": nama, "label": "Isian " + nama, "type": "text"}
	}
	uang := map[string]any{"name": "nilai_kontrak", "label": "Nilai Kontrak", "type": "money"}
	grup := func(bagi string, sub ...map[string]any) map[string]any {
		return map[string]any{"name": "termin", "label": "Termin", "type": "group", "split_of": bagi, "fields": sub}
	}
	persen := map[string]any{"name": "persen", "label": "Persen", "type": "percent"}
	nilai := map[string]any{"name": "nilai", "label": "Nilai", "type": "money"}

	for _, c := range []struct {
		nama  string
		isi   map[string]any
		pesan string
	}{
		{"KodeJenisSalahFormat", map[string]any{"code": "uji nda", "name": "Uji"}, "Kode jenis dokumen"},
		{"KodeJenisGanda", map[string]any{"code": "PKS", "name": "Uji"}, "sudah dipakai"},
		{"NamaJenisKosong", map[string]any{"code": kode, "name": " "}, "Nama jenis dokumen wajib diisi"},
		{"NamaBagianKosong", map[string]any{"code": kode, "name": "Uji", "fields": []map[string]any{{"title": "", "fields": []map[string]any{teks("a")}}}}, "Nama bagian"},
		{"KodeFieldSalahFormat", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("nama-mitra"))}, "NAMA-MITRA"},
		{"KodeFieldDiawaliAngka", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("1_nama"))}, "1_NAMA"},
		{"KodeFieldGanda", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("nama"), teks("NAMA"))}, "dipakai dua kali"},
		{"LabelKosong", map[string]any{"code": kode, "name": "Uji", "fields": field(map[string]any{"name": "nama", "label": " ", "type": "text"})}, "Label field NAMA"},
		{"TipeTidakDikenal", map[string]any{"code": kode, "name": "Uji", "fields": field(map[string]any{"name": "warna", "label": "Warna", "type": "color"})}, "Tipe field WARNA"},
		{"PilihanTanpaOpsi", map[string]any{"code": kode, "name": "Uji", "fields": field(map[string]any{"name": "jenis", "label": "Jenis", "type": "select"})}, "pilihan"},
		{"BentrokPlaceholderSistem", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("nomor_pihak_pertama"))}, "{{NOMOR_PIHAK_PERTAMA}}"},
		{"BentrokIsianApprover", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("pihak_pertama_pic"))}, "{{PIHAK_PERTAMA_PIC}}"},
		{"BentrokDaftarRuangLingkup", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("ruang_lingkup_list"))}, "{{RUANG_LINGKUP_LIST}}"},
		{"BentrokTurunan", map[string]any{"code": kode, "name": "Uji", "fields": field(map[string]any{"name": "nilai", "label": "Nilai", "type": "money"}, teks("nilai_terbilang"))}, "{{NILAI_TERBILANG}}"},
		{"GrupTanpaSubField", map[string]any{"code": kode, "name": "Uji", "fields": field(grup(""))}, "minimal satu sub-field"},
		{"SubFieldGanda", map[string]any{"code": kode, "name": "Uji", "fields": field(grup("", teks("syarat"), teks("SYARAT")))}, "TERMIN_SYARAT dipakai dua kali"},
		{"SubFieldSalahFormat", map[string]any{"code": kode, "name": "Uji", "fields": field(grup("", teks("syarat-bayar")))}, "TERMIN_SYARAT-BAYAR"},
		{"SubFieldBersarang", map[string]any{"code": kode, "name": "Uji", "fields": field(grup("", map[string]any{"name": "poin", "label": "Poin", "type": "list"}))}, "tidak boleh bertipe daftar atau grup"},
		{"BagiBukanUang", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("nama"), grup("nama", persen, nilai))}, "NAMA bukan field uang"},
		{"BagiFieldTidakAda", map[string]any{"code": kode, "name": "Uji", "fields": field(grup("nilai_kontrak", persen, nilai))}, "NILAI_KONTRAK bukan field uang"},
		{"BagiTanpaPersen", map[string]any{"code": kode, "name": "Uji", "fields": field(uang, grup("nilai_kontrak", nilai, teks("syarat")))}, "tepat satu sub-field persen dan satu sub-field uang"},
		{"BentrokTurunanGrup", map[string]any{"code": kode, "name": "Uji", "fields": field(grup("", persen), teks("termin_persen_terbilang"))}, "{{TERMIN_PERSEN_TERBILANG}}"},
		{"BentrokJumlahGrup", map[string]any{"code": kode, "name": "Uji", "fields": field(teks("termin_jumlah"), grup("", teks("syarat")))}, "{{TERMIN_JUMLAH}}"},
		{"NilaiUtamaBukanUang", map[string]any{"code": kode, "name": "Uji", "fields": field(map[string]any{"name": "nama", "label": "Nama", "type": "text", "primary_value": true})}, "Nilai utama hanya untuk field uang: NAMA"},
		{"DuaNilaiUtama", map[string]any{"code": kode, "name": "Uji", "fields": field(map[string]any{"name": "nilai_kontrak", "label": "Nilai", "type": "money", "primary_value": true}, map[string]any{"name": "nilai_denda", "label": "Denda", "type": "money", "primary_value": true})}, "Nilai utama hanya boleh satu field: NILAI_KONTRAK dan NILAI_DENDA"},
		{"NilaiUtamaDiSubField", map[string]any{"code": kode, "name": "Uji", "fields": field(grup("", map[string]any{"name": "nilai", "label": "Nilai", "type": "money", "primary_value": true}))}, "Sub-field TERMIN_NILAI tidak bisa menjadi nilai utama"},
	} {
		t.Run(c.nama, func(t *testing.T) {
			w := panggil(t, admin, http.MethodPost, alamatJenis, c.isi)
			harapKode(t, w, http.StatusBadRequest, c.nama)
			if !strings.Contains(w.Body.String(), c.pesan) {
				t.Errorf("pesan harus menyebut %q: %s", c.pesan, w.Body)
			}
		})
	}
	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_types WHERE code = ?", kode); n != 0 {
		t.Errorf("jenis dokumen tersimpan walau ditolak")
	}

	// Mengubah PKS dengan field yang ditolak tidak mengubah field yang tersimpan.
	sebelum := fieldJenis(t, "PKS")
	w := panggil(t, admin, http.MethodPut, alamatJenis+"/PKS", map[string]any{"name": "Perjanjian Kerja Sama", "fields": field(teks("tempat_ttd"))})
	harapKode(t, w, http.StatusBadRequest, "mengubah PKS dengan field isian sistem")
	if sesudah := fieldJenis(t, "PKS"); sesudah != sebelum {
		t.Errorf("field PKS berubah walau ditolak")
	}
}

const alamatKelolaTemplate = "/api/v1/admin/agreement-templates"

func unggahTemplate(t *testing.T, token, kode, catatan string, docx []byte) *httptest.ResponseRecorder {
	t.Helper()
	return panggilForm(t, token, http.MethodPost, alamatKelolaTemplate, map[string]string{"code": kode, "note": catatan}, berkas{"file", "template.docx", docx})
}

// Hasil pencocokan DOCX dengan field (dto.AgreementTemplateValidation).
type hasilPencocokan struct {
	Placeholders        []string `json:"placeholders"`
	UnknownPlaceholders []string `json:"unknown_placeholders"`
	UnusedPlaceholders  []string `json:"unused_placeholders"`
	Warnings            []string `json:"warnings"`
}

type versiBaru struct {
	Template struct {
		ID       string `json:"id"`
		Version  int    `json:"version"`
		Status   string `json:"status"`
		Checksum string `json:"checksum"`
	} `json:"template"`
	Validation hasilPencocokan `json:"validation"`
}

func ditolakKarena(t *testing.T, w *httptest.ResponseRecorder) hasilPencocokan {
	t.Helper()
	var isi struct {
		Errors hasilPencocokan `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &isi); err != nil {
		t.Fatalf("respons penolakan: %v: %s", err, w.Body)
	}
	return isi.Errors
}

// Unggahan dicocokkan dengan field kerja jenis dokumennya, juga untuk jenis baru
// yang belum punya versi: placeholder tanpa field ditolak beserta daftarnya,
// field yang tidak dipakai dan deretan Enter kosong menjadi peringatan. Semua
// penolakan terjadi sebelum konversi PDF, jadi diuji tanpa LibreOffice.
func TestTemplateDicocokkanDenganField(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	user := masukPeran(t, entity.RoleUser)
	const kode = "UJI-RAHASIA"
	hapusJenisUji(t, kode)
	t.Cleanup(func() { dbTes.Exec("DELETE FROM agreement_templates WHERE code = ?", kode) })
	harapKode(t, panggil(t, admin, http.MethodPost, alamatJenis, map[string]any{"code": kode, "name": "Perjanjian Kerahasiaan Uji", "fields": []map[string]any{{"title": "Mitra", "fields": []map[string]any{
		{"name": "nama_mitra", "label": "Nama Mitra", "type": "text", "required": true},
		{"name": "jangka_waktu", "label": "Jangka Waktu", "type": "date"},
		{"name": "catatan", "label": "Catatan", "type": "textarea"},
	}}}}), http.StatusCreated, "membuat jenis baru")

	// Daftar placeholder untuk disalin ke Word sudah ada sebelum versi pertama.
	w := panggil(t, admin, http.MethodGet, alamatKelolaTemplate+"/placeholders?code="+kode, nil)
	harapKode(t, w, http.StatusOK, "placeholder jenis tanpa versi")
	if b := w.Body.String(); !strings.Contains(b, "{{NAMA_MITRA}}") || !strings.Contains(b, "{{JANGKA_WAKTU_HARI}}") || strings.Contains(b, "{{RUANG_LINGKUP_LIST}}") {
		t.Errorf("placeholder jenis baru: %s", b)
	}

	cocok := docxSederhana(t, "Perjanjian dengan {{NAMA_MITRA}} sampai {{JANGKA_WAKTU}} di {{TEMPAT_TTD}}.", "", "", "", "", "", "Penutup.")
	w = unggahTemplate(t, admin, kode, " ", cocok)
	harapKode(t, w, http.StatusBadRequest, "unggah tanpa catatan versi")
	if !strings.Contains(w.Body.String(), "Catatan perubahan wajib diisi") {
		t.Errorf("pesan tanpa catatan versi: %s", w.Body)
	}

	w = unggahTemplate(t, admin, kode, "Versi pertama", docxSederhana(t, "{{NAMA_MITRA}} {{NILAI_DENDA}} {{ALAMAT_MITRA}}", "{{RUANG_LINGKUP_LIST}}"))
	harapKode(t, w, http.StatusBadRequest, "unggah dengan placeholder tanpa field")
	if asing := strings.Join(ditolakKarena(t, w).UnknownPlaceholders, " "); asing != "{{ALAMAT_MITRA}} {{NILAI_DENDA}} {{RUANG_LINGKUP_LIST}}" {
		t.Errorf("placeholder tanpa field: %q; %s", asing, w.Body)
	}
	harapKode(t, unggahTemplate(t, admin, "TIDAK-ADA", "Versi pertama", cocok), http.StatusBadRequest, "unggah untuk jenis yang tidak ada")
	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_templates WHERE code = ?", kode); n != 0 {
		t.Errorf("%d versi tersimpan walau ditolak", n)
	}

	t.Run("Diterima", func(t *testing.T) {
		butuhMinIO(t)
		butuhLibreOffice(t)
		w := unggahTemplate(t, admin, kode, "Versi pertama", cocok)
		harapKode(t, w, http.StatusCreated, "unggah DOCX yang cocok")
		var hasil versiBaru
		isiData(t, w, &hasil)
		if hasil.Template.Version != 1 || hasil.Template.Status != "DRAFT" {
			t.Errorf("versi pertama jenis baru: %+v", hasil.Template)
		}
		if tidak := strings.Join(hasil.Validation.UnusedPlaceholders, " "); tidak != "{{CATATAN}}" {
			t.Errorf("field tidak terpakai: %q", tidak)
		}
		if !strings.Contains(strings.Join(hasil.Validation.Warnings, " "), "5 paragraf kosong") {
			t.Errorf("peringatan deretan Enter kosong: %v", hasil.Validation.Warnings)
		}
		harapKode(t, panggil(t, admin, http.MethodPost, alamatKelolaTemplate+"/"+hasil.Template.ID+"/activate", nil), http.StatusOK, "mengaktifkan versi pertama")
		if id := isiPengajuan(t, user, kode, nil)["template_id"]; id != hasil.Template.ID {
			t.Errorf("form jenis baru memakai template %v, harus %s", id, hasil.Template.ID)
		}
	})
}

// "Simpan sebagai versi baru" menyalin DOCX versi terakhir dan field kerja, dan
// melewati pencocokan yang sama. Setelah versi baru diaktifkan, pengajuan baru
// memakainya, sedangkan pengajuan yang sudah berjalan tetap di versinya.
func TestAktivasiTemplateBaru(t *testing.T) {
	butuhMinIO(t)
	admin := masukPeran(t, entity.RoleAdmin)
	user := masukPeran(t, entity.RoleUser)
	aktif := templateAktif(t)
	asli := fieldJenis(t, "PKS")
	var dibuat []string
	t.Cleanup(func() {
		dbTes.Exec("UPDATE agreement_types SET fields = ?::jsonb WHERE code = 'PKS'", asli)
		dbTes.Exec("UPDATE agreement_templates SET status = 'ARCHIVED' WHERE code = 'PKS' AND status = 'ACTIVE'")
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
		for _, id := range dibuat {
			dbTes.Exec("DELETE FROM agreement_documents WHERE template_id = ?", id)
			dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", id)
		}
	})
	berjalan := buatPerjanjian(t, user, formPerjanjian())

	ubahField := func(ubah func(f map[string]any) bool) []map[string]any {
		var salinan []map[string]any
		json.Unmarshal([]byte(asli), &salinan)
		for _, b := range salinan {
			var sisa []any
			for _, f := range b["fields"].([]any) {
				if ubah(f.(map[string]any)) {
					sisa = append(sisa, f)
				}
			}
			b["fields"] = sisa
		}
		return salinan
	}
	versiBaruDari := func(catatan string) *httptest.ResponseRecorder {
		return panggil(t, admin, http.MethodPost, alamatKelolaTemplate+"/new-version", map[string]string{"code": "PKS", "note": catatan})
	}

	// Field yang masih dipakai DOCX dihapus: versi baru ditolak.
	tanpaBank := ubahField(func(f map[string]any) bool { return f["name"] != "bank" })
	harapKode(t, panggil(t, admin, http.MethodPut, alamatJenis+"/PKS", map[string]any{"name": "Perjanjian Kerja Sama", "fields": tanpaBank}), http.StatusOK, "menghapus field bank")
	w := versiBaruDari("Uji tanpa bank")
	harapKode(t, w, http.StatusBadRequest, "versi baru tanpa field yang dipakai DOCX")
	if asing := strings.Join(ditolakKarena(t, w).UnknownPlaceholders, " "); asing != "{{BANK}}" {
		t.Errorf("placeholder tanpa field: %q", asing)
	}
	harapKode(t, versiBaruDari(""), http.StatusBadRequest, "versi baru tanpa catatan")

	t.Run("VersiBaruAktif", func(t *testing.T) {
		butuhLibreOffice(t)
		labelBaru := ubahField(func(f map[string]any) bool {
			if f["name"] == "pihak_kedua_nama" {
				f["label"] = "Nama Mitra"
			}
			return true
		})
		harapKode(t, panggil(t, admin, http.MethodPut, alamatJenis+"/PKS", map[string]any{"name": "Perjanjian Kerja Sama", "fields": labelBaru}), http.StatusOK, "mengganti label")
		w := versiBaruDari("Uji label baru")
		harapKode(t, w, http.StatusCreated, "simpan sebagai versi baru")
		var hasil versiBaru
		isiData(t, w, &hasil)
		dibuat = append(dibuat, hasil.Template.ID)
		if hasil.Template.Version <= aktif.Version || hasil.Template.Status != "DRAFT" || hasil.Template.Checksum != aktif.Checksum {
			t.Errorf("versi baru (DOCX disalin dari versi %d): %+v", aktif.Version, hasil.Template)
		}
		harapKode(t, panggil(t, admin, http.MethodPost, alamatKelolaTemplate+"/"+hasil.Template.ID+"/activate", nil), http.StatusOK, "mengaktifkan versi baru")

		w = panggil(t, user, http.MethodGet, "/api/v1/agreement-document-types/PKS/schema", nil)
		if !strings.Contains(w.Body.String(), "Nama Mitra") || !strings.Contains(w.Body.String(), hasil.Template.ID) {
			t.Errorf("form PKS setelah versi baru aktif: %s", w.Body)
		}
		if baru := buatPerjanjian(t, user, formPerjanjian()); baru.TemplateID != hasil.Template.ID {
			t.Errorf("pengajuan baru memakai template %s, harus versi baru %s", baru.TemplateID, hasil.Template.ID)
		}
		if id, _ := templatePengajuan(t, berjalan.ID); id != aktif.ID {
			t.Errorf("pengajuan berjalan pindah ke %s; harus tetap di versi %s", id, aktif.ID)
		}
		harapKode(t, panggil(t, user, http.MethodPut, urlPerjanjian("", "/", berjalan.ID), map[string]any{"form_data": formPerjanjian()}), http.StatusOK, "mengubah pengajuan di versi lama")
	})
}

// Versi yang masih dipakai pengajuan yang belum Selesai, dan versi aktif, tidak
// bisa dihapus. Setelah pengajuannya Selesai, versi boleh dihapus.
func TestVersiDipakaiTidakBisaDihapus(t *testing.T) {
	butuhMinIO(t)
	admin := masukPeran(t, entity.RoleAdmin)
	user := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	aktif := templateAktif(t)
	usang := uuid.NewString()
	salinTemplate(t, aktif.ID, usang, "dipakai-uji", string(entity.TemplateStatusArchived), false, nil)
	t.Cleanup(func() { dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", usang) })
	p := buatPerjanjian(t, user, formPerjanjian())
	eksekusi(t, "UPDATE agreement_documents SET template_id = ? WHERE id = ?", usang, p.ID)

	// Daftar versi menyebut jumlah pengajuan berjalan per versi (layar Admin, tahap 9.5).
	w := panggil(t, admin, http.MethodGet, alamatKelolaTemplate+"?code=PKS", nil)
	harapKode(t, w, http.StatusOK, "daftar versi PKS")
	var daftar []struct {
		ID               string `json:"id"`
		RunningDocuments int64  `json:"running_documents"`
	}
	isiData(t, w, &daftar)
	dipakai := map[string]int64{}
	for _, v := range daftar {
		dipakai[v.ID] = v.RunningDocuments
	}
	if dipakai[usang] != 1 {
		t.Errorf("versi %s dipakai %d pengajuan berjalan menurut daftar, harus 1", usang, dipakai[usang])
	}

	harapKode(t, panggil(t, legal, http.MethodDelete, "/api/v1/legal/agreement-templates/"+usang, nil), http.StatusForbidden, "LEGAL menghapus versi")
	w = panggil(t, admin, http.MethodDelete, alamatKelolaTemplate+"/"+usang, nil)
	harapKode(t, w, http.StatusBadRequest, "menghapus versi yang dipakai pengajuan berjalan")
	if !strings.Contains(w.Body.String(), "1 pengajuan") {
		t.Errorf("pesan versi dipakai: %s", w.Body)
	}
	w = panggil(t, admin, http.MethodDelete, alamatKelolaTemplate+"/"+aktif.ID, nil)
	harapKode(t, w, http.StatusBadRequest, "menghapus versi aktif")

	eksekusi(t, "UPDATE agreement_documents SET status = ? WHERE id = ?", entity.StatusCompleted, p.ID)
	harapKode(t, panggil(t, admin, http.MethodDelete, alamatKelolaTemplate+"/"+usang, nil), http.StatusOK, "menghapus versi setelah pengajuannya Selesai")
	harapKode(t, panggil(t, admin, http.MethodGet, alamatKelolaTemplate+"/"+usang, nil), http.StatusNotFound, "versi yang dihapus")
	ambilPerjanjian(t, user, "", p.ID)
}

const alamatPratinjauDraf = "/api/v1/agreement-documents/preview-draft"

// Pratinjau sebelum kirim (tahap 9.6a): PDF dibentuk dari isian wizard tanpa
// menyimpan apa pun, dengan validasi yang sama dengan saat Kirim. Untuk revisi,
// pratinjau memakai versi template yang dikunci pengajuan itu.
func TestPratinjauSebelumKirim(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)
	isi := func(form map[string]any) map[string]any {
		b := isiPengajuan(t, user, "PKS", form)
		b["attachment_names"] = []string{"penawaran.pdf"}
		return b
	}

	// Pengajuan yang sedang direvisi, dengan versi template baru sudah aktif.
	p := buatPerjanjian(t, user, formPerjanjian())
	aktif := templateAktif(t)
	baru := uuid.NewString()
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED' WHERE id = ?", aktif.ID)
	salinTemplate(t, aktif.ID, baru, "pratinjau-draf-uji", string(entity.TemplateStatusActive), false, nil)
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", baru)
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
	})
	revisi := func(templateID string) map[string]any {
		return map[string]any{"document_type_code": "PKS", "template_id": templateID, "document_id": p.ID, "form_data": formPerjanjian()}
	}

	dokumen := hitungBaris(t, "SELECT COUNT(*) FROM agreement_documents")
	lampiran := hitungBaris(t, "SELECT COUNT(*) FROM agreement_attachments")

	harapKode(t, panggil(t, legal, http.MethodPost, alamatPratinjauDraf, isi(formPerjanjian())), http.StatusForbidden, "LEGAL mempratinjau draf")
	tanpaNama := formPerjanjian()
	delete(tanpaNama, "pihak_kedua_nama")
	w := panggil(t, user, http.MethodPost, alamatPratinjauDraf, isi(tanpaNama))
	harapKode(t, w, http.StatusBadRequest, "pratinjau dengan isian wajib kosong")
	if !strings.Contains(w.Body.String(), "wajib diisi") {
		t.Errorf("pesan isian wajib kosong: %s", w.Body)
	}
	usang := isi(formPerjanjian())
	usang["template_id"] = aktif.ID
	harapKode(t, panggil(t, user, http.MethodPost, alamatPratinjauDraf, usang), http.StatusConflict, "pratinjau dengan versi usang")
	usang["template_id"] = ""
	harapKode(t, panggil(t, user, http.MethodPost, alamatPratinjauDraf, usang), http.StatusBadRequest, "pratinjau tanpa versi template")
	harapKode(t, panggil(t, user, http.MethodPost, alamatPratinjauDraf, revisi(baru)), http.StatusConflict, "pratinjau revisi dengan versi selain versi terkuncinya")
	harapKode(t, panggil(t, lain, http.MethodPost, alamatPratinjauDraf, revisi(aktif.ID)), http.StatusNotFound, "pratinjau revisi pengajuan milik orang lain")

	t.Run("PDF", func(t *testing.T) {
		butuhLibreOffice(t)
		w := panggil(t, user, http.MethodPost, alamatPratinjauDraf, isi(formPerjanjian()))
		harapKode(t, w, http.StatusOK, "pratinjau pengajuan baru")
		if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) || w.Header().Get("Content-Type") != "application/pdf" {
			t.Errorf("pratinjau bukan PDF (%s)", w.Header().Get("Content-Type"))
		}
		w = panggil(t, user, http.MethodPost, alamatPratinjauDraf, revisi(aktif.ID))
		harapKode(t, w, http.StatusOK, "pratinjau revisi dengan versi terkuncinya")
	})

	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_documents"); n != dokumen {
		t.Errorf("pengajuan bertambah %d karena pratinjau", n-dokumen)
	}
	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_attachments"); n != lampiran {
		t.Errorf("lampiran bertambah %d karena pratinjau", n-lampiran)
	}
}

// Pratinjau pengajuan yang tidak ada (atau milik orang lain) dijawab 404, bukan 500
// (temuan PR 8 tahap 8.4).
func TestPratinjauPengajuanTidakAda(t *testing.T) {
	user := masukPeran(t, entity.RoleUser)
	harapKode(t, panggil(t, user, http.MethodGet, urlPerjanjian("", "/", uuid.NewString(), "/preview"), nil), http.StatusNotFound, "pratinjau pengajuan yang tidak ada")
}

// Form revisi digambar dari versi template yang dikunci pengajuan, bukan versi
// aktif, karena pengajuan itu juga divalidasi dengan versi terkuncinya.
func TestSkemaPengajuanTerkunci(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	p := buatPerjanjian(t, user, formPerjanjian())

	aktif := templateAktif(t)
	baru := uuid.NewString()
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED' WHERE id = ?", aktif.ID)
	salinTemplate(t, aktif.ID, baru, "skema-terkunci-uji", string(entity.TemplateStatusActive), false, nil)
	eksekusi(t, `UPDATE agreement_templates SET fields = jsonb_set(fields, '{0,title}', '"Informasi Versi Baru"') WHERE id = ?`, baru)
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", baru)
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
	})

	type skema struct {
		Schema struct {
			TemplateID string `json:"template_id"`
			Sections   []struct {
				Title  string `json:"title"`
				Fields []struct {
					Name string `json:"name"`
				} `json:"fields"`
			} `json:"sections"`
		} `json:"schema"`
	}
	baca := func(token, alamat string) skema {
		t.Helper()
		w := panggil(t, token, http.MethodGet, alamat, nil)
		harapKode(t, w, http.StatusOK, alamat)
		var s skema
		isiData(t, w, &s)
		return s
	}
	terkunci := baca(user, urlPerjanjian("", "/", p.ID, "/schema"))
	if terkunci.Schema.TemplateID != p.TemplateID || terkunci.Schema.Sections[0].Title != "Informasi Umum" {
		t.Errorf("skema revisi harus dari versi terkunci %s: template %s, bagian pertama %q", p.TemplateID, terkunci.Schema.TemplateID, terkunci.Schema.Sections[0].Title)
	}
	pertama := terkunci.Schema.Sections[0].Fields
	if len(pertama) == 0 || pertama[len(pertama)-1].Name != "tempat_ttd" {
		t.Errorf("skema revisi tanpa isian sistem tempat_ttd: %+v", pertama)
	}
	if s := baca(user, "/api/v1/agreement-document-types/PKS/schema"); s.Schema.Sections[0].Title != "Informasi Versi Baru" {
		t.Errorf("skema pengajuan baru harus dari versi aktif: %q", s.Schema.Sections[0].Title)
	}
	harapKode(t, panggil(t, lain, http.MethodGet, urlPerjanjian("", "/", p.ID, "/schema"), nil), http.StatusNotFound, "skema pengajuan milik orang lain")

	// Approver membaca label isian dari versi yang sama untuk halaman detail (D4).
	t.Run("Approver", func(t *testing.T) {
		for _, c := range []struct {
			peran  entity.UserRole
			awalan string
		}{{entity.RoleLegal, "/legal"}, {entity.RoleAdmin, "/admin"}} {
			token := masukPeran(t, c.peran)
			if s := baca(token, urlPerjanjian(c.awalan, "/", p.ID, "/schema")); s.Schema.TemplateID != p.TemplateID || s.Schema.Sections[0].Title != "Informasi Umum" {
				t.Errorf("%s: skema harus dari versi terkunci %s: template %s, bagian pertama %q", c.awalan, p.TemplateID, s.Schema.TemplateID, s.Schema.Sections[0].Title)
			}
			harapKode(t, panggil(t, token, http.MethodGet, urlPerjanjian(c.awalan, "/", uuid.NewString(), "/schema"), nil), http.StatusNotFound, c.awalan+" skema pengajuan yang tidak ada")
		}
		harapKode(t, panggil(t, user, http.MethodGet, urlPerjanjian("/legal", "/", p.ID, "/schema"), nil), http.StatusForbidden, "USER lewat alamat approver")
	})
}

type penandaTemplate struct {
	ID          string `json:"id"`
	OldTemplate *struct {
		Version       int    `json:"version"`
		ActiveVersion int    `json:"active_version"`
		ActiveNote    string `json:"active_note"`
		ActivatedAt   string `json:"activated_at"`
		ActivatedBy   string `json:"activated_by"`
	} `json:"old_template"`
}

func daftarPenanda(t *testing.T, token, awalan string) map[string]penandaTemplate {
	t.Helper()
	w := panggil(t, token, http.MethodGet, urlPerjanjian(awalan)+"?limit=100", nil)
	harapKode(t, w, http.StatusOK, "daftar perjanjian "+awalan)
	var isi struct {
		Items []penandaTemplate `json:"items"`
	}
	isiData(t, w, &isi)
	out := map[string]penandaTemplate{}
	for _, p := range isi.Items {
		out[p.ID] = p
	}
	return out
}

// Approver melihat pengajuan berjalan yang masih memakai versi template lama,
// beserta versi aktif dan catatannya, karena persetujuan membentuk dokumen final
// dengan versi lama itu. Pengajuan Selesai sudah final dan pengaju tidak diberi
// penanda.
func TestPenandaTemplateLama(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	berjalan := buatPerjanjian(t, user, formPerjanjian())
	selesai := buatPerjanjian(t, user, formPerjanjian())
	eksekusi(t, "UPDATE agreement_documents SET status = ? WHERE id = ?", entity.StatusCompleted, selesai.ID)

	var admin struct{ ID, FullName string }
	dbTes.Raw("SELECT id, full_name FROM users WHERE role = ? AND deleted_at IS NULL LIMIT 1", entity.RoleAdmin).Scan(&admin)
	aktif := templateAktif(t)
	baru := uuid.NewString()
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED' WHERE id = ?", aktif.ID)
	salinTemplate(t, aktif.ID, baru, "penanda-uji", string(entity.TemplateStatusActive), false, nil)
	eksekusi(t, "UPDATE agreement_templates SET note = 'Klausul denda diperbarui', activated_by = ?, activated_at = now() WHERE id = ?", admin.ID, baru)
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_documents WHERE template_id = ?", baru)
		dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", baru)
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
	})
	pengajuanBaru := buatPerjanjian(t, user, formPerjanjian())
	versiBaru := templateDB(t, baru).Version

	var detail penandaTemplate
	isiData(t, panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", berjalan.ID), nil), &detail)
	if o := detail.OldTemplate; o == nil || o.Version != aktif.Version || o.ActiveVersion != versiBaru || o.ActiveNote != "Klausul denda diperbarui" || o.ActivatedBy != admin.FullName || o.ActivatedAt == "" {
		t.Errorf("penanda template lama di detail approver: %+v (harus v%d, aktif v%d oleh %s)", o, aktif.Version, versiBaru, admin.FullName)
	}

	daftar := daftarPenanda(t, legal, "/legal")
	if daftar[berjalan.ID].OldTemplate == nil {
		t.Error("daftar approver tidak menandai pengajuan berjalan di versi lama")
	}
	if daftar[pengajuanBaru.ID].OldTemplate != nil {
		t.Error("pengajuan di versi aktif ditandai template lama")
	}
	if daftar[selesai.ID].OldTemplate != nil {
		t.Error("pengajuan Selesai ditandai template lama")
	}

	var milik penandaTemplate
	isiData(t, panggil(t, user, http.MethodGet, urlPerjanjian("", "/", berjalan.ID), nil), &milik)
	if milik.OldTemplate != nil || daftarPenanda(t, user, "")[berjalan.ID].OldTemplate != nil {
		t.Error("pengaju diberi penanda template lama")
	}
}

// Tempat tanda tangan tidak wajib bagi pengaju, tetapi dokumen final tidak boleh
// tanpa tempat: persetujuan ditolak kalau isian approver, usulan pengaju, dan
// tempat bawaan master semuanya kosong (keputusan G).
func TestTempatTandaTanganSaatPersetujuan(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	var bawaan string
	dbTes.Raw("SELECT default_signing_place FROM agreement_company_masters WHERE is_active LIMIT 1").Scan(&bawaan)
	eksekusi(t, "UPDATE agreement_company_masters SET default_signing_place = '' WHERE is_active")
	t.Cleanup(func() {
		dbTes.Exec("UPDATE agreement_company_masters SET default_signing_place = ? WHERE is_active", bawaan)
	})

	p := buatPerjanjian(t, pemilik, formPerjanjian())
	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_date": "2026-10-20"}), http.StatusOK, "mengisi tanggal tanda tangan")
	w := panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/status"), map[string]string{"status": "COMPLETED"})
	harapKode(t, w, http.StatusBadRequest, "menyetujui tanpa tempat tanda tangan")
	if !strings.Contains(w.Body.String(), "Tempat penandatanganan wajib diisi") {
		t.Errorf("pesan tanpa tempat tanda tangan: %s", w.Body)
	}

	t.Run("TempatDariApprover", func(t *testing.T) {
		butuhLibreOffice(t)
		harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_place": "Bandung"}), http.StatusOK, "approver mengisi tempat")
		statusPerjanjian(t, legal, p.ID, "COMPLETED", "", http.StatusOK)
	})
}
