package router

// Alur Agreement Document, template perjanjian, dan data perusahaan (master
// pihak pertama).
//
// Membuat dokumen butuh template aktif yang disimpan di MinIO, jadi hampir
// semua tes di sini dilewati di GitHub Actions. Persetujuan akhir dan pratinjau
// mengubah DOCX ke PDF lewat LibreOffice, jadi dilewati juga di mesin yang
// tidak memasangnya.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
)

const alamatPerjanjian = "/agreement-documents"

type perjanjian struct {
	ID              string         `json:"id"`
	TicketNumber    string         `json:"ticket_number"`
	Status          string         `json:"status"`
	ApproverNote    string         `json:"approver_note"`
	AgreementNumber *string        `json:"agreement_number"`
	TemplateID      string         `json:"template_id"`
	FormData        map[string]any `json:"form_data"`
	Attachments     []struct {
		ID          string `json:"id"`
		FileName    string `json:"file_name"`
		UploadRound int    `json:"upload_round"`
	} `json:"attachments"`
}

// formPerjanjian mengembalikan isi form PKS yang sah. Tiap pemanggilan
// menghasilkan map baru supaya tes boleh mengubahnya.
func formPerjanjian() map[string]any {
	return map[string]any{
		"pihak_kedua_nama": "PT Mitra Tes", "pihak_kedua_bidang": "Konsultan", "pihak_kedua_alamat": "Jl. Tes No. 1",
		"pihak_kedua_pejabat": "Siti", "pihak_kedua_jabatan": "Direktur",
		"jenis_pekerjaan": "Jasa konsultasi", "ruang_lingkup": "Menyusun laporan",
		"jangka_waktu_mulai": "2026-01-01", "jangka_waktu_selesai": "2026-12-31",
		"nilai_kontrak": 1000000, "termin_1_persen": 40, "termin_1_nilai": 400000,
		"termin_2_persen": 60, "termin_2_nilai": 600000,
		"bank": "Bank Tes", "nomor_rekening": "1234567890", "atas_nama": "PT Mitra Tes",
	}
}

func urlPerjanjian(awalan string, sisa ...string) string {
	u := "/api/v1" + awalan + alamatPerjanjian
	for _, s := range sisa {
		u += s
	}
	return u
}

func buatPerjanjian(t *testing.T, token string, form map[string]any) perjanjian {
	t.Helper()
	w := panggil(t, token, http.MethodPost, urlPerjanjian(""), map[string]any{"document_type_code": "PKS", "form_data": form})
	harapKode(t, w, http.StatusCreated, "membuat perjanjian")
	var p perjanjian
	isiData(t, w, &p)
	return p
}

func ambilPerjanjian(t *testing.T, token, awalan, id string) perjanjian {
	t.Helper()
	w := panggil(t, token, http.MethodGet, urlPerjanjian(awalan, "/", id), nil)
	harapKode(t, w, http.StatusOK, "mengambil perjanjian")
	var p perjanjian
	isiData(t, w, &p)
	return p
}

func statusPerjanjian(t *testing.T, token, id, status, catatan string, harap int) {
	t.Helper()
	w := panggil(t, token, http.MethodPatch, urlPerjanjian("/legal", "/", id, "/status"), map[string]string{"status": status, "note": catatan})
	harapKode(t, w, harap, "mengubah status perjanjian ke "+status)
}

func perjanjianDiDaftar(t *testing.T, token, awalan, id string) bool {
	t.Helper()
	w := panggil(t, token, http.MethodGet, urlPerjanjian(awalan, "?limit=100"), nil)
	harapKode(t, w, http.StatusOK, "mengambil daftar perjanjian")
	var daftar struct {
		Items []perjanjian `json:"items"`
	}
	isiData(t, w, &daftar)
	for _, p := range daftar.Items {
		if p.ID == id {
			return true
		}
	}
	return false
}

// butuhLibreOffice melewati tes yang mengubah DOCX ke PDF.
func butuhLibreOffice(t *testing.T) {
	t.Helper()
	for _, nama := range []string{"libreoffice", "soffice"} {
		if _, err := exec.LookPath(nama); err == nil {
			return
		}
	}
	t.Skip("butuh LibreOffice untuk mengubah DOCX ke PDF; tidak terpasang di mesin ini")
}

// Jenis perjanjian dan skema formnya, yang dipakai frontend untuk menggambar form.
func TestJenisPerjanjian(t *testing.T) {
	user := masukPeran(t, entity.RoleUser)

	w := panggil(t, user, http.MethodGet, "/api/v1/agreement-document-types", nil)
	harapKode(t, w, http.StatusOK, "daftar jenis perjanjian")
	var jenis []struct {
		Code string `json:"code"`
	}
	isiData(t, w, &jenis)
	if len(jenis) != 1 || jenis[0].Code != "PKS" {
		t.Errorf("jenis perjanjian: %+v", jenis)
	}

	w = panggil(t, user, http.MethodGet, "/api/v1/agreement-document-types/PKS/schema", nil)
	harapKode(t, w, http.StatusOK, "skema PKS")
	var skema struct {
		Schema struct {
			Sections []json.RawMessage `json:"sections"`
		} `json:"schema"`
	}
	isiData(t, w, &skema)
	if len(skema.Schema.Sections) == 0 {
		t.Error("skema PKS tidak punya bagian form")
	}

	harapKode(t, panggil(t, user, http.MethodGet, "/api/v1/agreement-document-types/TIDAK-ADA/schema", nil), http.StatusNotFound, "skema jenis yang tidak ada")
}

// Data perusahaan (pihak pertama) yang ikut dicetak di setiap perjanjian.
func TestMasterPerjanjian(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/admin/agreement-company-master"

	w := panggil(t, admin, http.MethodGet, alamat, nil)
	harapKode(t, w, http.StatusOK, "mengambil master")
	var awal map[string]any
	isiData(t, w, &awal)
	// Dikembalikan setelah tes, karena pratinjau dan persetujuan memakai data ini.
	t.Cleanup(func() { panggil(t, admin, http.MethodPut, alamat, awal) })

	baru := map[string]any{
		"name": "PT Tes Master", "address": "Jl. Master No. 2",
		"default_signatory_name": "Andi", "default_signatory_position": "Direktur Utama", "default_signing_place": "Jakarta",
	}
	harapKode(t, panggil(t, admin, http.MethodPut, alamat, baru), http.StatusOK, "mengubah master")
	var sesudah map[string]any
	isiData(t, panggil(t, admin, http.MethodGet, alamat, nil), &sesudah)
	if sesudah["name"] != "PT Tes Master" || sesudah["default_signatory_name"] != "Andi" {
		t.Errorf("master setelah diubah: %v", sesudah)
	}

	delete(baru, "name")
	harapKode(t, panggil(t, admin, http.MethodPut, alamat, baru), http.StatusBadRequest, "master tanpa nama")
}

// Alur tanpa lampiran dan tanpa persetujuan akhir: buat, validasi form,
// kepemilikan, ubah, mesin status, ajukan ulang, metadata, hapus.
func TestAlurPerjanjian(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)

	p := buatPerjanjian(t, pemilik, formPerjanjian())
	if p.Status != "SUBMITTED" || p.TicketNumber == "" || p.TemplateID == "" {
		t.Fatalf("perjanjian baru: status %q, tiket %q, template %q", p.Status, p.TicketNumber, p.TemplateID)
	}

	// Beberapa pengajuan boleh menunggu bersamaan: nomor tiket tidak bergantung
	// pada sudah atau belumnya pengajuan lain diberi nomor perjanjian.
	t.Run("DuaPengajuanMenunggu", func(t *testing.T) {
		kedua := buatPerjanjian(t, lain, formPerjanjian())
		if kedua.TicketNumber == p.TicketNumber {
			t.Errorf("dua perjanjian mendapat nomor tiket yang sama: %s", kedua.TicketNumber)
		}
	})

	t.Run("ValidasiForm", func(t *testing.T) {
		kasus := map[string]func(map[string]any){
			"kolom wajib kosong":            func(f map[string]any) { delete(f, "pihak_kedua_nama") },
			"persen termin tidak 100":       func(f map[string]any) { f["termin_2_persen"] = 50 },
			"nilai termin tidak sama":       func(f map[string]any) { f["termin_2_nilai"] = 1 },
			"tanggal selesai sebelum mulai": func(f map[string]any) { f["jangka_waktu_selesai"] = "2025-01-01" },
		}
		for nama, rusak := range kasus {
			form := formPerjanjian()
			rusak(form)
			w := panggil(t, pemilik, http.MethodPost, urlPerjanjian(""), map[string]any{"document_type_code": "PKS", "form_data": form})
			harapKode(t, w, http.StatusBadRequest, nama)
		}
		w := panggil(t, pemilik, http.MethodPost, urlPerjanjian(""), map[string]any{"document_type_code": "TIDAK-ADA", "form_data": formPerjanjian()})
		harapKode(t, w, http.StatusBadRequest, "jenis perjanjian tidak dikenal")
	})

	t.Run("Kepemilikan", func(t *testing.T) {
		if !perjanjianDiDaftar(t, pemilik, "", p.ID) {
			t.Error("perjanjian tidak muncul di daftar pemiliknya")
		}
		if perjanjianDiDaftar(t, lain, "", p.ID) {
			t.Error("perjanjian muncul di daftar USER lain")
		}
		harapKode(t, panggil(t, lain, http.MethodGet, urlPerjanjian("", "/", p.ID), nil), http.StatusNotFound, "USER lain melihat")
		harapKode(t, panggil(t, lain, http.MethodPut, urlPerjanjian("", "/", p.ID), map[string]any{"form_data": formPerjanjian()}), http.StatusBadRequest, "USER lain mengubah")
		harapKode(t, panggil(t, lain, http.MethodDelete, urlPerjanjian("", "/", p.ID), nil), http.StatusBadRequest, "USER lain menghapus")
		harapKode(t, panggil(t, lain, http.MethodPost, urlPerjanjian("", "/", p.ID, "/resubmit"), nil), http.StatusBadRequest, "USER lain mengajukan ulang")

		if !perjanjianDiDaftar(t, legal, "/legal", p.ID) {
			t.Error("perjanjian tidak muncul di daftar LEGAL")
		}
		ambilPerjanjian(t, legal, "/legal", p.ID)
		ambilPerjanjian(t, admin, "/admin", p.ID)
	})

	t.Run("UbahDanStatus", func(t *testing.T) {
		form := formPerjanjian()
		form["jenis_pekerjaan"] = "Jasa audit"
		harapKode(t, panggil(t, pemilik, http.MethodPut, urlPerjanjian("", "/", p.ID), map[string]any{"form_data": form}), http.StatusOK, "pemilik mengubah")
		if got := ambilPerjanjian(t, pemilik, "", p.ID).FormData["jenis_pekerjaan"]; got != "Jasa audit" {
			t.Errorf("jenis pekerjaan setelah diubah: %v", got)
		}

		// Minta revisi hanya bisa dari UNDER_REVIEW, dan wajib ada catatan.
		statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "catatan", http.StatusBadRequest)
		statusPerjanjian(t, legal, p.ID, "STATUS_NGAWUR", "", http.StatusBadRequest)
		statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
		harapKode(t, panggil(t, pemilik, http.MethodPut, urlPerjanjian("", "/", p.ID), map[string]any{"form_data": form}), http.StatusBadRequest, "ubah saat UNDER_REVIEW")
		harapKode(t, panggil(t, pemilik, http.MethodDelete, urlPerjanjian("", "/", p.ID), nil), http.StatusBadRequest, "hapus saat UNDER_REVIEW")
		statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "", http.StatusBadRequest)
		statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "lengkapi alamat", http.StatusOK)
		if got := ambilPerjanjian(t, pemilik, "", p.ID); got.Status != "NEED_REVISION" || got.ApproverNote != "lengkapi alamat" {
			t.Errorf("setelah minta revisi: status %q, catatan %q", got.Status, got.ApproverNote)
		}

		// Saat NEED_REVISION pemilik boleh memperbaiki lalu mengajukan ulang.
		harapKode(t, panggil(t, pemilik, http.MethodPut, urlPerjanjian("", "/", p.ID), map[string]any{"form_data": form}), http.StatusOK, "ubah saat NEED_REVISION")
		harapKode(t, panggil(t, pemilik, http.MethodPost, urlPerjanjian("", "/", p.ID, "/resubmit"), nil), http.StatusOK, "ajukan ulang")
		if got := ambilPerjanjian(t, pemilik, "", p.ID); got.Status != "RESUBMITTED" || got.ApproverNote != "" {
			t.Errorf("setelah ajukan ulang: status %q, catatan %q", got.Status, got.ApproverNote)
		}
	})

	t.Run("MetadataDanFileAkhir", func(t *testing.T) {
		statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
		w := panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"),
			map[string]string{"agreement_number": "001/PKS/TES/2026", "signing_date": "2026-10-11", "signing_place": "Jakarta"})
		harapKode(t, w, http.StatusOK, "mengisi metadata")
		got := ambilPerjanjian(t, pemilik, "", p.ID)
		if got.AgreementNumber == nil || *got.AgreementNumber != "001/PKS/TES/2026" || got.FormData["tanggal_ttd"] != "2026-10-11" {
			t.Errorf("setelah metadata diisi: nomor %v, tanggal %v", got.AgreementNumber, got.FormData["tanggal_ttd"])
		}

		// Nomor perjanjian yang sama tidak boleh dipakai dua pengajuan.
		kembar := buatPerjanjian(t, pemilik, formPerjanjian())
		w = panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", kembar.ID, "/meta"), map[string]string{"agreement_number": "001/PKS/TES/2026"})
		harapKode(t, w, http.StatusBadRequest, "nomor perjanjian kembar")

		// File akhir baru ada setelah disetujui.
		harapKode(t, panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", p.ID, "/pdf"), nil), http.StatusBadRequest, "unduh PDF sebelum selesai")

		// Ditolak adalah akhir: metadata tidak bisa diubah lagi.
		statusPerjanjian(t, legal, p.ID, "REJECTED", "", http.StatusBadRequest)
		statusPerjanjian(t, legal, p.ID, "REJECTED", "tidak sesuai kebijakan", http.StatusOK)
		w = panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_place": "Bandung"})
		harapKode(t, w, http.StatusBadRequest, "metadata setelah ditolak")
		harapKode(t, panggil(t, pemilik, http.MethodPost, urlPerjanjian("", "/", p.ID, "/resubmit"), nil), http.StatusBadRequest, "ajukan ulang setelah ditolak")
	})

	t.Run("Hapus", func(t *testing.T) {
		q := buatPerjanjian(t, pemilik, formPerjanjian())
		if q.TicketNumber == p.TicketNumber {
			t.Errorf("dua perjanjian mendapat nomor tiket yang sama: %s", q.TicketNumber)
		}
		harapKode(t, panggil(t, pemilik, http.MethodDelete, urlPerjanjian("", "/", q.ID), nil), http.StatusOK, "pemilik menghapus")
		harapKode(t, panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", q.ID), nil), http.StatusNotFound, "melihat yang sudah dihapus")
		if lagi := buatPerjanjian(t, pemilik, formPerjanjian()); lagi.TicketNumber == q.TicketNumber {
			t.Errorf("nomor tiket %s dipakai ulang setelah perjanjiannya dihapus", lagi.TicketNumber)
		}
	})
}

// Lampiran perjanjian: unggah, unduh, kepemilikan, dan putaran saat ajukan ulang.
func TestLampiranPerjanjian(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)

	data, _ := json.Marshal(map[string]any{"document_type_code": "PKS", "form_data": formPerjanjian()})
	kolom := map[string]string{"data": string(data)}
	isiA, isiB := []byte("penawaran"), []byte("penawaran revisi")

	w := panggilForm(t, pemilik, http.MethodPost, urlPerjanjian(""), kolom, berkas{"attachments", "penawaran.pdf", isiA})
	harapKode(t, w, http.StatusCreated, "membuat perjanjian dengan lampiran")
	var p perjanjian
	isiData(t, w, &p)
	if len(p.Attachments) != 1 || p.Attachments[0].FileName != "penawaran.pdf" || p.Attachments[0].UploadRound != 1 {
		t.Fatalf("lampiran setelah dibuat: %+v", p.Attachments)
	}

	lampiran := "/attachments/" + p.Attachments[0].ID
	w = panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", p.ID, lampiran), nil)
	harapKode(t, w, http.StatusOK, "pemilik mengunduh lampiran")
	if !bytes.Equal(w.Body.Bytes(), isiA) {
		t.Errorf("isi lampiran yang diunduh berbeda: %q", w.Body)
	}
	harapKode(t, panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", p.ID, lampiran), nil), http.StatusOK, "LEGAL mengunduh lampiran")
	harapKode(t, panggil(t, lain, http.MethodGet, urlPerjanjian("", "/", p.ID, lampiran), nil), http.StatusNotFound, "USER lain mengunduh lampiran")

	// Lampiran yang ditolak membatalkan seluruh pengajuan.
	w = panggilForm(t, pemilik, http.MethodPost, urlPerjanjian(""), kolom, berkas{"attachments", "program.exe", []byte("x")})
	harapKode(t, w, http.StatusBadRequest, "lampiran .exe")

	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "lampiran kurang", http.StatusOK)
	w = panggilForm(t, pemilik, http.MethodPost, urlPerjanjian("", "/", p.ID, "/resubmit"), nil, berkas{"attachments", "revisi.pdf", isiB})
	harapKode(t, w, http.StatusOK, "ajukan ulang dengan lampiran")
	p = ambilPerjanjian(t, pemilik, "", p.ID)
	putaran := map[int]int{}
	for _, a := range p.Attachments {
		putaran[a.UploadRound]++
	}
	if len(p.Attachments) != 2 || putaran[1] != 1 || putaran[2] != 1 {
		t.Errorf("lampiran setelah ajukan ulang: %+v", p.Attachments)
	}
}

// Persetujuan akhir membentuk DOCX dan PDF dari template.
//
// Butuh LibreOffice, jadi hanya jalan di dalam container backend: lihat
// scripts/cek-docker.ps1.
func TestPersetujuanPerjanjian(t *testing.T) {
	butuhMinIO(t)
	butuhLibreOffice(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)

	p := buatPerjanjian(t, pemilik, formPerjanjian())
	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)

	// Tanggal tanda tangan wajib diisi approver sebelum menyetujui.
	statusPerjanjian(t, legal, p.ID, "COMPLETED", "", http.StatusBadRequest)
	w := panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_date": "2026-10-11"})
	harapKode(t, w, http.StatusOK, "mengisi tanggal tanda tangan")

	w = panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", p.ID, "/preview"), nil)
	harapKode(t, w, http.StatusOK, "pratinjau sebelum disetujui")
	if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
		t.Error("pratinjau bukan PDF")
	}

	statusPerjanjian(t, legal, p.ID, "COMPLETED", "", http.StatusOK)
	if s := ambilPerjanjian(t, pemilik, "", p.ID).Status; s != "COMPLETED" {
		t.Fatalf("status setelah disetujui: %s", s)
	}

	w = panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", p.ID, "/pdf"), nil)
	harapKode(t, w, http.StatusOK, "pemilik mengunduh PDF akhir")
	if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
		t.Error("file akhir bukan PDF")
	}
	w = panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", p.ID, "/docx"), nil)
	harapKode(t, w, http.StatusOK, "LEGAL mengunduh DOCX akhir")
	if !bytes.HasPrefix(w.Body.Bytes(), []byte("PK")) {
		t.Error("file akhir bukan DOCX")
	}
}

// Harapan yang benar tetapi belum dipenuhi kode. Untuk melihatnya gagal apa
// adanya: TES_BUG=1 go test ./internal/router -run TestPerjanjianBelumBenar -v
func TestPerjanjianBelumBenar(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)

	t.Run("PesanNomorKembarBukanErrorDatabase", func(t *testing.T) {
		bugDiketahui(t, "nomor perjanjian yang sudah dipakai ditolak dengan pesan mentah dari database (SQLSTATE ...), bukan pesan untuk pengguna")
		var w *httptest.ResponseRecorder
		for range 2 {
			p := buatPerjanjian(t, pemilik, formPerjanjian())
			w = panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"agreement_number": "002/PKS/TES/2026"})
		}
		harapKode(t, w, http.StatusBadRequest, "nomor perjanjian kembar")
		if strings.Contains(w.Body.String(), "SQLSTATE") {
			t.Errorf("pesan error database bocor ke pengguna: %s", w.Body)
		}
	})
}

// Template perjanjian: daftar, rujukan placeholder, unduh, dan penolakan unggahan.
func TestTemplatePerjanjian(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	const alamat = "/api/v1/legal/agreement-templates"

	w := panggil(t, legal, http.MethodGet, alamat+"/placeholders", nil)
	harapKode(t, w, http.StatusOK, "rujukan placeholder")
	var rujukan []struct {
		Token string `json:"token"`
	}
	isiData(t, w, &rujukan)
	if len(rujukan) == 0 {
		t.Error("rujukan placeholder kosong")
	}

	harapKode(t, panggil(t, legal, http.MethodGet, alamat+"/00000000-0000-0000-0000-000000000000", nil), http.StatusNotFound, "template yang tidak ada")
	harapKode(t, panggilForm(t, legal, http.MethodPost, alamat, nil), http.StatusBadRequest, "unggah tanpa file")
	harapKode(t, panggil(t, legal, http.MethodPost, alamat+"/00000000-0000-0000-0000-000000000000/activate", nil), http.StatusBadRequest, "mengaktifkan template yang tidak ada")

	t.Run("TemplateBawaan", func(t *testing.T) {
		butuhMinIO(t)
		w := panggil(t, legal, http.MethodGet, alamat, nil)
		harapKode(t, w, http.StatusOK, "daftar template")
		var daftar []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		isiData(t, w, &daftar)
		if len(daftar) != 1 || daftar[0].Status != "ACTIVE" {
			t.Fatalf("template setelah seed: %+v", daftar)
		}
		id := daftar[0].ID

		harapKode(t, panggil(t, legal, http.MethodGet, alamat+"/"+id, nil), http.StatusOK, "detail template")
		w = panggil(t, legal, http.MethodGet, alamat+"/"+id+"/download", nil)
		harapKode(t, w, http.StatusOK, "unduh template")
		if !bytes.HasPrefix(w.Body.Bytes(), []byte("PK")) {
			t.Error("template yang diunduh bukan DOCX")
		}

		// Mengaktifkan template yang sudah aktif tidak mengubah apa pun.
		harapKode(t, panggil(t, legal, http.MethodPost, alamat+"/"+id+"/activate", nil), http.StatusOK, "mengaktifkan template aktif")
		harapKode(t, panggilForm(t, legal, http.MethodPost, alamat, nil, berkas{"file", "template.pdf", []byte("%PDF")}), http.StatusBadRequest, "unggah template bukan DOCX")
	})
}
