package service

import (
	"fmt"
	"math"
)

// BarisTermin: satu termin hasil pembagian nilai kontrak.
type BarisTermin struct {
	Persen float64
	Nilai  int64
}

// bagiTermin membagi nilai kontrak ke termin menurut persennya (keputusan K5 dan D5).
// Persen dibulatkan ke dua desimal; persen termin terakhir diabaikan dan diganti
// sisanya. Nilai tiap termin dibulatkan ke rupiah, dan termin terakhir menyerap
// selisih pembulatan sehingga jumlahnya selalu tepat sama dengan nilai kontrak.
// Hitungan memakai basis poin (1% = 100) supaya tidak ada galat pecahan.
func bagiTermin(kontrak int64, persen []float64) ([]BarisTermin, error) {
	if len(persen) == 0 {
		return nil, fmt.Errorf("minimal satu termin")
	}
	if kontrak <= 0 || kontrak > math.MaxInt64/10000 {
		return nil, fmt.Errorf("nilai kontrak harus lebih dari 0")
	}
	hasil := make([]BarisTermin, len(persen))
	var totalBP, totalNilai int64
	for i, p := range persen[:len(persen)-1] {
		// Dibulatkan dua kali supaya 1,005 menjadi 1,01 walaupun float64 menyimpannya
		// sebagai 1,00499...; cara yang sama dipakai pratinjau di browser.
		bp := int64(math.Round(math.Round(p*1e6) / 1e4))
		if bp <= 0 {
			return nil, fmt.Errorf("persen termin %d harus lebih dari 0", i+1)
		}
		totalBP += bp
		if totalBP > 10000 {
			return nil, fmt.Errorf("jumlah persen termin melebihi 100%%")
		}
		nilai := (kontrak*bp + 5000) / 10000
		hasil[i] = BarisTermin{float64(bp) / 100, nilai}
		totalNilai += nilai
	}
	if totalBP == 10000 {
		return nil, fmt.Errorf("persen sudah 100%%, tidak ada sisa untuk termin terakhir")
	}
	hasil[len(hasil)-1] = BarisTermin{float64(10000-totalBP) / 100, kontrak - totalNilai}
	return hasil, nil
}

// urutanTerbilang: 1 menjadi "pertama", 12 menjadi "kedua belas".
func urutanTerbilang(n int) string {
	if n == 1 {
		return "pertama"
	}
	return "ke" + terbilang(int64(n))
}

// jumlahTerbilang: 3 menjadi "3 (tiga)".
func jumlahTerbilang(n int) string {
	return fmt.Sprintf("%d (%s)", n, terbilang(int64(n)))
}
