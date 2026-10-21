package service

import (
	"errors"
	"fmt"
	"strings"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
)

// RingkasPenandatangan: satu baris Daftar Penandatangan untuk audit log, mis.
// "Budi, Kepala Divisi, sampai Rp 500.000.000, default".
func RingkasPenandatangan(p *entity.AgreementSignatory) string {
	batas := "tak terbatas"
	if p.MaxValue != nil {
		batas = "sampai " + rupiah(*p.MaxValue)
	}
	out := p.Name + ", " + p.Position + ", " + batas
	if p.IsDefault {
		out += ", default"
	}
	if !p.IsActive {
		out += ", nonaktif"
	}
	return out
}

// ErrPenandatanganTidakAda: penandatangan yang diminta tidak ada (atau sudah dihapus).
var ErrPenandatanganTidakAda = errors.New("penandatangan tidak ditemukan")

func (s *agreementDocumentService) ListSignatories() ([]entity.AgreementSignatory, error) {
	return s.repo.ListSignatories()
}

// CreateSignatory menambah penandatangan; yang baru selalu aktif.
func (s *agreementDocumentService) CreateSignatory(req dto.AgreementSignatoryRequest) (*entity.AgreementSignatory, error) {
	p := &entity.AgreementSignatory{IsActive: true}
	if e := s.isiPenandatangan(p, req); e != nil {
		return nil, e
	}
	return p, s.repo.SaveSignatory(p)
}

// UpdateSignatory mengubah penandatangan; mengembalikan juga keadaan sebelumnya
// untuk audit log.
func (s *agreementDocumentService) UpdateSignatory(id string, req dto.AgreementSignatoryRequest) (lama, baru *entity.AgreementSignatory, e error) {
	p, e := s.cariPenandatangan(id)
	if e != nil {
		return nil, nil, e
	}
	sebelum := *p
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}
	if e = s.isiPenandatangan(p, req); e != nil {
		return nil, nil, e
	}
	return &sebelum, p, s.repo.SaveSignatory(p)
}

func (s *agreementDocumentService) DeleteSignatory(id string) (*entity.AgreementSignatory, error) {
	p, e := s.cariPenandatangan(id)
	if e != nil {
		return nil, e
	}
	return p, s.repo.DeleteSignatory(p.ID)
}

func (s *agreementDocumentService) cariPenandatangan(id string) (*entity.AgreementSignatory, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, ErrPenandatanganTidakAda
	}
	p, e := s.repo.FindSignatory(uid)
	if e != nil {
		return nil, ErrPenandatanganTidakAda
	}
	return p, nil
}

// isiPenandatangan memeriksa isian lalu mengisinya ke p. Di antara penandatangan
// aktif, batas nilai maksimum tidak boleh kembar dan hanya satu yang tak terbatas,
// supaya setiap nilai kontrak jatuh ke tepat satu penandatangan. Yang nonaktif tidak
// dihitung, jadi penggantinya boleh memakai batas yang sama.
func (s *agreementDocumentService) isiPenandatangan(p *entity.AgreementSignatory, req dto.AgreementSignatoryRequest) error {
	nama, jabatan := strings.TrimSpace(req.Name), strings.TrimSpace(req.Position)
	switch {
	case nama == "":
		return errors.New("Nama penandatangan wajib diisi")
	case jabatan == "":
		return errors.New("Jabatan penandatangan wajib diisi")
	case req.MaxValue != nil && *req.MaxValue <= 0:
		return errors.New("Nilai maksimum harus lebih dari 0, atau dikosongkan untuk tak terbatas")
	}
	if p.IsActive {
		semua, e := s.repo.ListSignatories()
		if e != nil {
			return e
		}
		for _, lain := range semua {
			if lain.ID == p.ID || !lain.IsActive {
				continue
			}
			switch {
			case req.MaxValue == nil && lain.MaxValue == nil:
				return fmt.Errorf("Batas tak terbatas sudah dipakai %s (%s); hanya satu penandatangan yang boleh tak terbatas", lain.Name, lain.Position)
			case req.MaxValue != nil && lain.MaxValue != nil && *req.MaxValue == *lain.MaxValue:
				return fmt.Errorf("Nilai maksimum %s sudah dipakai %s (%s)", rupiah(*req.MaxValue), lain.Name, lain.Position)
			}
		}
	}
	p.Name, p.Position, p.MaxValue, p.IsDefault = nama, jabatan, req.MaxValue, req.IsDefault
	return nil
}

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
