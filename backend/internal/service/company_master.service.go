package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"

	"github.com/google/uuid"
)

type CompanyMasterService interface {
	GetAll() ([]entity.CompanyMaster, error)
	GetByID(id string) (*entity.CompanyMaster, error)
	GetActive() (*entity.CompanyMaster, error)
	Create(data entity.CompanyMaster) (*entity.CompanyMaster, error)
	Update(id string, data entity.CompanyMaster) (*entity.CompanyMaster, error)
	Delete(id string) error
	UploadTemplate(ctx context.Context, version string, docxData []byte) (*CompanyMasterTemplate, error)
	GetTemplate(version string) (*CompanyMasterTemplate, error)
	DeleteTemplate(ctx context.Context, version string) error
	GetActiveTemplate() (*CompanyMasterTemplate, error)
	GetFieldPositions(version string) ([]entity.TemplateFieldPosition, error)
	SaveFieldPositions(version string, positions []entity.TemplateFieldPosition) error
	GetTemplateBaseImage(version string, page int) ([]byte, error)
}

type CompanyMasterTemplate struct {
	Version      string `json:"version"`
	TemplatePath string `json:"template_path"`
	BasePDFPath  string `json:"base_pdf_path"`
	UploadedAt   string `json:"uploaded_at"`
}

type companyMasterService struct {
	repo              repository.CompanyMasterRepository
	storage           *storage.MinIOClient
	templateSvc       TemplateConversionService
	fieldPositionRepo repository.TemplateFieldPositionRepository
	templates         map[string]CompanyMasterTemplate
	mu                sync.RWMutex
}

func NewCompanyMasterService(repo repository.CompanyMasterRepository, storage *storage.MinIOClient, templateSvc TemplateConversionService, fieldPositionRepo repository.TemplateFieldPositionRepository) CompanyMasterService {
	return &companyMasterService{
		repo:              repo,
		storage:           storage,
		templateSvc:       templateSvc,
		fieldPositionRepo: fieldPositionRepo,
		templates:         make(map[string]CompanyMasterTemplate),
	}
}

func (s *companyMasterService) GetAll() ([]entity.CompanyMaster, error) {
	return s.repo.GetAll()
}

func (s *companyMasterService) GetByID(id string) (*entity.CompanyMaster, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(uid)
}

func (s *companyMasterService) GetActive() (*entity.CompanyMaster, error) {
	return s.repo.GetFirstActive()
}

func (s *companyMasterService) Create(data entity.CompanyMaster) (*entity.CompanyMaster, error) {
	if data.ID == uuid.Nil {
		data.ID = uuid.New()
	}
	if err := s.repo.Create(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *companyMasterService) Update(id string, data entity.CompanyMaster) (*entity.CompanyMaster, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetByID(uid)
	if err != nil {
		return nil, err
	}
	existing.Name = data.Name
	existing.Address = data.Address
	existing.NPWP = data.NPWP
	existing.Phone = data.Phone
	existing.Email = data.Email
	existing.DefaultPejabat = data.DefaultPejabat
	existing.DefaultJabatan = data.DefaultJabatan
	existing.DefaultTempatTtd = data.DefaultTempatTtd
	existing.IsActive = data.IsActive
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *companyMasterService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(uid)
}

func (s *companyMasterService) UploadTemplate(ctx context.Context, version string, docxData []byte) (*CompanyMasterTemplate, error) {
	if s.templateSvc == nil {
		return nil, fmt.Errorf("template conversion service not configured")
	}

	if version == "" || !isValidVersion(version) {
		return nil, fmt.Errorf("versi template tidak valid")
	}

	templatePath := fmt.Sprintf("templates/pks-template/v%s.docx", version)
	basePDFPath := fmt.Sprintf("templates/pks-base/v%s.pdf", version)

	if s.storage.GetFileContentType(ctx, templatePath) != "" {
		return nil, fmt.Errorf("template version %s sudah ada", version)
	}

	log.Printf("[UPLOAD][v%s] Uploading template docx...", version)
	if _, err := s.storage.UploadTemplate(ctx, version, docxData); err != nil {
		return nil, fmt.Errorf("gagal upload template: %w", err)
	}
	log.Printf("[UPLOAD][v%s] Template docx uploaded", version)

	log.Printf("[UPLOAD][v%s] Starting conversion...", version)
	if _, err := s.templateSvc.ConvertDocxToPDF(ctx, docxData, version); err != nil {
		log.Printf("[UPLOAD][v%s] Conversion failed, cleaning up: %v", version, err)
		s.storage.DeleteFile(ctx, templatePath)
		s.storage.DeleteFile(ctx, basePDFPath)
		return nil, fmt.Errorf("gagal konversi template: %w", err)
	}
	log.Printf("[UPLOAD][v%s] Conversion completed", version)

	// Verify the base PDF actually landed in MinIO. A failed LibreOffice or
	// MinIO step can leave a dangling .docx with no base PDF, which then 500s
	// on preview. Remove both files so the upload is treated as failed.
	log.Printf("[UPLOAD][v%s] Verifying base PDF in MinIO...", version)
	if s.storage.GetFileContentType(ctx, basePDFPath) == "" {
		log.Printf("[UPLOAD][v%s] Base PDF NOT found in MinIO, cleaning up", version)
		s.storage.DeleteFile(ctx, templatePath)
		s.storage.DeleteFile(ctx, basePDFPath)
		return nil, fmt.Errorf("base PDF tidak terupload ke MinIO, template dihapus")
	}
	log.Printf("[UPLOAD][v%s] Base PDF verified in MinIO", version)

	tmpl := CompanyMasterTemplate{
		Version:      version,
		TemplatePath: templatePath,
		BasePDFPath:  basePDFPath,
		UploadedAt:   time.Now().Format(time.RFC3339),
	}
	s.mu.Lock()
	s.templates[version] = tmpl
	s.mu.Unlock()
	return &tmpl, nil
}

func (s *companyMasterService) GetTemplate(version string) (*CompanyMasterTemplate, error) {
	if version == "" {
		version = "1"
	}
	s.mu.RLock()
	tmpl, ok := s.templates[version]
	s.mu.RUnlock()
	if ok {
		return &tmpl, nil
	}

	templatePath := fmt.Sprintf("templates/pks-template/v%s.docx", version)
	if s.storage.GetFileContentType(context.Background(), templatePath) == "" {
		return nil, fmt.Errorf("template version %s tidak ditemukan", version)
	}

	tmpl = CompanyMasterTemplate{
		Version:      version,
		TemplatePath: templatePath,
		BasePDFPath:  fmt.Sprintf("templates/pks-base/v%s.pdf", version),
	}
	s.mu.Lock()
	s.templates[version] = tmpl
	s.mu.Unlock()
	return &tmpl, nil
}

func (s *companyMasterService) DeleteTemplate(ctx context.Context, version string) error {
	if version == "" {
		version = "1"
	}
	templatePath := fmt.Sprintf("templates/pks-template/v%s.docx", version)
	basePDFPath := fmt.Sprintf("templates/pks-base/v%s.pdf", version)

	if err := s.storage.DeleteFile(ctx, templatePath); err != nil {
		return fmt.Errorf("gagal menghapus template: %w", err)
	}
	if err := s.storage.DeleteFile(ctx, basePDFPath); err != nil {
		return fmt.Errorf("gagal menghapus base PDF: %w", err)
	}

	s.mu.Lock()
	delete(s.templates, version)
	s.mu.Unlock()
	return nil
}

func (s *companyMasterService) ListTemplates() []CompanyMasterTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]CompanyMasterTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		result = append(result, t)
	}
	return result
}

func (s *companyMasterService) GetFieldPositions(version string) ([]entity.TemplateFieldPosition, error) {
	if version == "" {
		version = "1"
	}
	return s.fieldPositionRepo.GetByVersion(version)
}

func (s *companyMasterService) SaveFieldPositions(version string, positions []entity.TemplateFieldPosition) error {
	if version == "" {
		version = "1"
	}
	return s.fieldPositionRepo.Upsert(version, positions)
}

// GetTemplateBaseImage returns the base PDF for a template version as a PNG
// for the requested page, suitable for embedding in the calibration UI. It
// reuses the cached/converted base PDF from MinIO.
func (s *companyMasterService) GetTemplateBaseImage(version string, page int) ([]byte, error) {
	if version == "" {
		version = "1"
	}
	if page < 1 {
		page = 1
	}

	pdfData, err := s.templateSvc.EnsureBasePDFExists(context.Background(), version)
	if err != nil {
		return nil, fmt.Errorf("gagal menyiapkan base PDF: %w", err)
	}

	tmpDir := fmt.Sprintf("%s/templates", os.TempDir())
	os.MkdirAll(tmpDir, 0755)

	pdfPath := filepath.Join(tmpDir, fmt.Sprintf("calib-v%s.pdf", version))
	if err := os.WriteFile(pdfPath, pdfData, 0644); err != nil {
		return nil, fmt.Errorf("gagal menulis base PDF: %w", err)
	}
	defer os.Remove(pdfPath)

	// Use prefix without page number; pdftoppm will append -1, -2, etc.
	prefix := filepath.Join(tmpDir, "calib-page")
	cmd := exec.Command("pdftoppm", "-png", "-f", strconv.Itoa(page), "-l", strconv.Itoa(page), "-r", "150", pdfPath, prefix)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm gagal: %s: %s", err.Error(), string(output))
	}

	// Find the generated PNG file(s) instead of hardcoding the name.
	matches, err := filepath.Glob(fmt.Sprintf("%s-*.png", prefix))
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("pdftoppm output not found in %s", tmpDir)
	}

	generatedImg := matches[0]
	defer os.Remove(generatedImg)

	data, err := os.ReadFile(generatedImg)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca gambar hasil: %w", err)
	}
	return data, nil
}

func (s *companyMasterService) GetActiveTemplate() (*CompanyMasterTemplate, error) {
	if len(s.templates) > 0 {
		s.mu.RLock()
		defer s.mu.RUnlock()
		for _, t := range s.templates {
			return &t, nil
		}
	}

	templatePath := "templates/pks-template/v1.docx"
	if s.storage.GetFileContentType(context.Background(), templatePath) != "" {
		tmpl := CompanyMasterTemplate{
			Version:      "1",
			TemplatePath: templatePath,
			BasePDFPath:  "templates/pks-base/v1.pdf",
		}
		s.mu.Lock()
		s.templates["1"] = tmpl
		s.mu.Unlock()
		return &tmpl, nil
	}

	return nil, fmt.Errorf("tidak ada template yang diupload")
}

func isValidVersion(v string) bool {
	if v == "" {
		return false
	}
	for _, c := range v {
		if c < '0' || c > '9' {
			return false
		}
	}
	return v != "0"
}
