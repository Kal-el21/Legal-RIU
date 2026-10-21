package router

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

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
