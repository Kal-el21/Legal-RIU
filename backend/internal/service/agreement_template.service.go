package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"

	"github.com/google/uuid"
)

const (
	docxContentType       = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	maxTemplateSizeBytes  = 25 * 1024 * 1024
	templateStoragePrefix = "agreement-templates/"
)

type AgreementTemplateService interface {
	List(string) ([]entity.AgreementTemplate, error)
	Get(string) (*entity.AgreementTemplate, error)
	Upload(context.Context, string, dto.UploadAgreementTemplateRequest, *multipart.FileHeader) (*entity.AgreementTemplate, dto.AgreementTemplateValidation, error)
	NewVersion(context.Context, string, dto.UploadAgreementTemplateRequest) (*entity.AgreementTemplate, dto.AgreementTemplateValidation, error)
	Activate(context.Context, string, string) (*entity.AgreementTemplate, error)
	Delete(string) error
	Download(context.Context, string) ([]byte, string, error)
	Preview(context.Context, string) ([]byte, error)
	Placeholders(string) ([]dto.PlaceholderReference, error)

	// Dipakai service dokumen untuk mengunci dan memuat versi template.
	ActiveTemplate(context.Context, string) (*entity.AgreementTemplate, []byte, error)
	TemplateByID(context.Context, uuid.UUID) (*entity.AgreementTemplate, []byte, error)
	ActiveTypes() ([]entity.AgreementType, error)
	ActiveType(string) (*entity.AgreementType, *entity.AgreementTemplate, error)
}

type agreementTemplateService struct {
	repo      repository.AgreementTemplateRepository
	storage   *storage.MinIOClient
	generator *AgreementGenerator
	converter DOCXConverter
	// Isi setiap versi tidak pernah berubah, jadi cache cukup di-key oleh ID
	// tanpa perlu invalidasi.
	cache sync.Map
}

func NewAgreementTemplateService(r repository.AgreementTemplateRepository, s *storage.MinIOClient) AgreementTemplateService {
	return &agreementTemplateService{repo: r, storage: s, generator: NewAgreementGenerator()}
}

// Placeholders mengembalikan token yang bisa dipakai di template jenis dokumen
// itu, dari field kerjanya, sehingga Admin bisa menyusun Word sebelum versi
// pertama ada.
func (s *agreementTemplateService) Placeholders(code string) ([]dto.PlaceholderReference, error) {
	typ, e := s.repo.TypeByCode(strings.ToUpper(strings.TrimSpace(code)))
	if e != nil {
		return nil, ErrJenisTidakAda
	}
	sections, e := fieldKerja(typ)
	if e != nil {
		return nil, e
	}
	return PlaceholderCatalog(sections), nil
}

// ErrTemplateTidakAda: versi template yang diminta tidak ada (atau sudah dihapus).
var ErrTemplateTidakAda = errors.New("template tidak ditemukan")

// GagalPencocokan: DOCX memuat placeholder yang tidak punya field. Hasil memuat
// daftarnya, supaya layar Admin bisa menawarkan "Jadikan field".
type GagalPencocokan struct {
	Hasil dto.AgreementTemplateValidation
}

func (g *GagalPencocokan) Error() string {
	return "placeholder tidak dikenal: " + strings.Join(g.Hasil.UnknownPlaceholders, ", ") + ". Buat field-nya di jenis dokumen, atau perbaiki ejaannya di Word"
}

// fieldKerja membaca field kerja jenis dokumen (tahap 9.3), yang disalin ke
// setiap versi baru.
func fieldKerja(typ *entity.AgreementType) ([]AgreementSectionDefinition, error) {
	var sections []AgreementSectionDefinition
	if len(typ.Fields) == 0 || json.Unmarshal(typ.Fields, &sections) != nil || len(sections) == 0 {
		return nil, fmt.Errorf("jenis dokumen %s belum punya field; susun field sebelum membuat versi template", typ.Code)
	}
	return sections, nil
}

func (s *agreementTemplateService) List(code string) ([]entity.AgreementTemplate, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		normalized = "PKS"
	}
	items, e := s.repo.FindByCode(normalized)
	if e != nil {
		return nil, e
	}
	// Catatan: satu hitungan per versi; cukup karena versi per jenis hanya puluhan.
	for i := range items {
		if items[i].RunningDocuments, e = s.repo.CountRunningDocuments(items[i].ID); e != nil {
			return nil, e
		}
	}
	return items, nil
}

func (s *agreementTemplateService) Get(id string) (*entity.AgreementTemplate, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, errors.New("ID template tidak valid")
	}
	v, e := s.repo.FindByID(uid)
	if e != nil {
		return nil, ErrTemplateTidakAda
	}
	return v, nil
}

func (s *agreementTemplateService) Upload(ctx context.Context, uploaderID string, req dto.UploadAgreementTemplateRequest, file *multipart.FileHeader) (*entity.AgreementTemplate, dto.AgreementTemplateValidation, error) {
	data, e := readTemplateUpload(file)
	if e != nil {
		return nil, dto.AgreementTemplateValidation{}, e
	}
	return s.buatVersi(ctx, uploaderID, req, data, filepath.Base(file.Filename))
}

// NewVersion: "Simpan sebagai versi baru" tanpa DOCX baru, mis. setelah Admin
// mengganti label atau status wajib field. DOCX disalin dari versi terakhir.
func (s *agreementTemplateService) NewVersion(ctx context.Context, uploaderID string, req dto.UploadAgreementTemplateRequest) (*entity.AgreementTemplate, dto.AgreementTemplateValidation, error) {
	items, e := s.repo.FindByCode(strings.ToUpper(strings.TrimSpace(req.Code)))
	if e != nil {
		return nil, dto.AgreementTemplateValidation{}, e
	}
	for i := range items {
		if items[i].IsLegacy {
			continue
		}
		data, e := s.load(ctx, &items[i])
		if e != nil {
			return nil, dto.AgreementTemplateValidation{}, e
		}
		return s.buatVersi(ctx, uploaderID, req, data, items[i].FileName)
	}
	return nil, dto.AgreementTemplateValidation{}, errors.New("belum ada versi untuk disalin; unggah DOCX untuk versi pertama")
}

// buatVersi menguji DOCX terhadap field kerja jenis dokumen, lalu menyimpannya
// sebagai versi Draft berikutnya bersama salinan field kerja itu.
func (s *agreementTemplateService) buatVersi(ctx context.Context, uploaderID string, req dto.UploadAgreementTemplateRequest, data []byte, fileName string) (*entity.AgreementTemplate, dto.AgreementTemplateValidation, error) {
	var hasil dto.AgreementTemplateValidation
	uid, e := uuid.Parse(uploaderID)
	if e != nil {
		return nil, hasil, errors.New("user tidak valid")
	}
	note := strings.TrimSpace(req.Note)
	if note == "" {
		return nil, hasil, errors.New("Catatan perubahan wajib diisi: tulis apa yang berubah dibanding versi sebelumnya")
	}
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		code = "PKS"
	}
	typ, e := s.repo.TypeByCode(code)
	if e != nil {
		return nil, hasil, fmt.Errorf("tipe dokumen %q tidak terdaftar", code)
	}
	sections, e := fieldKerja(typ)
	if e != nil {
		return nil, hasil, e
	}
	if hasil, e = periksaTemplate(data, sections); e != nil {
		return nil, hasil, e
	}

	// Dry-run memakai jalur generate yang sama dengan dokumen sungguhan supaya
	// template bermasalah tertolak sebelum sempat dipakai.
	docx, checksum, e := s.generator.Generate(data, sampleAgreementValues(sections), GenerateOptions{Draft: true})
	if e != nil {
		return nil, hasil, errors.New("uji coba pembentukan dokumen gagal: " + e.Error())
	}
	pdf, e := s.converter.ToPDF(ctx, docx)
	if e != nil {
		return nil, hasil, errors.New("uji coba konversi PDF gagal: " + e.Error())
	}

	version, e := s.repo.NextVersion(code)
	if e != nil {
		return nil, hasil, e
	}
	id := uuid.New()
	base := templateStoragePrefix + code + "/" + id.String()
	if e = s.storage.UploadBytes(ctx, base+".docx", data, docxContentType); e != nil {
		return nil, hasil, e
	}
	if e = s.storage.UploadBytes(ctx, base+"-preview.pdf", pdf, "application/pdf"); e != nil {
		s.storage.DeleteFile(ctx, base+".docx")
		return nil, hasil, e
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("%s versi %d", typ.Name, version)
	}
	tpl := &entity.AgreementTemplate{
		Code:          code,
		Version:       version,
		Name:          name,
		FileName:      fileName,
		FilePath:      base + ".docx",
		PreviewPath:   base + "-preview.pdf",
		Checksum:      checksum,
		Status:        entity.TemplateStatusDraft,
		Placeholders:  marshalPlaceholders(hasil.Placeholders),
		Note:          note,
		EffectiveDate: parseEffectiveDate(req.EffectiveDate),
		UploadedBy:    &uid,
		Fields:        typ.Fields,
	}
	tpl.ID = id
	if e = s.repo.Create(tpl); e != nil {
		s.storage.DeleteFile(ctx, base+".docx")
		s.storage.DeleteFile(ctx, base+"-preview.pdf")
		return nil, hasil, errors.New("gagal menyimpan template")
	}
	s.cache.Store(tpl.ID, data)
	v, e := s.repo.FindByID(tpl.ID)
	return v, hasil, e
}

// Delete menghapus versi yang tidak lagi dipakai. Versi aktif dan versi yang
// masih dikunci pengajuan yang belum selesai ditolak; pengajuan Selesai tidak
// bergantung pada versinya karena dokumen finalnya sudah tersimpan.
func (s *agreementTemplateService) Delete(id string) error {
	tpl, e := s.Get(id)
	if e != nil {
		return e
	}
	if tpl.Status == entity.TemplateStatusActive {
		return errors.New("Versi aktif tidak bisa dihapus; aktifkan versi lain lebih dulu")
	}
	n, e := s.repo.CountRunningDocuments(tpl.ID)
	if e != nil {
		return e
	}
	if n > 0 {
		return fmt.Errorf("Versi ini masih dipakai %d pengajuan yang belum selesai, jadi tidak bisa dihapus", n)
	}
	return s.repo.Delete(tpl.ID)
}

func (s *agreementTemplateService) Activate(ctx context.Context, id, actorID string) (*entity.AgreementTemplate, error) {
	aid, e := uuid.Parse(actorID)
	if e != nil {
		return nil, errors.New("user tidak valid")
	}
	tpl, e := s.Get(id)
	if e != nil {
		return nil, e
	}
	if tpl.IsLegacy {
		return nil, ErrTemplateLama
	}
	if tpl.Status == entity.TemplateStatusActive {
		return tpl, nil
	}
	if e = s.repo.Activate(tpl.ID, tpl.Code, aid); e != nil {
		return nil, errors.New("gagal mengaktifkan template")
	}
	return s.repo.FindByID(tpl.ID)
}

func (s *agreementTemplateService) Download(ctx context.Context, id string) ([]byte, string, error) {
	tpl, e := s.Get(id)
	if e != nil {
		return nil, "", e
	}
	data, e := s.load(ctx, tpl)
	if e != nil {
		return nil, "", e
	}
	return data, tpl.FileName, nil
}

func (s *agreementTemplateService) Preview(ctx context.Context, id string) ([]byte, error) {
	tpl, e := s.Get(id)
	if e != nil {
		return nil, e
	}
	if tpl.IsLegacy {
		return nil, ErrTemplateLama
	}
	if tpl.PreviewPath != "" {
		if data, err := s.fetch(ctx, tpl.PreviewPath); err == nil {
			return data, nil
		}
	}
	// Template hasil migrasi belum punya pratinjau tersimpan, bentuk sekarang.
	sections, e := templateSections(tpl)
	if e != nil {
		return nil, e
	}
	source, e := s.load(ctx, tpl)
	if e != nil {
		return nil, e
	}
	docx, _, e := s.generator.Generate(source, sampleAgreementValues(sections), GenerateOptions{Draft: true})
	if e != nil {
		return nil, e
	}
	return s.converter.ToPDF(ctx, docx)
}

func (s *agreementTemplateService) ActiveTemplate(ctx context.Context, code string) (*entity.AgreementTemplate, []byte, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	tpl, e := s.repo.FindActive(normalized)
	if e != nil {
		return nil, nil, fmt.Errorf("belum ada template aktif untuk tipe dokumen %q", normalized)
	}
	if tpl.IsLegacy {
		return nil, nil, ErrTemplateLama
	}
	data, e := s.load(ctx, tpl)
	if e != nil {
		return nil, nil, e
	}
	return tpl, data, nil
}

func (s *agreementTemplateService) ActiveTypes() ([]entity.AgreementType, error) {
	return s.repo.FindTypes()
}

// ActiveType mengembalikan jenis dokumen aktif beserta versi template aktifnya,
// tanpa memuat file DOCX.
func (s *agreementTemplateService) ActiveType(code string) (*entity.AgreementType, *entity.AgreementTemplate, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	typ, e := s.repo.FindType(normalized)
	if e != nil {
		return nil, nil, errors.New("tipe dokumen tidak valid")
	}
	tpl, e := s.repo.FindActive(normalized)
	if e != nil {
		return nil, nil, fmt.Errorf("belum ada template aktif untuk tipe dokumen %q", normalized)
	}
	if tpl.IsLegacy {
		return nil, nil, ErrTemplateLama
	}
	return typ, tpl, nil
}

func (s *agreementTemplateService) TemplateByID(ctx context.Context, id uuid.UUID) (*entity.AgreementTemplate, []byte, error) {
	tpl, e := s.repo.FindByID(id)
	if e != nil {
		return nil, nil, errors.New("template dokumen tidak ditemukan")
	}
	if tpl.IsLegacy {
		return nil, nil, ErrTemplateLama
	}
	data, e := s.load(ctx, tpl)
	if e != nil {
		return nil, nil, e
	}
	return tpl, data, nil
}

func (s *agreementTemplateService) load(ctx context.Context, tpl *entity.AgreementTemplate) ([]byte, error) {
	if cached, ok := s.cache.Load(tpl.ID); ok {
		return cached.([]byte), nil
	}
	data, e := s.fetch(ctx, tpl.FilePath)
	if e != nil {
		return nil, fmt.Errorf("file template %s tidak dapat dibaca", tpl.Name)
	}
	s.cache.Store(tpl.ID, data)
	return data, nil
}

func (s *agreementTemplateService) fetch(ctx context.Context, path string) ([]byte, error) {
	obj, e := s.storage.GetFileObject(ctx, path)
	if e != nil {
		return nil, e
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func readTemplateUpload(file *multipart.FileHeader) ([]byte, error) {
	if file == nil {
		return nil, errors.New("file template wajib diunggah")
	}
	if strings.ToLower(filepath.Ext(file.Filename)) != ".docx" {
		return nil, errors.New("template harus berformat .docx")
	}
	if file.Size <= 0 {
		return nil, errors.New("file template kosong")
	}
	if file.Size > maxTemplateSizeBytes {
		return nil, fmt.Errorf("ukuran template maksimal %d MB", maxTemplateSizeBytes/(1024*1024))
	}
	f, e := file.Open()
	if e != nil {
		return nil, errors.New("file template tidak dapat dibuka")
	}
	defer f.Close()
	return io.ReadAll(f)
}

func validateTemplatePlaceholders(found []string, sections []AgreementSectionDefinition) error {
	supported := map[string]bool{}
	for _, token := range SupportedPlaceholderTokens(sections) {
		supported[token] = true
	}
	unknown := make([]string, 0)
	for _, token := range found {
		if !supported[token] {
			unknown = append(unknown, token)
		}
	}
	if len(unknown) > 0 {
		return &GagalPencocokan{dto.AgreementTemplateValidation{Placeholders: found, UnknownPlaceholders: unknown}}
	}
	return nil
}

// periksaTemplate mencocokkan DOCX dengan field: placeholder tanpa field ditolak;
// field yang tidak dipakai dan deretan Enter kosong menjadi peringatan. Berjalan
// sebelum konversi PDF.
func periksaTemplate(data []byte, sections []AgreementSectionDefinition) (dto.AgreementTemplateValidation, error) {
	hasil := dto.AgreementTemplateValidation{UnknownPlaceholders: []string{}, UnusedPlaceholders: []string{}, Warnings: []string{}}
	found, scopeListNumbered, e := InspectTemplate(data)
	if e != nil {
		return hasil, e
	}
	hasil.Placeholders = found
	if e = validateTemplatePlaceholders(found, sections); e != nil {
		return e.(*GagalPencocokan).Hasil, e
	}
	// scopeListNumbered hanya false kalau {{RUANG_LINGKUP_LIST}} ada dan paragrafnya
	// bukan numbered list; DOCX tanpa placeholder itu tidak ditolak.
	if !scopeListNumbered {
		return hasil, errors.New("paragraf " + scopeListPlaceholder + " harus diformat sebagai numbered list Word agar poin ruang lingkup bernomor a, b, c")
	}
	if e = periksaParagrafUlang(data, sections); e != nil {
		return hasil, e
	}
	dipakai := map[string]bool{}
	for _, p := range found {
		dipakai[p] = true
	}
	for _, s := range sections {
		for _, f := range s.Fields {
			terpakai := f.Name == scopeField && dipakai[scopeListPlaceholder]
			for token := range tokenField(f) {
				terpakai = terpakai || dipakai["{{"+token+"}}"]
			}
			if !terpakai {
				hasil.UnusedPlaceholders = append(hasil.UnusedPlaceholders, "{{"+strings.ToUpper(f.Name)+"}}")
			}
		}
	}
	if len(hasil.UnusedPlaceholders) > 0 {
		hasil.Warnings = append(hasil.Warnings, "Field berikut tidak dipakai di DOCX, jadi isiannya tidak tercetak: "+strings.Join(hasil.UnusedPlaceholders, ", "))
	}
	kosong, sebelum, e := paragrafKosongTerpanjang(data)
	if e != nil {
		return hasil, e
	}
	if kosong >= 4 {
		hasil.Warnings = append(hasil.Warnings, fmt.Sprintf("Ada %d paragraf kosong berturut-turut sebelum %q. Pakai pemisah halaman (Ctrl+Enter), bukan Enter berulang, supaya tidak muncul ruang atau halaman kosong saat isian panjang.", kosong, sebelum))
	}
	return hasil, nil
}

func marshalPlaceholders(v []string) json.RawMessage {
	raw, e := json.Marshal(v)
	if e != nil {
		return nil
	}
	return raw
}

func parseEffectiveDate(raw string) *time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	t, e := time.Parse("2006-01-02", value)
	if e != nil {
		return nil
	}
	return &t
}

// TemplateChecksum dipakai juga oleh seed saat mendaftarkan template bawaan
// sebagai versi pertama.
// ErrTemplateLama: template versi lama (tanpa placeholder, dulu diisi lewat
// pencocokan kalimat) tidak bisa lagi dipakai membentuk dokumen, dipratinjau,
// atau diaktifkan. Versinya tetap tersimpan sebagai riwayat.
var ErrTemplateLama = errors.New("template ini memakai format lama yang tidak didukung lagi; pakai versi template yang lebih baru")

func TemplateChecksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
