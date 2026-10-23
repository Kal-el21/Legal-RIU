package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// denganNomorTiket mengulang hanya kalau nomor tiket bentrok, maksimal 3 kali (PR A tahap A.3).
func TestDenganNomorTiket(t *testing.T) {
	bentrok := &pgconn.PgError{Code: "23505", ConstraintName: "idx_legal_opinions_ticket_number"}
	nomorPerjanjian := &pgconn.PgError{Code: "23505", ConstraintName: "idx_agreement_documents_agreement_number"}
	lain := errors.New("database mati")

	kasus := []struct {
		nama       string
		hasil      []error // jawaban simpan per percobaan
		harapCoba  int
		harapGalat error
	}{
		{"langsung berhasil", []error{nil}, 1, nil},
		{"bentrok sekali lalu berhasil", []error{bentrok, nil}, 2, nil},
		{"bentrok dua kali lalu berhasil", []error{bentrok, bentrok, nil}, 3, nil},
		{"bentrok terus: berhenti di 3", []error{bentrok, bentrok, bentrok, nil}, 3, bentrok},
		{"galat lain tidak diulang", []error{lain, nil}, 1, lain},
		{"nomor perjanjian kembar tidak diulang", []error{nomorPerjanjian, nil}, 1, nomorPerjanjian},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			urut, coba := 6, 0
			var dipakai []string
			err := denganNomorTiket("LO", func(prefix string) (int, error) {
				urut++ // pengajuan lain mengambil nomor di antara percobaan
				return urut, nil
			}, func(nomor string) error {
				dipakai = append(dipakai, nomor)
				coba++
				return k.hasil[coba-1]
			})
			if coba != k.harapCoba || !errors.Is(err, k.harapGalat) {
				t.Fatalf("percobaan %d, galat %v; harap %d percobaan, galat %v", coba, err, k.harapCoba, k.harapGalat)
			}
			for i, nomor := range dipakai {
				if !strings.HasPrefix(nomor, "LO-") || !strings.HasSuffix(nomor, []string{"-0007", "-0008", "-0009"}[i]) {
					t.Errorf("percobaan %d memakai nomor %s; harap nomor baru tiap percobaan", i+1, nomor)
				}
			}
		})
	}

	t.Run("nomor berikutnya gagal diambil", func(t *testing.T) {
		simpan := 0
		err := denganNomorTiket("LO", func(string) (int, error) { return 0, lain }, func(string) error { simpan++; return nil })
		if !errors.Is(err, lain) || simpan != 0 {
			t.Errorf("galat %v, simpan dipanggil %d kali; harap galat dikembalikan tanpa simpan", err, simpan)
		}
	})
}
