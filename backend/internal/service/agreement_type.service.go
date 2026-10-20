package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
)

// ErrJenisTidakAda: jenis dokumen yang diminta tidak terdaftar.
var ErrJenisTidakAda = errors.New("jenis dokumen tidak ditemukan")

var (
	kodeJenisRE = regexp.MustCompile(`^[A-Z][A-Z0-9_-]{0,49}$`)
	kodeFieldRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,59}$`)
)

var tipeField = map[string]bool{"text": true, "textarea": true, "email": true, "tel": true, "number": true, "money": true, "percent": true, "date": true, "select": true}

// AgreementTypeInput: isian Admin untuk jenis dokumen. Fields nil pada
// perubahan berarti field kerja tidak diubah.
type AgreementTypeInput struct {
	Code     string                       `json:"code"`
	Name     string                       `json:"name"`
	IsActive *bool                        `json:"is_active"`
	Fields   []AgreementSectionDefinition `json:"fields"`
}

// AgreementTypeService mengelola jenis dokumen dan field kerjanya. Field kerja
// baru dipakai pengajuan setelah disalin ke versi template (PR 9 tahap 9.4).
type AgreementTypeService interface {
	List() ([]entity.AgreementType, error)
	Create(AgreementTypeInput) (*entity.AgreementType, error)
	Update(string, AgreementTypeInput) (*entity.AgreementType, error)
}

type agreementTypeService struct {
	repo repository.AgreementTemplateRepository
}

func NewAgreementTypeService(r repository.AgreementTemplateRepository) AgreementTypeService {
	return &agreementTypeService{r}
}

func (s *agreementTypeService) List() ([]entity.AgreementType, error) {
	return s.repo.AllTypes()
}

func (s *agreementTypeService) Create(in AgreementTypeInput) (*entity.AgreementType, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	if !kodeJenisRE.MatchString(code) {
		return nil, errors.New("Kode jenis dokumen hanya boleh huruf, angka, garis bawah, dan tanda hubung, diawali huruf, maksimal 50 karakter")
	}
	if _, e := s.repo.TypeByCode(code); e == nil {
		return nil, fmt.Errorf("kode jenis dokumen %s sudah dipakai", code)
	}
	t := &entity.AgreementType{Code: code, IsActive: true}
	if e := isiJenis(t, in, true); e != nil {
		return nil, e
	}
	if e := s.repo.CreateType(t); e != nil {
		return nil, errors.New("gagal menyimpan jenis dokumen")
	}
	return t, nil
}

func (s *agreementTypeService) Update(code string, in AgreementTypeInput) (*entity.AgreementType, error) {
	t, e := s.repo.TypeByCode(strings.ToUpper(strings.TrimSpace(code)))
	if e != nil {
		return nil, ErrJenisTidakAda
	}
	if e = isiJenis(t, in, in.Fields != nil); e != nil {
		return nil, e
	}
	if in.IsActive != nil {
		t.IsActive = *in.IsActive
	}
	if e = s.repo.SaveType(t); e != nil {
		return nil, errors.New("gagal menyimpan jenis dokumen")
	}
	return t, nil
}

// isiJenis memeriksa nama dan, kalau gantiField, field kerja, lalu mengisinya ke t.
// Tidak ada yang diubah kalau pemeriksaan gagal.
func isiJenis(t *entity.AgreementType, in AgreementTypeInput, gantiField bool) error {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return errors.New("Nama jenis dokumen wajib diisi")
	}
	if gantiField {
		sections, e := rapikanField(in.Fields)
		if e != nil {
			return e
		}
		raw, e := json.Marshal(sections)
		if e != nil {
			return e
		}
		t.Fields = raw
	}
	t.Name = name
	return nil
}

// rapikanField memeriksa definisi field dari Admin dan merapikannya: kode field
// disimpan huruf kecil (di layar dan di Word ditulis huruf besar, keputusan A),
// label dan pilihan tanpa spasi di tepi. Setiap placeholder yang dihasilkan,
// termasuk turunannya, harus unik dan tidak memakai placeholder sistem.
func rapikanField(in []AgreementSectionDefinition) ([]AgreementSectionDefinition, error) {
	pemilik := map[string]string{scopeListPlaceholder[2 : len(scopeListPlaceholder)-2]: ""}
	for key := range systemPlaceholders {
		pemilik[key] = ""
	}
	dipakai := map[string]bool{}
	out := make([]AgreementSectionDefinition, 0, len(in))
	for i, s := range in {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			return nil, fmt.Errorf("Nama bagian ke-%d wajib diisi", i+1)
		}
		bagian := AgreementSectionDefinition{Title: title, Fields: make([]AgreementFieldDefinition, 0, len(s.Fields))}
		for _, f := range s.Fields {
			f.Name = strings.ToLower(strings.TrimSpace(f.Name))
			kode := strings.ToUpper(f.Name)
			if !kodeFieldRE.MatchString(f.Name) {
				return nil, fmt.Errorf("Kode field %s hanya boleh huruf, angka, dan garis bawah, diawali huruf, maksimal 60 karakter", kode)
			}
			if dipakai[f.Name] {
				return nil, fmt.Errorf("Kode field %s dipakai dua kali", kode)
			}
			dipakai[f.Name] = true
			if f.Label = strings.TrimSpace(f.Label); f.Label == "" {
				return nil, fmt.Errorf("Label field %s wajib diisi", kode)
			}
			if !tipeField[f.Type] {
				return nil, fmt.Errorf("Tipe field %s tidak dikenal: %q", kode, f.Type)
			}
			opsi := make([]string, 0, len(f.Options))
			for _, o := range f.Options {
				if o = strings.TrimSpace(o); o != "" {
					opsi = append(opsi, o)
				}
			}
			f.Options = opsi
			if f.Type == "select" && len(f.Options) == 0 {
				return nil, fmt.Errorf("Field %s bertipe pilihan harus punya minimal satu pilihan", kode)
			}
			for token := range fieldPlaceholders(f, contohNilai(f)) {
				lain, ada := pemilik[token]
				switch {
				case ada && lain == "":
					return nil, fmt.Errorf("Kode field %s bentrok dengan placeholder sistem {{%s}}", kode, token)
				case ada:
					return nil, fmt.Errorf("Placeholder {{%s}} dihasilkan dua field: %s dan %s", token, lain, kode)
				}
				pemilik[token] = kode
			}
			bagian.Fields = append(bagian.Fields, f)
		}
		out = append(out, bagian)
	}
	return out, nil
}
