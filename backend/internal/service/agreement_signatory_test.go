package service

import (
	"testing"

	"legal-riu-portal/internal/entity"
)

// Penandatangan dipilih menurut nilai kontrak: batas atas terkecil yang masih lebih
// besar atau sama dengan nilai (batas tepat jatuh ke baris itu), lalu baris tak
// terbatas. Tanpa nilai utama dipakai yang bertanda default; tidak ada yang cocok
// berarti approver memilih sendiri (keputusan K9, PR 11).
func TestPilihPenandatangan(t *testing.T) {
	batas := func(n int64) *int64 { return &n }
	nilai := batas
	budi := entity.AgreementSignatory{Name: "Budi", Position: "Kepala Divisi", MaxValue: batas(500_000_000), IsActive: true}
	sari := entity.AgreementSignatory{Name: "Sari", Position: "Direktur", MaxValue: batas(5_000_000_000), IsActive: true}
	andi := entity.AgreementSignatory{Name: "Andi", Position: "Direktur Utama", IsDefault: true, IsActive: true}
	nonaktif := entity.AgreementSignatory{Name: "Lama", Position: "Direktur", MaxValue: batas(1_000_000_000), IsActive: false}
	// Urutan daftar tidak menentukan pilihan.
	lengkap := []entity.AgreementSignatory{andi, nonaktif, sari, budi}
	tanpaTakTerbatas := []entity.AgreementSignatory{sari, budi}

	for _, c := range []struct {
		nama   string
		daftar []entity.AgreementSignatory
		nilai  *int64
		harap  string
	}{
		{"NolRupiah", lengkap, nilai(0), "Budi"},
		{"TepatDiBatas", lengkap, nilai(500_000_000), "Budi"},
		{"SatuRupiahDiAtasBatas", lengkap, nilai(500_000_001), "Sari"},
		{"YangNonaktifDilewati", lengkap, nilai(750_000_000), "Sari"},
		{"DiAtasSemuaBatas", lengkap, nilai(6_000_000_000), "Andi"},
		{"DiAtasSemuaBatasTanpaTakTerbatas", tanpaTakTerbatas, nilai(6_000_000_000), ""},
		{"TanpaNilaiUtamaPakaiDefault", lengkap, nil, "Andi"},
		{"TanpaNilaiUtamaTanpaDefault", tanpaTakTerbatas, nil, ""},
		{"DaftarKosong", nil, nilai(1), ""},
	} {
		t.Run(c.nama, func(t *testing.T) {
			got := ""
			if p := pilihPenandatangan(c.daftar, c.nilai); p != nil {
				got = p.Name
			}
			if got != c.harap {
				t.Errorf("terpilih %q, harus %q", got, c.harap)
			}
		})
	}
}

// Nilai utama dibaca dari field uang bertanda primary_value (keputusan P1).
func TestNilaiUtama(t *testing.T) {
	bagian := func(tanda bool) []AgreementSectionDefinition {
		return []AgreementSectionDefinition{{Title: "Pembayaran", Fields: []AgreementFieldDefinition{
			{Name: "nama", Label: "Nama", Type: "text"},
			{Name: "nilai_kontrak", Label: "Nilai Kontrak", Type: "money", PrimaryValue: tanda},
		}}}
	}
	baca := func(sections []AgreementSectionDefinition, data map[string]interface{}) string {
		if n := nilaiUtama(sections, data); n != nil {
			return valueString(float64(*n))
		}
		return "kosong"
	}
	for _, c := range []struct {
		nama     string
		sections []AgreementSectionDefinition
		data     map[string]interface{}
		harap    string
	}{
		{"Bertanda", bagian(true), isianJSON(t, `{"nilai_kontrak": "Rp 750.000.000"}`), "750000000"},
		{"BertandaAngka", bagian(true), isianJSON(t, `{"nilai_kontrak": 1500}`), "1500"},
		{"IsianKosong", bagian(true), isianJSON(t, `{"nilai_kontrak": " "}`), "kosong"},
		{"TanpaTanda", bagian(false), isianJSON(t, `{"nilai_kontrak": 750000000}`), "kosong"},
	} {
		t.Run(c.nama, func(t *testing.T) {
			if got := baca(c.sections, c.data); got != c.harap {
				t.Errorf("nilai utama %s, harus %s", got, c.harap)
			}
		})
	}
}
