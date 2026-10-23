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

// Kunci isian pengajuan yang ditulis approver untuk penandatangan Pihak Pertama:
// pilihan dari Daftar, atau Lainnya (nama dan jabatan diketik sendiri).
const (
	isianPenandatanganID = "pihak_pertama_penandatangan_id"
	isianPejabat         = "pihak_pertama_pejabat"
	isianJabatan         = "pihak_pertama_jabatan"
)

// ErrPenandatanganTidakAktif: penandatangan pilihan approver sudah dihapus atau
// dinonaktifkan sesudah dipilih.
var ErrPenandatanganTidakAktif = errors.New("Penandatangan yang dipilih sudah tidak aktif di Daftar Penandatangan; pilih ulang di data persetujuan")

// PenandatanganDipakai: penandatangan Pihak Pertama untuk satu dokumen. Sumber:
// "daftar" (pilihan approver), "lainnya" (diketik approver), "usulan" (menurut nilai
// utama); kosong kalau belum ada yang bisa dipakai. ID kosong untuk Lainnya.
type PenandatanganDipakai struct {
	ID, Nama, Jabatan, Sumber string
}

func (p PenandatanganDipakai) String() string {
	if p.Nama == "" {
		return "tidak ada"
	}
	return p.Nama + ", " + p.Jabatan + " (" + map[string]string{"daftar": "dari Daftar", "lainnya": "Lainnya", "usulan": "usulan"}[p.Sumber] + ")"
}

// isianApprover: isian pengajuan yang hanya ditulis approver lewat data persetujuan
// (UpdateMeta). tempat_ttd tidak termasuk, karena juga usulan pengaju.
var isianApprover = []string{"tanggal_ttd", isianPenandatanganID, isianPejabat, isianJabatan, "pihak_pertama_pic", "pihak_pertama_telepon", "pihak_pertama_email"}

// pertahankanIsianApprover membuang isian approver yang dikirim pengaju dan
// memakai yang tersimpan (lama, nil untuk pengajuan baru), supaya pengaju tidak bisa
// menentukan penandatangan atau tanggal tanda tangan sendiri.
func pertahankanIsianApprover(baru, lama map[string]interface{}) {
	if baru == nil {
		return
	}
	for _, k := range isianApprover {
		delete(baru, k)
		if v, ada := lama[k]; ada {
			baru[k] = v
		}
	}
}

// PergantianPenandatangan: approver mengganti penandatangan usulan, untuk audit log.
type PergantianPenandatangan struct{ Usulan, Pilihan string }

// pilihPenandatanganApprover menerapkan pilihan penandatangan dari data persetujuan
// ke isian (P3): penandatangan dari Daftar (harus aktif), atau Lainnya dengan nama dan
// jabatan yang diketik. Pergantian dikembalikan hanya kalau pilihannya berubah dan
// bukan usulan, supaya yang tercatat di audit log hanya penggantian usulan. Penandatangan yang
// berubah tanpa dianggap pergantian (kembali ke usulan) dikembalikan sebagai rincian audit
// "Penandatangan Pihak Pertama: lama → baru" (PR Audit Log tahap L.3b).
func (s *agreementDocumentService) pilihPenandatanganApprover(d *entity.AgreementDocument, data map[string]interface{}, req dto.AgreementMetaRequest) (*PergantianPenandatangan, string, error) {
	sections, e := s.sectionsFor(d)
	if e != nil {
		return nil, "", e
	}
	daftar, e := s.repo.ListSignatories()
	if e != nil {
		return nil, "", e
	}
	sebelum, _ := penandatanganDipakai(daftar, sections, data)
	id := ""
	if req.PartyOneSignatoryID != nil {
		id = strings.TrimSpace(*req.PartyOneSignatoryID)
	}
	if id != "" {
		aktif := false
		for _, p := range daftar {
			aktif = aktif || (p.ID.String() == id && p.IsActive)
		}
		if !aktif {
			return nil, "", errors.New("Penandatangan tidak ditemukan atau nonaktif di Daftar Penandatangan")
		}
		data[isianPenandatanganID] = id
		delete(data, isianPejabat)
		delete(data, isianJabatan)
	} else {
		if req.PartyOneSignatoryID != nil {
			delete(data, isianPenandatanganID)
		}
		setOptional(data, isianPejabat, req.PartyOneSignatoryName)
		setOptional(data, isianJabatan, req.PartyOneSignatoryPosition)
	}
	sesudah, _ := penandatanganDipakai(daftar, sections, data)
	if sesudah == sebelum {
		return nil, "", nil
	}
	rincian := "Penandatangan Pihak Pertama: " + sebelum.String() + " → " + sesudah.String()
	if sesudah.Sumber == "usulan" {
		return nil, rincian, nil
	}
	usulan := PenandatanganDipakai{}
	if p := usulanPenandatangan(daftar, sections, data); p != nil {
		usulan = PenandatanganDipakai{ID: p.ID.String(), Nama: p.Name, Jabatan: p.Position, Sumber: "usulan"}
	}
	if sesudah.Sumber == "daftar" && sesudah.ID == usulan.ID {
		return nil, rincian, nil
	}
	return &PergantianPenandatangan{Usulan: usulan.String(), Pilihan: sesudah.String()}, "", nil
}

// pihakPertama: data Pihak Pertama untuk dokumen (partyOneSnapshot), dengan
// penandatangan pilihan approver atau usulan dari Daftar Penandatangan.
func (s *agreementDocumentService) pihakPertama(sections []AgreementSectionDefinition, data map[string]interface{}) (map[string]interface{}, PenandatanganDipakai, error) {
	master, e := s.repo.GetActiveMaster()
	if e != nil {
		return nil, PenandatanganDipakai{}, errors.New("Master Pihak Pertama belum tersedia")
	}
	daftar, e := s.repo.ListSignatories()
	if e != nil {
		return nil, PenandatanganDipakai{}, e
	}
	p, e := penandatanganDipakai(daftar, sections, data)
	return partyOneSnapshot(master, data, p), p, e
}

// usulanUntuk: usulan penandatangan untuk halaman detail approver; nil kalau
// definisi isian atau Daftar tidak terbaca.
func (s *agreementDocumentService) usulanUntuk(d *entity.AgreementDocument) *entity.UsulanPenandatangan {
	sections, e := s.sectionsFor(d)
	if e != nil {
		return nil
	}
	data, e := decodeForm(d.FormData)
	if e != nil {
		return nil
	}
	daftar, e := s.repo.ListSignatories()
	if e != nil {
		return nil
	}
	u := &entity.UsulanPenandatangan{Value: nilaiUtama(sections, data)}
	if p := pilihPenandatangan(daftar, u.Value); p != nil {
		u.ID, u.Name, u.Position = &p.ID, p.Name, p.Position
	}
	return u
}

// usulanPenandatangan: penandatangan menurut nilai utama pengajuan (P2: dihitung
// setiap kali dibaca, jadi mengikuti Daftar terbaru sampai disetujui).
func usulanPenandatangan(daftar []entity.AgreementSignatory, sections []AgreementSectionDefinition, data map[string]interface{}) *entity.AgreementSignatory {
	return pilihPenandatangan(daftar, nilaiUtama(sections, data))
}

// penandatanganDipakai: pilihan approver dari Daftar, lalu Lainnya, lalu usulan
// (P3). Pilihan yang sudah tidak aktif ditolak, supaya dokumen tidak diam-diam
// ditandatangani orang lain.
func penandatanganDipakai(daftar []entity.AgreementSignatory, sections []AgreementSectionDefinition, data map[string]interface{}) (PenandatanganDipakai, error) {
	if id := strings.TrimSpace(valueString(data[isianPenandatanganID])); id != "" {
		for _, p := range daftar {
			if p.ID.String() == id && p.IsActive {
				return PenandatanganDipakai{ID: id, Nama: p.Name, Jabatan: p.Position, Sumber: "daftar"}, nil
			}
		}
		return PenandatanganDipakai{}, ErrPenandatanganTidakAktif
	}
	if nama := strings.TrimSpace(valueString(data[isianPejabat])); nama != "" {
		return PenandatanganDipakai{Nama: nama, Jabatan: strings.TrimSpace(valueString(data[isianJabatan])), Sumber: "lainnya"}, nil
	}
	if p := usulanPenandatangan(daftar, sections, data); p != nil {
		return PenandatanganDipakai{ID: p.ID.String(), Nama: p.Name, Jabatan: p.Position, Sumber: "usulan"}, nil
	}
	return PenandatanganDipakai{}, nil
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
