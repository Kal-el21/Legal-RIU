package service

import (
	"strings"

	"legal-riu-portal/internal/entity"
)

// pilihPenandatangan memilih penandatangan Pihak Pertama untuk nilai kontrak:
// yang aktif dengan batas atas terkecil yang masih lebih besar atau sama dengan
// nilai, atau baris tak terbatas kalau nilai melewati semua batas. Tanpa nilai
// (jenis dokumen tanpa nilai utama) dipakai yang bertanda default. nil berarti
// tidak ada yang cocok, dan approver memilih sendiri.
func pilihPenandatangan(daftar []entity.AgreementSignatory, nilai *int64) *entity.AgreementSignatory {
	var terpilih *entity.AgreementSignatory
	for i := range daftar {
		p := &daftar[i]
		if !p.IsActive {
			continue
		}
		if nilai == nil {
			if p.IsDefault {
				return p
			}
			continue
		}
		switch {
		case p.MaxValue == nil:
			if terpilih == nil {
				terpilih = p
			}
		case *p.MaxValue >= *nilai && (terpilih == nil || terpilih.MaxValue == nil || *p.MaxValue < *terpilih.MaxValue):
			terpilih = p
		}
	}
	return terpilih
}

// nilaiUtama: isian field bertanda nilai utama (nilai kontrak, keputusan P1), atau
// nil kalau jenis dokumennya tidak punya field itu atau isiannya kosong.
func nilaiUtama(sections []AgreementSectionDefinition, data map[string]interface{}) *int64 {
	for _, s := range sections {
		for _, f := range s.Fields {
			if !f.PrimaryValue {
				continue
			}
			if strings.TrimSpace(valueString(data[f.Name])) == "" {
				return nil
			}
			n := valueInt64(data[f.Name])
			return &n
		}
	}
	return nil
}
