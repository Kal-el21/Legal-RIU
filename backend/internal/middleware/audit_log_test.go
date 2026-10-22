package middleware

import (
	"testing"

	"github.com/google/uuid"
)

type idTeks string

func (i idTeks) String() string { return string(i) }

// Semua bentuk ID yang dikirim handler lewat SetAuditContext terbaca. Sebelum PR A tahap A.1b
// hanya string yang terbaca, sehingga ID bertipe uuid.UUID diam-diam tidak tercatat.
func TestIdEntitas(t *testing.T) {
	id := uuid.MustParse("3f2a91c0-1111-4222-8333-444455556666")
	kasus := []struct {
		nama  string
		nilai any
		harap uuid.UUID
	}{
		{"string", id.String(), id},
		{"uuid.UUID", id, id},
		{"*uuid.UUID", &id, id},
		{"Stringer", idTeks(id.String()), id},
		{"teks bukan UUID (impor)", "import", uuid.Nil},
		{"pointer kosong", (*uuid.UUID)(nil), uuid.Nil},
		{"tidak dipasang", nil, uuid.Nil},
		{"tipe lain", 42, uuid.Nil},
	}
	for _, k := range kasus {
		if got := idEntitas(k.nilai); got != k.harap {
			t.Errorf("%s: idEntitas(%v) = %v, harap %v", k.nama, k.nilai, got, k.harap)
		}
	}
}
