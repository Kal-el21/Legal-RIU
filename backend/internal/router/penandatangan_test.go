package router

import (
	"encoding/json"
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
