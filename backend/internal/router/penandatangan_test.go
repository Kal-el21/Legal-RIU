package router

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"

	"github.com/google/uuid"
)

// Migrasi PR 11: penandatangan tunggal master Pihak Pertama menjadi baris pertama
// Daftar Penandatangan (tak terbatas, default), sekali saja. Baris yang kemudian
// dihapus Admin tidak dibuat ulang saat server menyala lagi.
func TestMigrasiPenandatangan(t *testing.T) {
	var asli struct{ ID, Nama, Jabatan string }
	dbTes.Raw("SELECT id, default_signatory_name AS nama, default_signatory_position AS jabatan FROM agreement_company_masters WHERE is_active ORDER BY created_at LIMIT 1").Scan(&asli)
	if asli.ID == "" {
		t.Fatal("master Pihak Pertama tidak ada di database tes")
	}
	t.Cleanup(func() {
		dbTes.Exec("UPDATE agreement_company_masters SET default_signatory_name = ?, default_signatory_position = ? WHERE id = ?", asli.Nama, asli.Jabatan, asli.ID)
		dbTes.Exec("DELETE FROM agreement_signatories")
		seed.SeedAgreementSignatories(dbTes)
	})
	eksekusi(t, "UPDATE agreement_company_masters SET default_signatory_name = 'Budi Uji', default_signatory_position = 'Direktur Uji' WHERE id = ?", asli.ID)
	eksekusi(t, "DELETE FROM agreement_signatories")

	jalan := func() {
		t.Helper()
		if err := seed.SeedAgreementSignatories(dbTes); err != nil {
			t.Fatal(err)
		}
	}
	jalan()
	jalan()
	var baris []struct {
		Name, Position      string
		MaxValue            *int64
		IsDefault, IsActive bool
	}
	dbTes.Raw("SELECT name, position, max_value, is_default, is_active FROM agreement_signatories").Scan(&baris)
	if len(baris) != 1 || baris[0].Name != "Budi Uji" || baris[0].Position != "Direktur Uji" || baris[0].MaxValue != nil || !baris[0].IsDefault || !baris[0].IsActive {
		t.Fatalf("Daftar Penandatangan setelah migrasi dua kali: %+v, harus satu baris Budi Uji, tak terbatas, default, aktif", baris)
	}

	eksekusi(t, "UPDATE agreement_signatories SET deleted_at = now()")
	jalan()
	if n := hitungBaris(t, "SELECT COUNT(*) FROM agreement_signatories WHERE deleted_at IS NULL"); n != 0 {
		t.Errorf("%d penandatangan dibuat ulang setelah dihapus Admin", n)
	}
}

const alamatPenandatangan = "/api/v1/admin/agreement-signatories"

type penandatangan struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Position  string `json:"position"`
	MaxValue  *int64 `json:"max_value"`
	IsDefault bool   `json:"is_default"`
	IsActive  bool   `json:"is_active"`
}

// Admin mengelola Daftar Penandatangan. Batas nilai maksimum tidak boleh kembar dan
// hanya satu yang tak terbatas, di antara penandatangan aktif; hanya satu default.
// LEGAL hanya membaca (untuk pilihan saat persetujuan), dan setiap perubahan tercatat
// di audit log.
func TestKelolaPenandatangan(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	legal := masukPeran(t, entity.RoleLegal)
	user := masukPeran(t, entity.RoleUser)
	eksekusi(t, "DELETE FROM agreement_signatories")
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_signatories")
		seed.SeedAgreementSignatories(dbTes)
	})
	auditSebelum := hitungBaris(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'SIGNATORY_CHANGE'")
	rp := func(n int64) *int64 { return &n }
	isi := func(nama, jabatan string, maks *int64, def bool) map[string]any {
		return map[string]any{"name": nama, "position": jabatan, "max_value": maks, "is_default": def, "is_active": true}
	}
	buat := func(nama string, maks *int64, def bool) penandatangan {
		t.Helper()
		w := panggil(t, admin, http.MethodPost, alamatPenandatangan, isi(nama, "Jabatan "+nama, maks, def))
		harapKode(t, w, http.StatusCreated, "menambah "+nama)
		var p penandatangan
		isiData(t, w, &p)
		return p
	}
	ditolak := func(w *httptest.ResponseRecorder, nama, pesan string) {
		t.Helper()
		harapKode(t, w, http.StatusBadRequest, nama)
		if !strings.Contains(w.Body.String(), pesan) {
			t.Errorf("%s: pesan harus menyebut %q: %s", nama, pesan, w.Body)
		}
	}
	daftar := func(token, awalan string) []penandatangan {
		t.Helper()
		w := panggil(t, token, http.MethodGet, "/api/v1"+awalan+"/agreement-signatories", nil)
		harapKode(t, w, http.StatusOK, awalan+" membaca Daftar Penandatangan")
		var d []penandatangan
		isiData(t, w, &d)
		return d
	}
	nama := func(d []penandatangan) string {
		var out []string
		for _, p := range d {
			out = append(out, p.Name)
		}
		return strings.Join(out, " ")
	}

	buat("Andi", nil, true)
	sari := buat("Sari", rp(5_000_000_000), false)
	budi := buat("Budi", rp(500_000_000), false)
	// Urut dari batas terkecil; tak terbatas paling akhir.
	if got := nama(daftar(admin, "/admin")); got != "Budi Sari Andi" {
		t.Errorf("urutan daftar %q, harus Budi Sari Andi", got)
	}

	ditolak(panggil(t, admin, http.MethodPost, alamatPenandatangan, isi("Rina", "Direktur", rp(500_000_000), false)), "batas kembar", "Rp 500.000.000 sudah dipakai Budi")
	ditolak(panggil(t, admin, http.MethodPost, alamatPenandatangan, isi("Rina", "Direktur", nil, false)), "dua tak terbatas", "tak terbatas sudah dipakai Andi")
	ditolak(panggil(t, admin, http.MethodPost, alamatPenandatangan, isi(" ", "Direktur", rp(1), false)), "nama kosong", "Nama penandatangan wajib diisi")
	ditolak(panggil(t, admin, http.MethodPost, alamatPenandatangan, isi("Rina", " ", rp(1), false)), "jabatan kosong", "Jabatan penandatangan wajib diisi")
	ditolak(panggil(t, admin, http.MethodPost, alamatPenandatangan, isi("Rina", "Direktur", rp(0), false)), "batas nol", "Nilai maksimum harus lebih dari 0")
	ditolak(panggil(t, admin, http.MethodPut, alamatPenandatangan+"/"+sari.ID, isi("Sari", "Direktur", rp(500_000_000), false)), "ubah ke batas kembar", "sudah dipakai Budi")

	// Penandatangan nonaktif tidak dihitung: penggantinya boleh memakai batas yang sama,
	// tetapi mengaktifkannya kembali ditolak selama batasnya dipakai.
	nonaktif := isi("Sari", "Jabatan Sari", rp(5_000_000_000), false)
	nonaktif["is_active"] = false
	harapKode(t, panggil(t, admin, http.MethodPut, alamatPenandatangan+"/"+sari.ID, nonaktif), http.StatusOK, "menonaktifkan Sari")
	rina := buat("Rina", rp(5_000_000_000), false)
	nonaktif["position"] = "Direktur (purnatugas)"
	harapKode(t, panggil(t, admin, http.MethodPut, alamatPenandatangan+"/"+sari.ID, nonaktif), http.StatusOK, "mengubah Sari yang nonaktif")
	ditolak(panggil(t, admin, http.MethodPut, alamatPenandatangan+"/"+sari.ID, isi("Sari", "Jabatan Sari", rp(5_000_000_000), false)), "aktifkan Sari", "sudah dipakai Rina")

	// Hanya satu default: menjadikan Budi default melepas default Andi.
	harapKode(t, panggil(t, admin, http.MethodPut, alamatPenandatangan+"/"+budi.ID, isi("Budi", "Jabatan Budi", rp(500_000_000), true)), http.StatusOK, "menjadikan Budi default")
	for _, p := range daftar(admin, "/admin") {
		if p.IsDefault != (p.ID == budi.ID) {
			t.Errorf("%s default %v; hanya Budi yang boleh default", p.Name, p.IsDefault)
		}
	}

	harapKode(t, panggil(t, admin, http.MethodDelete, alamatPenandatangan+"/"+rina.ID, nil), http.StatusOK, "menghapus Rina")
	harapKode(t, panggil(t, admin, http.MethodPut, alamatPenandatangan+"/"+rina.ID, isi("Rina", "Direktur", rp(1), false)), http.StatusNotFound, "mengubah penandatangan yang dihapus")
	if got := nama(daftar(admin, "/admin")); got != "Budi Sari Andi" {
		t.Errorf("daftar setelah hapus %q, harus Budi Sari Andi", got)
	}

	// LEGAL membaca daftar yang sama untuk pilihan persetujuan, tetapi tidak bisa mengubah.
	if got := nama(daftar(legal, "/legal")); got != "Budi Sari Andi" {
		t.Errorf("daftar untuk LEGAL %q", got)
	}
	harapKode(t, panggil(t, legal, http.MethodPost, alamatPenandatangan, isi("Rina", "Direktur", rp(1), false)), http.StatusForbidden, "LEGAL menambah penandatangan")
	harapKode(t, panggil(t, legal, http.MethodPut, alamatPenandatangan+"/"+budi.ID, isi("Budi", "Direktur", rp(1), false)), http.StatusForbidden, "LEGAL mengubah penandatangan")
	harapKode(t, panggil(t, user, http.MethodGet, "/api/v1/legal/agreement-signatories", nil), http.StatusForbidden, "USER membaca Daftar Penandatangan")

	// 4 tambah + 3 ubah + 1 hapus yang berhasil; penolakan tidak tercatat.
	if n := hitungBaris(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'SIGNATORY_CHANGE'") - auditSebelum; n != 8 {
		t.Errorf("%d perubahan Daftar Penandatangan di audit log, harus 8", n)
	}
	var catatan struct{ OldValue, NewValue string }
	dbTes.Raw("SELECT old_value, new_value FROM audit_logs WHERE action = 'SIGNATORY_CHANGE' AND entity_id = ? ORDER BY created_at DESC LIMIT 1", budi.ID).Scan(&catatan)
	if catatan.OldValue != "Budi, Jabatan Budi, sampai Rp 500.000.000" || catatan.NewValue != "Budi, Jabatan Budi, sampai Rp 500.000.000, default" {
		t.Errorf("audit log perubahan Budi: %+v", catatan)
	}
}

// isiDokumenDOCX: teks word/document.xml dari DOCX.
func isiDokumenDOCX(t *testing.T, docx []byte) string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(docx), int64(len(docx)))
	if err != nil {
		t.Fatalf("bukan DOCX: %v", err)
	}
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			return string(b)
		}
	}
	t.Fatal("word/document.xml tidak ada")
	return ""
}

// Penandatangan Pihak Pertama di persetujuan (PR 11): usulan dihitung saat dibaca
// dari nilai kontrak dan mengikuti Daftar terbaru (P2); approver menggantinya dari
// Daftar atau Lainnya (P3), dan penggantian tercatat di audit log; tanpa usulan dan
// tanpa pilihan, persetujuan ditolak. Dokumen final membekukan penandatangannya.
func TestPenandatanganPersetujuan(t *testing.T) {
	butuhMinIO(t)
	admin := masukPeran(t, entity.RoleAdmin)
	legal := masukPeran(t, entity.RoleLegal)
	user := masukPeran(t, entity.RoleUser)
	eksekusi(t, "DELETE FROM agreement_signatories")
	t.Cleanup(func() {
		dbTes.Exec("DELETE FROM agreement_signatories")
		seed.SeedAgreementSignatories(dbTes)
	})
	rp := func(n int64) *int64 { return &n }
	simpan := func(metode, alamat, nama, jabatan string, maks *int64, aktif bool) penandatangan {
		t.Helper()
		w := panggil(t, admin, metode, alamat, map[string]any{"name": nama, "position": jabatan, "max_value": maks, "is_active": aktif})
		if w.Code != http.StatusCreated && w.Code != http.StatusOK {
			t.Fatalf("menyimpan %s: %d %s", nama, w.Code, w.Body)
		}
		var p penandatangan
		isiData(t, w, &p)
		return p
	}
	simpan(http.MethodPost, alamatPenandatangan, "Budi", "Kepala Divisi", rp(500_000_000), true)
	sari := simpan(http.MethodPost, alamatPenandatangan, "Sari", "Direktur", rp(5_000_000_000), true)
	andi := simpan(http.MethodPost, alamatPenandatangan, "Andi", "Direktur Utama", nil, true)

	ajukan := func(nilai int64) perjanjian {
		t.Helper()
		form := formPerjanjian()
		form["nilai_kontrak"] = nilai
		p := buatPerjanjian(t, user, form)
		statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
		return p
	}
	type usulan struct {
		SignatorySuggestion *struct {
			ID       *string `json:"id"`
			Name     string  `json:"name"`
			Position string  `json:"position"`
			Value    *int64  `json:"value"`
		} `json:"signatory_suggestion"`
	}
	usulanUntuk := func(id string) string {
		t.Helper()
		w := panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", id), nil)
		harapKode(t, w, http.StatusOK, "detail approver")
		var u usulan
		isiData(t, w, &u)
		if u.SignatorySuggestion == nil {
			return "tanpa data"
		}
		if u.SignatorySuggestion.Value == nil || *u.SignatorySuggestion.Value != 750_000_000 && *u.SignatorySuggestion.Value != 6_000_000_000 {
			t.Errorf("nilai kontrak usulan %v", u.SignatorySuggestion.Value)
		}
		return u.SignatorySuggestion.Name
	}
	meta := func(id string, isi map[string]any, harap int) {
		t.Helper()
		harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", id, "/meta"), isi), harap, "data persetujuan")
	}
	auditDokumen := func(id string) int64 {
		return hitungBaris(t, "SELECT COUNT(*) FROM audit_logs WHERE action = 'SIGNATORY_CHANGE' AND entity_id = ?", id)
	}

	p := ajukan(750_000_000)
	if got := usulanUntuk(p.ID); got != "Sari" {
		t.Errorf("usulan untuk Rp750 juta %q, harus Sari", got)
	}
	// P2: usulan mengikuti Daftar terbaru sampai disetujui.
	simpan(http.MethodPut, alamatPenandatangan+"/"+sari.ID, "Sari", "Direktur", rp(700_000_000), true)
	if got := usulanUntuk(p.ID); got != "Andi" {
		t.Errorf("usulan sesudah batas Sari diturunkan %q, harus Andi", got)
	}
	simpan(http.MethodPut, alamatPenandatangan+"/"+sari.ID, "Sari", "Direktur", rp(5_000_000_000), true)

	// P3: pilihan dari Daftar. Mengganti usulan tercatat sekali; memilih usulannya sendiri tidak.
	meta(p.ID, map[string]any{"party_one_signatory_id": uuid.NewString()}, http.StatusBadRequest)
	meta(p.ID, map[string]any{"party_one_signatory_id": sari.ID}, http.StatusOK)
	if n := auditDokumen(p.ID); n != 0 {
		t.Errorf("memilih usulan tercatat %d kali di audit log, harus tidak", n)
	}
	meta(p.ID, map[string]any{"party_one_signatory_id": andi.ID}, http.StatusOK)
	meta(p.ID, map[string]any{"party_one_signatory_id": andi.ID}, http.StatusOK)
	if n := auditDokumen(p.ID); n != 1 {
		t.Errorf("penggantian usulan tercatat %d kali di audit log, harus 1", n)
	}
	var catatan struct{ OldValue, NewValue string }
	dbTes.Raw("SELECT old_value, new_value FROM audit_logs WHERE action = 'SIGNATORY_CHANGE' AND entity_id = ?", p.ID).Scan(&catatan)
	if catatan.OldValue != "Sari, Direktur (usulan)" || catatan.NewValue != "Andi, Direktur Utama (dari Daftar)" {
		t.Errorf("audit log penggantian: %+v", catatan)
	}
	// Pergantian penandatangan bersama perubahan data persetujuan lain di satu simpanan: baris
	// SIGNATORY_CHANGE ikut menyebut perubahan lainnya (PR Audit Log tahap L.3b).
	q := ajukan(750_000_000)
	meta(q.ID, map[string]any{"party_one_signatory_id": andi.ID, "agreement_number": "PKS/UJI/L3B"}, http.StatusOK)
	var ketGanti string
	dbTes.Raw("SELECT description FROM audit_logs WHERE action = 'SIGNATORY_CHANGE' AND entity_id = ?", q.ID).Scan(&ketGanti)
	if !strings.Contains(ketGanti, "Nomor perjanjian: (kosong) → PKS/UJI/L3B") {
		t.Errorf("keterangan pergantian penandatangan bersama nomor perjanjian: %q", ketGanti)
	}
	// Kembali ke usulan bukan pergantian (tanpa SIGNATORY_CHANGE), tetapi tetap tercatat di rincian.
	meta(q.ID, map[string]any{"party_one_signatory_id": sari.ID}, http.StatusOK)
	var ketKembali string
	dbTes.Raw("SELECT description FROM audit_logs WHERE action = 'USER_UPDATE' AND entity_id = ? ORDER BY created_at DESC LIMIT 1", q.ID).Scan(&ketKembali)
	if !strings.Contains(ketKembali, "Penandatangan Pihak Pertama: Andi, Direktur Utama (dari Daftar) → Sari, Direktur (") {
		t.Errorf("keterangan kembali ke usulan: %q", ketKembali)
	}

	// Isian approver yang dikirim pengaju tidak tersimpan: saat membuat dibuang, saat
	// revisi yang tersimpan tetap dipakai. Pengaju tidak bisa menentukan penandatangan.
	isiApprover := func(id string) string {
		var v string
		dbTes.Raw("SELECT concat_ws('|', form_data->>'pihak_pertama_penandatangan_id', form_data->>'pihak_pertama_pejabat', form_data->>'tanggal_ttd') FROM agreement_documents WHERE id = ?", id).Scan(&v)
		return v
	}
	susup := func(form map[string]any) map[string]any {
		form["pihak_pertama_penandatangan_id"] = andi.ID
		form["pihak_pertama_pejabat"] = "Penyusup"
		form["tanggal_ttd"] = "2020-01-01"
		return form
	}
	baru := buatPerjanjian(t, user, susup(formPerjanjian()))
	if v := isiApprover(baru.ID); v != "" {
		t.Errorf("isian approver dari pengaju tersimpan saat membuat: %q", v)
	}
	meta(baru.ID, map[string]any{"party_one_signatory_id": "", "party_one_signatory_name": "Rudi", "party_one_signatory_position": "Plt. Direktur", "signing_date": "2026-10-11"}, http.StatusOK)
	harapKode(t, panggil(t, user, http.MethodPut, urlPerjanjian("", "/", baru.ID), map[string]any{"form_data": susup(formPerjanjian())}), http.StatusOK, "revisi pengaju")
	if v := isiApprover(baru.ID); v != "Rudi|2026-10-11" {
		t.Errorf("isian approver sesudah revisi pengaju %q, harus tetap Rudi|2026-10-11", v)
	}
	// Memilih dari Daftar membuang Lainnya, jadi mengosongkan pilihan kembali ke usulan,
	// bukan ke nama Lainnya yang lama.
	meta(baru.ID, map[string]any{"party_one_signatory_id": sari.ID}, http.StatusOK)
	meta(baru.ID, map[string]any{"party_one_signatory_id": ""}, http.StatusOK)
	if v := isiApprover(baru.ID); v != "2026-10-11" {
		t.Errorf("isian approver sesudah pilihan dikosongkan %q, harus hanya tanggal", v)
	}

	// Lainnya: nama dan jabatan diketik approver.
	lainnya := ajukan(750_000_000)
	meta(lainnya.ID, map[string]any{"party_one_signatory_id": "", "party_one_signatory_name": "Rudi", "party_one_signatory_position": "Plt. Direktur"}, http.StatusOK)
	if n := auditDokumen(lainnya.ID); n != 1 {
		t.Errorf("penandatangan Lainnya tercatat %d kali di audit log, harus 1", n)
	}

	// Tanpa usulan (tidak ada yang tak terbatas) dan tanpa pilihan: ditolak sebelum dokumen dibentuk.
	tanpa := ajukan(6_000_000_000)
	simpan(http.MethodPut, alamatPenandatangan+"/"+andi.ID, "Andi", "Direktur Utama", nil, false)
	if got := usulanUntuk(tanpa.ID); got != "" {
		t.Errorf("usulan untuk Rp6 miliar tanpa baris tak terbatas %q, harus kosong", got)
	}
	meta(tanpa.ID, map[string]any{"signing_date": "2026-10-11"}, http.StatusOK)
	w := panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", tanpa.ID, "/status"), map[string]string{"status": "COMPLETED"})
	harapKode(t, w, http.StatusBadRequest, "menyetujui tanpa penandatangan")
	if !strings.Contains(w.Body.String(), "Penandatangan Pihak Pertama belum dipilih") {
		t.Errorf("pesan persetujuan tanpa penandatangan: %s", w.Body)
	}
	// Pilihan yang dinonaktifkan sesudah dipilih juga ditolak, bukan diganti diam-diam.
	meta(p.ID, map[string]any{"signing_date": "2026-10-11"}, http.StatusOK)
	w = panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/status"), map[string]string{"status": "COMPLETED"})
	harapKode(t, w, http.StatusBadRequest, "menyetujui dengan pilihan yang dinonaktifkan")
	if !strings.Contains(w.Body.String(), "sudah tidak aktif") {
		t.Errorf("pesan pilihan nonaktif: %s", w.Body)
	}
	simpan(http.MethodPut, alamatPenandatangan+"/"+andi.ID, "Andi", "Direktur Utama", nil, true)

	t.Run("DokumenFinal", func(t *testing.T) {
		butuhLibreOffice(t)
		meta(lainnya.ID, map[string]any{"signing_date": "2026-10-11"}, http.StatusOK)
		for _, c := range []struct{ id, nama string }{{p.ID, "Andi"}, {lainnya.ID, "Rudi"}} {
			statusPerjanjian(t, legal, c.id, "COMPLETED", "", http.StatusOK)
			var beku string
			dbTes.Raw("SELECT party_one_snapshot->>'signatory_name' FROM agreement_documents WHERE id = ?", c.id).Scan(&beku)
			if beku != c.nama {
				t.Errorf("snapshot penandatangan %q, harus %q", beku, c.nama)
			}
			w := panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", c.id, "/docx"), nil)
			harapKode(t, w, http.StatusOK, "DOCX final")
			if !strings.Contains(isiDokumenDOCX(t, w.Body.Bytes()), c.nama) {
				t.Errorf("DOCX final tanpa nama penandatangan %s", c.nama)
			}
		}
		// Mengubah Daftar sesudah persetujuan tidak mengubah dokumen yang sudah disetujui.
		simpan(http.MethodPut, alamatPenandatangan+"/"+andi.ID, "Andi Baru", "Direktur Utama", nil, true)
		var beku string
		dbTes.Raw("SELECT party_one_snapshot->>'signatory_name' FROM agreement_documents WHERE id = ?", p.ID).Scan(&beku)
		w := panggil(t, legal, http.MethodGet, urlPerjanjian("/legal", "/", p.ID, "/docx"), nil)
		if beku != "Andi" || strings.Contains(isiDokumenDOCX(t, w.Body.Bytes()), "Andi Baru") {
			t.Errorf("dokumen final berubah sesudah Daftar diubah: snapshot %q", beku)
		}
	})
}

// nilaiUtamaDi: kode field bertanda primary_value di definisi isian (JSON).
func nilaiUtamaDi(t *testing.T, raw string) []string {
	t.Helper()
	var bagian []struct {
		Fields []struct {
			Name         string `json:"name"`
			PrimaryValue bool   `json:"primary_value"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(raw), &bagian); err != nil {
		t.Fatalf("definisi isian: %v: %s", err, raw)
	}
	var out []string
	for _, b := range bagian {
		for _, f := range b.Fields {
			if f.PrimaryValue {
				out = append(out, f.Name)
			}
		}
	}
	return out
}

func samaJSON(t *testing.T, a, b string) bool {
	t.Helper()
	var x, y any
	if json.Unmarshal([]byte(a), &x) != nil || json.Unmarshal([]byte(b), &y) != nil {
		t.Fatalf("bukan JSON:\n%s\n%s", a, b)
	}
	return reflect.DeepEqual(x, y)
}

// Migrasi PR 11 (keputusan P1): nilai_kontrak PKS ditandai sebagai nilai utama di
// isian kerja dan setiap versi yang belum bertanda. Hanya tanda itu yang ditambahkan.
// Versi yang sudah bertanda di field lain, atau tanpa nilai_kontrak, dibiarkan.
func TestMigrasiNilaiUtama(t *testing.T) {
	bertanda := fieldPR9(t)
	tanpaTanda := strings.Replace(bertanda, `,"primary_value":true`, "", 1)
	if tanpaTanda == bertanda {
		t.Fatal("tanda nilai utama tidak ditemukan di form-schema.json PR 9")
	}
	terminBertanda := strings.Replace(tanpaTanda, `"name":"termin_1_nilai","label":"Nilai Termin 1","type":"money","required":true`, `"name":"termin_1_nilai","label":"Nilai Termin 1","type":"money","required":true,"primary_value":true`, 1)
	if terminBertanda == tanpaTanda {
		t.Fatal("field termin_1_nilai tidak ditemukan di form-schema.json PR 9")
	}
	tanpaNilai := `[{"title":"Data","fields":[{"name":"nama","label":"Nama","type":"text","required":true}]}]`

	asliKerja := fieldJenis(t, "PKS")
	versi := map[string]string{"TanpaTanda": uuid.NewString(), "BertandaLain": uuid.NewString(), "TanpaNilaiKontrak": uuid.NewString()}
	t.Cleanup(func() {
		for _, id := range versi {
			dbTes.Exec("DELETE FROM agreement_templates WHERE id = ?", id)
		}
		dbTes.Exec("UPDATE agreement_types SET fields = CAST(? AS jsonb) WHERE code = 'PKS'", asliKerja)
	})
	for nama, isi := range map[string]string{"TanpaTanda": tanpaTanda, "BertandaLain": terminBertanda, "TanpaNilaiKontrak": tanpaNilai} {
		eksekusi(t, `INSERT INTO agreement_templates
			(id, created_at, updated_at, code, version, name, file_name, file_path, checksum, status, is_legacy, note, fields)
			VALUES (?, now(), now(), 'PKS', (SELECT COALESCE(MAX(version), 0) + 1 FROM agreement_templates WHERE code = 'PKS'),
				'uji', 'uji.docx', 'uji/tidak-ada.docx', 'nilai-utama-uji', 'ARCHIVED', false, 'uji', CAST(? AS jsonb))`, versi[nama], isi)
	}
	eksekusi(t, "UPDATE agreement_types SET fields = CAST(? AS jsonb) WHERE code = 'PKS'", tanpaTanda)

	for i := 1; i <= 2; i++ {
		if err := seed.TandaiNilaiUtamaPKS(dbTes); err != nil {
			t.Fatalf("migrasi putaran %d: %v", i, err)
		}
	}

	for _, c := range []struct {
		nama, raw string
		harap     []string
	}{
		{"IsianKerja", fieldJenis(t, "PKS"), []string{"nilai_kontrak"}},
		{"TanpaTanda", fieldTemplate(t, versi["TanpaTanda"]), []string{"nilai_kontrak"}},
		{"BertandaLain", fieldTemplate(t, versi["BertandaLain"]), []string{"termin_1_nilai"}},
		{"TanpaNilaiKontrak", fieldTemplate(t, versi["TanpaNilaiKontrak"]), nil},
	} {
		if got := nilaiUtamaDi(t, c.raw); !reflect.DeepEqual(got, c.harap) {
			t.Errorf("%s: nilai utama %v, harus %v", c.nama, got, c.harap)
		}
	}
	// Selain tanda, definisinya sama persis.
	if raw := fieldTemplate(t, versi["TanpaTanda"]); !samaJSON(t, raw, bertanda) {
		t.Errorf("isi definisi berubah selain tanda nilai utama:\n%s", raw)
	}
	if raw := fieldTemplate(t, versi["TanpaNilaiKontrak"]); !samaJSON(t, raw, tanpaNilai) {
		t.Errorf("definisi tanpa nilai_kontrak berubah:\n%s", raw)
	}
}
