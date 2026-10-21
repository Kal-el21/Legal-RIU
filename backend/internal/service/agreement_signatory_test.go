package service

import (
	"fmt"
	"testing"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
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

// Penandatangan yang dipakai dokumen: pilihan approver dari Daftar, lalu Lainnya
// (nama dan jabatan diketik approver), lalu usulan menurut nilai utama (P2, P3).
// Pilihan yang sudah tidak aktif ditolak, supaya dokumen tidak diam-diam
// ditandatangani orang lain.
func TestPenandatanganDipakai(t *testing.T) {
	batas := func(n int64) *int64 { return &n }
	budi := entity.AgreementSignatory{Name: "Budi", Position: "Kepala Divisi", MaxValue: batas(500_000_000), IsActive: true}
	sari := entity.AgreementSignatory{Name: "Sari", Position: "Direktur", MaxValue: batas(5_000_000_000), IsActive: true}
	andi := entity.AgreementSignatory{Name: "Andi", Position: "Direktur Utama", IsDefault: true, IsActive: true}
	lama := entity.AgreementSignatory{Name: "Lama", Position: "Direktur", MaxValue: batas(900_000_000), IsActive: false}
	for i, p := range []*entity.AgreementSignatory{&budi, &sari, &andi, &lama} {
		p.ID = uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-00000000000%d", i+1))
	}
	daftar := []entity.AgreementSignatory{budi, sari, andi, lama}
	sections := []AgreementSectionDefinition{{Title: "Pembayaran", Fields: []AgreementFieldDefinition{{Name: "nilai_kontrak", Label: "Nilai", Type: "money", PrimaryValue: true}}}}
	isi := func(tambahan map[string]interface{}) map[string]interface{} {
		data := map[string]interface{}{"nilai_kontrak": "750000000"}
		for k, v := range tambahan {
			data[k] = v
		}
		return data
	}
	for _, c := range []struct {
		nama   string
		daftar []entity.AgreementSignatory
		data   map[string]interface{}
		harap  PenandatanganDipakai
		galat  bool
	}{
		{"Usulan", daftar, isi(nil), PenandatanganDipakai{ID: sari.ID.String(), Nama: "Sari", Jabatan: "Direktur", Sumber: "usulan"}, false},
		{"PilihanDariDaftar", daftar, isi(map[string]interface{}{"pihak_pertama_penandatangan_id": andi.ID.String()}), PenandatanganDipakai{ID: andi.ID.String(), Nama: "Andi", Jabatan: "Direktur Utama", Sumber: "daftar"}, false},
		{"PilihanMengalahkanLainnya", daftar, isi(map[string]interface{}{"pihak_pertama_penandatangan_id": budi.ID.String(), "pihak_pertama_pejabat": "Rudi"}), PenandatanganDipakai{ID: budi.ID.String(), Nama: "Budi", Jabatan: "Kepala Divisi", Sumber: "daftar"}, false},
		{"Lainnya", daftar, isi(map[string]interface{}{"pihak_pertama_pejabat": " Rudi ", "pihak_pertama_jabatan": "Plt. Direktur"}), PenandatanganDipakai{Nama: "Rudi", Jabatan: "Plt. Direktur", Sumber: "lainnya"}, false},
		{"LainnyaKosongPakaiUsulan", daftar, isi(map[string]interface{}{"pihak_pertama_pejabat": " ", "pihak_pertama_penandatangan_id": ""}), PenandatanganDipakai{ID: sari.ID.String(), Nama: "Sari", Jabatan: "Direktur", Sumber: "usulan"}, false},
		{"PilihanNonaktif", daftar, isi(map[string]interface{}{"pihak_pertama_penandatangan_id": lama.ID.String()}), PenandatanganDipakai{}, true},
		{"PilihanTerhapus", daftar, isi(map[string]interface{}{"pihak_pertama_penandatangan_id": uuid.NewString()}), PenandatanganDipakai{}, true},
		{"TidakAdaUsulan", []entity.AgreementSignatory{budi, sari}, isi(map[string]interface{}{"nilai_kontrak": "6000000000"}), PenandatanganDipakai{}, false},
	} {
		t.Run(c.nama, func(t *testing.T) {
			got, err := penandatanganDipakai(c.daftar, sections, c.data)
			if (err != nil) != c.galat {
				t.Fatalf("galat %v, harus galat %v", err, c.galat)
			}
			if got != c.harap {
				t.Errorf("dipakai %+v, harus %+v", got, c.harap)
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
