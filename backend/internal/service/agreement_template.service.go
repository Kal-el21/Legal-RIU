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
	Upload(context.Context, string, dto.UploadAgreementTemplateRequest, *multipart.FileHeader) (*entity.AgreementTemplate, error)
	Activate(context.Context, string, string) (*entity.AgreementTemplate, error)
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
// itu, dari field versi aktifnya.
func (s *agreementTemplateService) Placeholders(code string) ([]dto.PlaceholderReference, error) {
	_, tpl, e := s.ActiveType(code)
	if e != nil {
		return nil, e
	}
	sections, e := templateSections(tpl)
	if e != nil {
		return nil, e
	}
	return PlaceholderCatalog(sections), nil
}

func (s *agreementTemplateService) List(code string) ([]entity.AgreementTemplate, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		normalized = "PKS"
	}
	return s.repo.FindByCode(normalized)
}

func (s *agreementTemplateService) Get(id string) (*entity.AgreementTemplate, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, errors.New("ID template tidak valid")
	}
	v, e := s.repo.FindByID(uid)
	if e != nil {
		return nil, errors.New("template tidak ditemukan")
	}
	return v, nil
}

func (s *agreementTemplateService) Upload(ctx context.Context, uploaderID string, req dto.UploadAgreementTemplateRequest, file *multipart.FileHeader) (*entity.AgreementTemplate, error) {
	uid, e := uuid.Parse(uploaderID)
	if e != nil {
		return nil, errors.New("user tidak valid")
	}
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		code = "PKS"
	}
	data, e := readTemplateUpload(file)
	if e != nil {
		return nil, e
	}
	// Field belum bisa diubah lewat unggahan (PR 9 tahap 9.4), jadi versi baru
	// diuji dan disimpan dengan field versi aktif.
	typ, aktif, e := s.ActiveType(code)
	if e != nil {
		return nil, e
	}
	sections, e := templateSections(aktif)
	if e != nil {
		return nil, e
	}

	placeholders, scopeListNumbered, e := InspectTemplate(data)
	if e != nil {
		return nil, e
	}
	if e = validateTemplatePlaceholders(placeholders, sections); e != nil {
		return nil, e
	}
	if !scopeListNumbered {
		return nil, errors.New("paragraf " + scopeListPlaceholder + " harus diformat sebagai numbered list Word agar poin ruang lingkup bernomor a, b, c")
	}

	// Dry-run memakai jalur generate yang sama dengan dokumen sungguhan supaya
	// template bermasalah tertolak sebelum sempat dipakai.
	docx, checksum, e := s.generator.Generate(data, sampleAgreementValues(sections), GenerateOptions{Draft: true})
	if e != nil {
		return nil, errors.New("uji coba pembentukan dokumen gagal: " + e.Error())
	}
	pdf, e := s.converter.ToPDF(ctx, docx)
	if e != nil {
		return nil, errors.New("uji coba konversi PDF gagal: " + e.Error())
	}

	version, e := s.repo.NextVersion(code)
	if e != nil {
		return nil, e
	}
	id := uuid.New()
	base := templateStoragePrefix + code + "/" + id.String()
	if e = s.storage.UploadBytes(ctx, base+".docx", data, docxContentType); e != nil {
		return nil, e
	}
	if e = s.storage.UploadBytes(ctx, base+"-preview.pdf", pdf, "application/pdf"); e != nil {
		s.storage.DeleteFile(ctx, base+".docx")
		return nil, e
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = fmt.Sprintf("%s versi %d", typ.Name, version)
	}
	tpl := &entity.AgreementTemplate{
		Code:          code,
		Version:       version,
		Name:          name,
		FileName:      filepath.Base(file.Filename),
		FilePath:      base + ".docx",
		PreviewPath:   base + "-preview.pdf",
		Checksum:      checksum,
		Status:        entity.TemplateStatusDraft,
		Placeholders:  marshalPlaceholders(placeholders),
		Note:          strings.TrimSpace(req.Note),
		EffectiveDate: parseEffectiveDate(req.EffectiveDate),
		UploadedBy:    &uid,
		Fields:        aktif.Fields,
	}
	tpl.ID = id
	if e = s.repo.Create(tpl); e != nil {
		s.storage.DeleteFile(ctx, base+".docx")
		s.storage.DeleteFile(ctx, base+"-preview.pdf")
		return nil, errors.New("gagal menyimpan template")
	}
	s.cache.Store(tpl.ID, data)
	return s.repo.FindByID(tpl.ID)
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
		return errors.New("placeholder tidak dikenal: " + strings.Join(unknown, ", "))
	}
	return nil
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
