package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"legal-riu-portal/internal/errors"
	"legal-riu-portal/internal/storage"

	"github.com/google/uuid"
)

type TemplateConversionService interface {
	ConvertDocxToPDF(ctx context.Context, docxData []byte, version string) ([]byte, error)
	EnsureBasePDFExists(ctx context.Context, version string) ([]byte, error)
}

type templateConversionService struct {
	storage *storage.MinIOClient
}

func NewTemplateConversionService(s *storage.MinIOClient) TemplateConversionService {
	return &templateConversionService{storage: s}
}

func (s *templateConversionService) ConvertDocxToPDF(ctx context.Context, docxData []byte, version string) ([]byte, error) {
	tmpDir := fmt.Sprintf("%s/templates", os.TempDir())
	os.MkdirAll(tmpDir, 0755)

	docxUUID := uuid.New().String()[:8]
	docxPath := filepath.Join(tmpDir, fmt.Sprintf("template-v%s-%s.docx", version, docxUUID))
	pdfPath := filepath.Join(tmpDir, fmt.Sprintf("template-v%s-%s.pdf", version, docxUUID))

	log.Printf("[CONVERT][v%s] Writing temp docx: %s (size: %d bytes)", version, docxPath, len(docxData))
	if err := os.WriteFile(docxPath, docxData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp docx: %w", err)
	}
	defer os.Remove(docxPath)

	log.Printf("[CONVERT][v%s] Starting LibreOffice conversion...", version)
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", tmpDir, docxPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[CONVERT][v%s] LibreOffice FAILED: %s", version, string(output))
		return nil, fmt.Errorf("%w: %s: %s", errors.ErrConversionFailed, err.Error(), string(output))
	}
	log.Printf("[CONVERT][v%s] LibreOffice output: %s", version, string(output))

	log.Printf("[CONVERT][v%s] Reading generated PDF: %s", version, pdfPath)
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		log.Printf("[CONVERT][v%s] PDF read FAILED: %v", version, err)
		return nil, fmt.Errorf("%w: %w", errors.ErrConversionFailed, err)
	}
	log.Printf("[CONVERT][v%s] PDF generated successfully (size: %d bytes)", version, len(pdfData))
	defer os.Remove(pdfPath)

	log.Printf("[CONVERT][v%s] Uploading base PDF to MinIO...", version)
	if _, err := s.storage.UploadBasePDF(ctx, version, pdfData); err != nil {
		log.Printf("[CONVERT][v%s] MinIO upload FAILED: %v", version, err)
		return nil, fmt.Errorf("failed to upload base PDF: %w", err)
	}
	log.Printf("[CONVERT][v%s] Base PDF uploaded successfully", version)

	return pdfData, nil
}

func (s *templateConversionService) EnsureBasePDFExists(ctx context.Context, version string) ([]byte, error) {
	pdfData, err := s.storage.DownloadBasePDF(ctx, version)
	if err == nil {
		return pdfData, nil
	}

	docxData, err := s.storage.DownloadTemplate(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrTemplateNotFound, err)
	}

	return s.ConvertDocxToPDF(ctx, docxData, version)
}
