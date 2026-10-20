package service

import (
	"strings"
	"testing"
)

// Pembagian nilai kontrak ke termin (PR 10, keputusan K5 dan D5): persen termin
// terakhir adalah sisanya, nilai tiap termin dibulatkan ke rupiah, dan termin
// terakhir menyerap selisih pembulatan sehingga jumlahnya tepat sama dengan nilai
// kontrak. Kasus yang sama dipakai tes Vitest di frontend (tahap 10.4).
func TestPembagianTermin(t *testing.T) {
	for _, c := range []struct {
		nama    string
		kontrak int64
		persen  []float64
		harap   []BarisTermin
		galat   string
	}{
		{"SatuTermin", 750_000_000, []float64{40}, []BarisTermin{{100, 750_000_000}}, ""},
		{"TigaTerminBulat", 750_000_000, []float64{30, 30, 0}, []BarisTermin{{30, 225_000_000}, {30, 225_000_000}, {40, 300_000_000}}, ""},
		{"SepertigaDariSatuMiliar", 1_000_000_000, []float64{33.33, 33.33, 0}, []BarisTermin{{33.33, 333_300_000}, {33.33, 333_300_000}, {33.34, 333_400_000}}, ""},
		{"SelisihPembulatanKeTerminTerakhir", 1_000_000_001, []float64{33.33, 33.33, 0}, []BarisTermin{{33.33, 333_300_000}, {33.33, 333_300_000}, {33.34, 333_400_001}}, ""},
		{"NilaiKecilDibulatkan", 100, []float64{33.33, 33.33, 0}, []BarisTermin{{33.33, 33}, {33.33, 33}, {33.34, 34}}, ""},
		{"NilaiDibulatkanKeAtas", 1_000_005, []float64{33.33, 0}, []BarisTermin{{33.33, 333_302}, {66.67, 666_703}}, ""},
		{"PersenLebihDariDuaDesimal", 1_000_000, []float64{12.345, 0}, []BarisTermin{{12.35, 123_500}, {87.65, 876_500}}, ""},
		// float64 menyimpan 1,005 sebagai 1,00499...; tetap harus menjadi 1,01.
		{"PersenSetengahDibulatkanKeAtas", 1_000_000, []float64{1.005, 0}, []BarisTermin{{1.01, 10_100}, {98.99, 989_900}}, ""},
		{"PersenTerakhirDiabaikan", 1_000_000, []float64{25, 99}, []BarisTermin{{25, 250_000}, {75, 750_000}}, ""},
		{"NilaiTriliunan", 3_000_000_005_003, []float64{33.33, 0}, []BarisTermin{{33.33, 999_900_001_667}, {66.67, 2_000_100_003_336}}, ""},
		{"TanpaTermin", 1_000_000, nil, nil, "minimal satu termin"},
		{"PersenMelebihi100", 1_000_000, []float64{60, 50, 0}, nil, "melebihi 100"},
		{"PersenTepat100TanpaSisa", 1_000_000, []float64{50, 50, 0}, nil, "termin terakhir"},
		{"PersenNol", 1_000_000, []float64{0, 0}, nil, "termin 1"},
		{"PersenNegatif", 1_000_000, []float64{-10, 0}, nil, "termin 1"},
		{"NilaiKontrakNol", 0, []float64{50, 0}, nil, "nilai kontrak"},
	} {
		t.Run(c.nama, func(t *testing.T) {
			hasil, err := bagiTermin(c.kontrak, c.persen)
			if c.galat != "" {
				if err == nil || !strings.Contains(err.Error(), c.galat) {
					t.Fatalf("galat %v, harus menyebut %q", err, c.galat)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(hasil) != len(c.harap) {
				t.Fatalf("%d termin, harus %d: %+v", len(hasil), len(c.harap), hasil)
			}
			var jumlah int64
			for i := range hasil {
				if hasil[i] != c.harap[i] {
					t.Errorf("termin %d = %+v, harus %+v", i+1, hasil[i], c.harap[i])
				}
				jumlah += hasil[i].Nilai
			}
			if jumlah != c.kontrak {
				t.Errorf("jumlah nilai termin %d, harus sama dengan nilai kontrak %d", jumlah, c.kontrak)
			}
		})
	}
}

func TestUrutanTerbilang(t *testing.T) {
	for n, harap := range map[int]string{1: "pertama", 2: "kedua", 3: "ketiga", 10: "kesepuluh", 11: "kesebelas", 12: "kedua belas", 21: "kedua puluh satu"} {
		if v := urutanTerbilang(n); v != harap {
			t.Errorf("urutan %d = %q, harus %q", n, v, harap)
		}
	}
	for n, harap := range map[int]string{1: "1 (satu)", 3: "3 (tiga)", 12: "12 (dua belas)"} {
		if v := jumlahTerbilang(n); v != harap {
			t.Errorf("jumlah %d = %q, harus %q", n, v, harap)
		}
	}
}
