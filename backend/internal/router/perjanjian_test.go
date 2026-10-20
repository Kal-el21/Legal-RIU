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
	"regexp"
	"strings"
	"sync"
	"testing"

	"legal-riu-portal/internal/assets"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"
	"legal-riu-portal/internal/service"

	"github.com/google/uuid"
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
		"surat_penawaran_nomor": "SP-01/2026", "surat_penawaran_perihal": "Penawaran jasa konsultasi", "surat_penawaran_tanggal": "2025-12-01",
		"surat_penunjukan_nomor": "SPK-01/2026", "surat_penunjukan_perihal": "Penunjukan konsultan", "surat_penunjukan_tanggal": "2025-12-15",
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

// isiPengajuan membentuk isi permintaan membuat pengajuan seperti yang dikirim
// frontend: kode jenis, versi template dari skema jenis itu, dan isian form.
func isiPengajuan(t *testing.T, token, kode string, form map[string]any) map[string]any {
	t.Helper()
	isi := map[string]any{"document_type_code": kode, "form_data": form}
	w := panggil(t, token, http.MethodGet, "/api/v1/agreement-document-types/"+kode+"/schema", nil)
	if w.Code == http.StatusOK {
		var skema struct {
			Schema struct {
				TemplateID string `json:"template_id"`
			} `json:"schema"`
		}
		isiData(t, w, &skema)
		isi["template_id"] = skema.Schema.TemplateID
	}
	return isi
}

func buatPerjanjian(t *testing.T, token string, form map[string]any) perjanjian {
	t.Helper()
	w := panggil(t, token, http.MethodPost, urlPerjanjian(""), isiPengajuan(t, token, "PKS", form))
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
// Hanya jenis yang punya template aktif yang tampil, dan template butuh MinIO.
func TestJenisPerjanjian(t *testing.T) {
	user := masukPeran(t, entity.RoleUser)
	harapKode(t, panggil(t, user, http.MethodGet, "/api/v1/agreement-document-types/TIDAK-ADA/schema", nil), http.StatusNotFound, "skema jenis yang tidak ada")
	butuhMinIO(t)

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
			TemplateID string            `json:"template_id"`
			Sections   []json.RawMessage `json:"sections"`
		} `json:"schema"`
	}
	isiData(t, w, &skema)
	if len(skema.Schema.Sections) == 0 {
		t.Error("skema PKS tidak punya bagian form")
	}
	if skema.Schema.TemplateID != templateAktif(t).ID {
		t.Errorf("skema PKS menyebut template %q, harus versi aktif", skema.Schema.TemplateID)
	}
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
			w := panggil(t, pemilik, http.MethodPost, urlPerjanjian(""), isiPengajuan(t, pemilik, "PKS", form))
			harapKode(t, w, http.StatusBadRequest, nama)
			if strings.Contains(w.Body.String(), "versi template") {
				t.Errorf("%s ditolak karena versi template, bukan karena isiannya: %s", nama, w.Body)
			}
		}
		w := panggil(t, pemilik, http.MethodPost, urlPerjanjian(""), isiPengajuan(t, pemilik, "TIDAK-ADA", formPerjanjian()))
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

	data, _ := json.Marshal(isiPengajuan(t, pemilik, "PKS", formPerjanjian()))
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

// Dua persetujuan yang hampir bersamaan (mis. tombol Setujui ditekan dua kali
// selagi PDF disiapkan) hanya boleh menghasilkan satu persetujuan, dan file
// final milik persetujuan itu harus tetap utuh. Sebelumnya permintaan yang kalah
// menghapus file di path yang sama, sehingga perjanjian COMPLETED tanpa file.
func TestPersetujuanGandaTidakMenghapusFile(t *testing.T) {
	butuhMinIO(t)
	butuhLibreOffice(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)

	p := buatPerjanjian(t, pemilik, formPerjanjian())
	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	w := panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_date": "2026-10-11"})
	harapKode(t, w, http.StatusOK, "mengisi tanggal tanda tangan")

	kode := make([]int, 2)
	var wg sync.WaitGroup
	for i := range kode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kode[i] = panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/status"), map[string]string{"status": "COMPLETED"}).Code
		}()
	}
	wg.Wait()
	berhasil := 0
	for _, k := range kode {
		if k == http.StatusOK {
			berhasil++
		}
	}
	if berhasil != 1 {
		t.Errorf("dua persetujuan bersamaan: kode %v, harap tepat satu 200", kode)
	}
	if s := ambilPerjanjian(t, pemilik, "", p.ID).Status; s != "COMPLETED" {
		t.Fatalf("status setelah dua persetujuan: %s", s)
	}

	w = panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", p.ID, "/pdf"), nil)
	if w.Code != http.StatusOK || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
		t.Errorf("PDF akhir setelah dua persetujuan: %d (%d byte)", w.Code, w.Body.Len())
	}
	w = panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", p.ID, "/docx"), nil)
	if w.Code != http.StatusOK || !bytes.HasPrefix(w.Body.Bytes(), []byte("PK")) {
		t.Errorf("DOCX akhir setelah dua persetujuan: %d (%d byte)", w.Code, w.Body.Len())
	}
}

// Harapan yang benar tetapi belum dipenuhi kode. Untuk melihatnya gagal apa
// adanya: TES_BUG=1 go test ./internal/router -run TestPerjanjianBelumBenar -v
func TestPerjanjianBelumBenar(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)

	// Nomor yang sudah dipakai ditolak dengan pesan untuk pengguna, bukan pesan
	// mentah dari database.
	t.Run("PesanNomorKembarBukanErrorDatabase", func(t *testing.T) {
		var w *httptest.ResponseRecorder
		for range 2 {
			p := buatPerjanjian(t, pemilik, formPerjanjian())
			w = panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"agreement_number": "002/PKS/TES/2026"})
		}
		harapKode(t, w, http.StatusBadRequest, "nomor perjanjian kembar")
		if strings.Contains(w.Body.String(), "SQLSTATE") || !strings.Contains(w.Body.String(), "sudah dipakai") {
			t.Errorf("pesan untuk nomor kembar: %s", w.Body)
		}
	})
}

// Template perjanjian: daftar, rujukan placeholder, unduh, dan penolakan unggahan.
// LEGAL hanya melihat; unggah dan aktivasi hanya ADMIN (keputusan T1).
func TestTemplatePerjanjian(t *testing.T) {
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/legal/agreement-templates"
	const kelola = "/api/v1/admin/agreement-templates"

	harapKode(t, panggil(t, legal, http.MethodGet, alamat+"/00000000-0000-0000-0000-000000000000", nil), http.StatusNotFound, "template yang tidak ada")
	harapKode(t, panggilForm(t, legal, http.MethodPost, alamat, nil), http.StatusForbidden, "LEGAL mengunggah template")
	harapKode(t, panggil(t, legal, http.MethodPost, alamat+"/00000000-0000-0000-0000-000000000000/activate", nil), http.StatusForbidden, "LEGAL mengaktifkan template")
	w := panggilForm(t, admin, http.MethodPost, kelola, nil)
	harapKode(t, w, http.StatusBadRequest, "unggah tanpa file")
	if !strings.Contains(w.Body.String(), "emplate wajib diunggah") {
		t.Errorf("pesan unggah tanpa file: %s", w.Body)
	}
	harapKode(t, panggil(t, admin, http.MethodPost, kelola+"/00000000-0000-0000-0000-000000000000/activate", nil), http.StatusBadRequest, "mengaktifkan template yang tidak ada")
	harapKode(t, panggil(t, legal, http.MethodGet, alamat+"/placeholders?code=TIDAK-ADA", nil), http.StatusNotFound, "rujukan placeholder jenis yang tidak ada")

	t.Run("TemplateBawaan", func(t *testing.T) {
		butuhMinIO(t)
		// Rujukan placeholder dibentuk dari field versi aktif, jadi butuh template.
		w := panggil(t, legal, http.MethodGet, alamat+"/placeholders", nil)
		harapKode(t, w, http.StatusOK, "rujukan placeholder")
		var rujukan []struct {
			Token string `json:"token"`
		}
		isiData(t, w, &rujukan)
		if len(rujukan) == 0 {
			t.Error("rujukan placeholder kosong")
		}

		w = panggil(t, legal, http.MethodGet, alamat, nil)
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
		harapKode(t, panggil(t, admin, http.MethodPost, kelola+"/"+id+"/activate", nil), http.StatusOK, "mengaktifkan template aktif")
		harapKode(t, panggilForm(t, admin, http.MethodPost, kelola, nil, berkas{"file", "template.pdf", []byte("%PDF")}), http.StatusBadRequest, "unggah template bukan DOCX")
	})
}

// Baris agreement_templates yang dibaca tes pemasangan template bawaan.
type barisTemplate struct {
	ID       string
	Version  int
	Status   string
	IsLegacy bool
	Checksum string
}

const kolomTemplate = "SELECT id, version, status, is_legacy, checksum FROM agreement_templates"

func checksumTemplateBawaan(t *testing.T) string {
	t.Helper()
	data, err := assets.Files.ReadFile("templates/pks-template/Draft-PKS-RIU.docx")
	if err != nil {
		t.Fatal(err)
	}
	return service.TemplateChecksum(data)
}

func templateDB(t *testing.T, id string) barisTemplate {
	t.Helper()
	var b barisTemplate
	if err := dbTes.Raw(kolomTemplate+" WHERE id = ?", id).Scan(&b).Error; err != nil || b.ID == "" {
		t.Fatalf("template %s tidak terbaca: %v", id, err)
	}
	return b
}

func templateBawaanDB(t *testing.T) []barisTemplate {
	t.Helper()
	var daftar []barisTemplate
	if err := dbTes.Raw(kolomTemplate+" WHERE code = 'PKS' AND checksum = ? AND deleted_at IS NULL", checksumTemplateBawaan(t)).Scan(&daftar).Error; err != nil {
		t.Fatal(err)
	}
	return daftar
}

func templatePengajuan(t *testing.T, id string) (templateID, checksum string) {
	t.Helper()
	var b struct{ TemplateID, TemplateChecksum string }
	if err := dbTes.Raw("SELECT template_id, template_checksum FROM agreement_documents WHERE id = ?", id).Scan(&b).Error; err != nil {
		t.Fatal(err)
	}
	return b.TemplateID, b.TemplateChecksum
}

// Seed dijalankan dua kali, seperti server yang dinyalakan ulang. Field jenis
// dokumen ikut diisi, seperti urutan di RunAllMigrationsAndSeeds.
func jalankanSeedTemplate(t *testing.T) {
	t.Helper()
	for i := 1; i <= 2; i++ {
		if err := seed.SeedAgreementTemplate(dbTes, storeTes); err != nil {
			t.Fatalf("seed template putaran %d: %v", i, err)
		}
		if err := seed.SeedAgreementTypes(dbTes); err != nil {
			t.Fatalf("seed jenis dokumen putaran %d: %v", i, err)
		}
	}
}

func eksekusi(t *testing.T, sql string, args ...any) {
	t.Helper()
	if err := dbTes.Exec(sql, args...).Error; err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

func templateAktif(t *testing.T) barisTemplate {
	t.Helper()
	var b barisTemplate
	if err := dbTes.Raw(kolomTemplate + " WHERE code = 'PKS' AND status = 'ACTIVE'").Scan(&b).Error; err != nil || b.ID == "" {
		t.Fatalf("template aktif tidak terbaca: %v", err)
	}
	return b
}

// salinTemplate menambah baris template baru yang memakai file dan field
// template dari, supaya pengajuan tetap bisa dibuat dan dipratinjau.
func salinTemplate(t *testing.T, dari, id, checksum, status string, legacy bool, uploadedBy any) {
	t.Helper()
	eksekusi(t, `INSERT INTO agreement_templates
		(id, created_at, updated_at, code, version, name, file_name, file_path, checksum, status, is_legacy, note, uploaded_by, fields)
		SELECT ?, now(), now(), code, (SELECT MAX(version) + 1 FROM agreement_templates WHERE code = 'PKS'),
			'uji', file_name, file_path, ?, ?, ?, 'uji', ?, fields
		FROM agreement_templates WHERE id = ?`, id, checksum, status, legacy, uploadedBy, dari)
}

// Database dari sebelum PR 8: template aktif adalah template lama yang diisi
// lewat pencocokan kalimat. Seed memasang template bawaan baru sekali saja,
// mengaktifkannya, dan memindahkan pengajuan yang belum selesai ke sana;
// pengajuan yang sudah selesai tetap di template lamanya.
func TestTemplatePKSDiperbarui(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	berjalan := buatPerjanjian(t, user, formPerjanjian())
	selesai := buatPerjanjian(t, user, formPerjanjian())
	lama := berjalan.TemplateID

	// Status Selesai diisi langsung, karena alur persetujuan butuh LibreOffice.
	eksekusi(t, "UPDATE agreement_documents SET status = ? WHERE id = ?", entity.StatusCompleted, selesai.ID)
	eksekusi(t, "UPDATE agreement_templates SET is_legacy = true, checksum = 'sebelum-pr8-uji' WHERE id = ?", lama)

	jalankanSeedTemplate(t)

	bawaan := templateBawaanDB(t)
	if len(bawaan) != 1 {
		t.Fatalf("template bawaan terpasang %d kali, harus sekali", len(bawaan))
	}
	baru, versiLama := bawaan[0], templateDB(t, lama)
	if baru.Status != string(entity.TemplateStatusActive) || baru.IsLegacy || baru.Version <= versiLama.Version {
		t.Errorf("template bawaan baru: %+v (versi lama %d)", baru, versiLama.Version)
	}
	if versiLama.Status != string(entity.TemplateStatusArchived) {
		t.Errorf("template lama berstatus %s, harus ARCHIVED", versiLama.Status)
	}
	if id, checksum := templatePengajuan(t, berjalan.ID); id != baru.ID || checksum != baru.Checksum {
		t.Errorf("pengajuan berjalan menunjuk template %s (checksum %s), harus pindah ke %s", id, checksum, baru.ID)
	}
	if id, _ := templatePengajuan(t, selesai.ID); id != lama {
		t.Errorf("pengajuan selesai ikut dipindah ke %s; dokumen finalnya dari template lama", id)
	}

	t.Run("Pratinjau", func(t *testing.T) {
		butuhLibreOffice(t)
		w := panggil(t, user, http.MethodGet, urlPerjanjian("", "/", berjalan.ID, "/preview"), nil)
		harapKode(t, w, http.StatusOK, "pratinjau pengajuan yang dipindah")
		if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
			t.Error("pratinjau bukan PDF")
		}
	})
}

// Template yang diunggah dan diaktifkan Admin tidak diganti template bawaan.
// Pengajuan di template lama tetap pindah ke template bawaan, karena hanya isi
// template bawaan yang dijamin sama dengan template lama.
func TestTemplateUnggahanTidakDitimpa(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	checksum := checksumTemplateBawaan(t)

	aktif := templateAktif(t)
	var admin string
	if err := dbTes.Raw("SELECT id FROM users WHERE role = ? AND deleted_at IS NULL LIMIT 1", entity.RoleAdmin).Scan(&admin).Error; err != nil || admin == "" {
		t.Fatalf("admin tidak terbaca: %v", err)
	}
	t.Cleanup(func() {
		// Template bawaan kembali aktif untuk tes berikutnya.
		dbTes.Exec("UPDATE agreement_templates SET status = 'ARCHIVED' WHERE code = 'PKS' AND status = 'ACTIVE'")
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE code = 'PKS' AND checksum = ?", checksum)
	})

	unggahan, lama := uuid.NewString(), uuid.NewString()
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED' WHERE id = ?", aktif.ID)
	salinTemplate(t, aktif.ID, unggahan, "unggahan-admin-uji", string(entity.TemplateStatusActive), false, admin)
	salinTemplate(t, aktif.ID, lama, "sebelum-pr8-uji-2", string(entity.TemplateStatusArchived), true, nil)
	// Template bawaan versi ini belum terpasang.
	eksekusi(t, "UPDATE agreement_templates SET checksum = 'bawaan-sebelumnya-uji' WHERE code = 'PKS' AND checksum = ?", checksum)

	p := buatPerjanjian(t, user, formPerjanjian())
	eksekusi(t, "UPDATE agreement_documents SET template_id = ? WHERE id = ?", lama, p.ID)

	jalankanSeedTemplate(t)

	if s := templateDB(t, unggahan).Status; s != string(entity.TemplateStatusActive) {
		t.Errorf("template unggahan Admin berstatus %s setelah seed, harus tetap ACTIVE", s)
	}
	bawaan := templateBawaanDB(t)
	if len(bawaan) != 1 || bawaan[0].Status != string(entity.TemplateStatusDraft) {
		t.Fatalf("template bawaan setelah seed: %+v, harus terpasang sekali dan tidak aktif", bawaan)
	}
	if id, _ := templatePengajuan(t, p.ID); id != bawaan[0].ID {
		t.Errorf("pengajuan di template lama menunjuk %s, harus pindah ke template bawaan %s (bukan unggahan Admin)", id, bawaan[0].ID)
	}
}

// Perjanjian Selesai dipratinjau dari file finalnya. Kalau file itu hilang, jawabannya
// 404 seperti unduhan final, bukan dokumen yang dibentuk ulang dari template dan
// master Pihak Pertama yang sekarang berlaku.
func TestPratinjauFinalHilang(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	for _, c := range []struct{ nama, path string }{
		{"TanpaPathFile", ""},
		{"FileTidakAda", "agreement-documents/tidak-ada/final/agreement-hilang.pdf"},
	} {
		t.Run(c.nama, func(t *testing.T) {
			p := buatPerjanjian(t, user, formPerjanjian())
			eksekusi(t, "UPDATE agreement_documents SET status = ?, generated_pdf_path = ? WHERE id = ?", entity.StatusCompleted, c.path, p.ID)
			harapKode(t, panggil(t, user, http.MethodGet, urlPerjanjian("", "/", p.ID, "/preview"), nil), http.StatusNotFound, "pratinjau pemilik")
			harapKode(t, panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", p.ID, "/preview"), nil), http.StatusNotFound, "pratinjau approver")
		})
	}
}

// Template versi lama (tanpa placeholder, dulu diisi lewat pencocokan kalimat)
// tetap tersimpan sebagai riwayat, tetapi tidak bisa lagi dipratinjau, diaktifkan,
// atau dipakai membentuk pengajuan. Penolakannya berupa pesan, bukan galat server.
func TestTemplateLamaDitolak(t *testing.T) {
	butuhMinIO(t)
	admin := masukPeran(t, entity.RoleAdmin)
	user := masukPeran(t, entity.RoleUser)
	const alamat = "/api/v1/admin/agreement-templates/"

	lama := uuid.NewString()
	salinTemplate(t, templateAktif(t).ID, lama, "sebelum-pr8-uji-3", string(entity.TemplateStatusArchived), true, nil)

	w := panggil(t, admin, http.MethodGet, alamat+lama+"/preview", nil)
	harapKode(t, w, http.StatusBadRequest, "pratinjau template lama")
	if !strings.Contains(w.Body.String(), "format lama") {
		t.Errorf("pesan pratinjau template lama: %s", w.Body.String())
	}
	harapKode(t, panggil(t, admin, http.MethodPost, alamat+lama+"/activate", nil), http.StatusBadRequest, "mengaktifkan template lama")
	if s := templateDB(t, lama).Status; s != string(entity.TemplateStatusArchived) {
		t.Errorf("template lama berstatus %s setelah ditolak, harus tetap ARCHIVED", s)
	}

	p := buatPerjanjian(t, user, formPerjanjian())
	eksekusi(t, "UPDATE agreement_documents SET template_id = ? WHERE id = ?", lama, p.ID)
	harapKode(t, panggil(t, user, http.MethodGet, urlPerjanjian("", "/", p.ID, "/preview"), nil), http.StatusBadRequest, "pratinjau pengajuan di template lama")
}

var halamanPDFRE = regexp.MustCompile(`/Type\s*/Page[^s]`)

// Watermark DRAFT dipasang di header, bukan di isi dokumen, jadi pratinjau dan
// dokumen final berpisah halaman di titik yang sama.
func TestWatermarkTampilDiPDF(t *testing.T) {
	butuhMinIO(t)
	butuhLibreOffice(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)

	form := formPerjanjian()
	form["pihak_kedua_alamat"] = strings.Repeat("Jl. Contoh Panjang Nomor 123 Blok A, Kawasan Industri, ", 11)
	var poin []string
	for i := 0; i < 12; i++ {
		poin = append(poin, "Pelaksanaan pekerjaan tahap lanjutan sesuai jadwal yang disepakati PARA PIHAK")
	}
	form["ruang_lingkup"] = strings.Join(poin, "\n")
	p := buatPerjanjian(t, pemilik, form)
	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_date": "2026-10-11"}), http.StatusOK, "mengisi tanggal tanda tangan")

	w := panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", p.ID, "/preview"), nil)
	harapKode(t, w, http.StatusOK, "pratinjau")
	pratinjau := len(halamanPDFRE.FindAll(w.Body.Bytes(), -1))

	statusPerjanjian(t, legal, p.ID, "COMPLETED", "", http.StatusOK)
	w = panggil(t, pemilik, http.MethodGet, urlPerjanjian("", "/", p.ID, "/pdf"), nil)
	harapKode(t, w, http.StatusOK, "PDF final")
	final := len(halamanPDFRE.FindAll(w.Body.Bytes(), -1))

	if pratinjau < 20 {
		t.Fatalf("pratinjau hanya %d halaman; penghitung halaman PDF kemungkinan rusak", pratinjau)
	}
	if pratinjau != final {
		t.Errorf("pratinjau %d halaman, final %d halaman; penanda draft menggeser isi", pratinjau, final)
	}
}
