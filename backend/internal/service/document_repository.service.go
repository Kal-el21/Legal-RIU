package service

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type RepositoryDocumentService interface {
	GetAllDocuments(featureCode string, search string) ([]dto.RepositoryDocumentResponse, error)
	GetDocumentByID(id string) (*dto.RepositoryDocumentResponse, error)
	DownloadDocument(id string) (*minio.Object, string, error)
	BackfillFromExistingData(db *gorm.DB) error
	SyncCaseDocument(caseID string, filePath string, fileName string, fileSize int64) error
	SyncCasePhoto(caseID string, filePath string, fileName string, fileSize int64) error
	SyncChronologyDocument(chronID string, caseID string, filePath string) error
	RemoveCaseDocuments(caseID string) error
	RemoveChronologyDocuments(chronID string) error
}

type repositoryDocumentService struct {
	repo    repository.RepositoryDocumentRepository
	storage *storage.MinIOClient
}

func NewRepositoryDocumentService(repo repository.RepositoryDocumentRepository, s *storage.MinIOClient) RepositoryDocumentService {
	return &repositoryDocumentService{repo: repo, storage: s}
}

func (s *repositoryDocumentService) GetAllDocuments(featureCode string, search string) ([]dto.RepositoryDocumentResponse, error) {
	docs, err := s.repo.FindAll(featureCode, search)
	if err != nil {
		return nil, err
	}
	return toRepositoryDocumentResponses(docs), nil
}

func (s *repositoryDocumentService) GetDocumentByID(id string) (*dto.RepositoryDocumentResponse, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if !doc.IsActive {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	return &dto.RepositoryDocumentResponse{
		ID:           doc.ID.String(),
		Title:        doc.Title,
		FeatureCode:  doc.FeatureCode,
		SourceType:   doc.SourceType,
		SourceID:     doc.SourceID,
		UploaderType: doc.UploaderType,
		FileName:     doc.FileName,
		MIMEType:     doc.MIMEType,
		FileSize:     doc.FileSize,
		IsActive:     doc.IsActive,
		CreatedBy:    doc.CreatedBy.String(),
		UpdatedBy:    doc.UpdatedBy.String(),
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}, nil
}

func (s *repositoryDocumentService) DownloadDocument(id string) (*minio.Object, string, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, "", errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, "", errors.New("dokumen tidak ditemukan")
	}
	if !doc.IsActive {
		return nil, "", errors.New("dokumen tidak ditemukan")
	}
	obj, err := s.storage.GetFileObject(context.Background(), doc.FilePath)
	if err != nil {
		return nil, "", errors.New("gagal mengambil file dari storage")
	}
	return obj, doc.FileName, nil
}

func (s *repositoryDocumentService) SyncCaseDocument(caseID string, filePath string, fileName string, fileSize int64) error {
	if filePath == "" {
		return nil
	}
	doc := entity.RepositoryDocument{
		Title:        fileName,
		FeatureCode:  "case_management",
		SourceType:   "legal_cases",
		SourceID:     caseID,
		UploaderType: "requester",
		FilePath:     filePath,
		FileName:     fileName,
		FileSize:     fileSize,
		MIMEType:     detectMIME(fileName),
		IsActive:     true,
	}
	return s.repo.UpsertBySource(doc)
}

func (s *repositoryDocumentService) SyncCasePhoto(caseID string, filePath string, fileName string, fileSize int64) error {
	if filePath == "" {
		return nil
	}
	doc := entity.RepositoryDocument{
		Title:        fileName,
		FeatureCode:  "case_management",
		SourceType:   "legal_cases_photo",
		SourceID:     caseID,
		UploaderType: "requester",
		FilePath:     filePath,
		FileName:     fileName,
		FileSize:     fileSize,
		MIMEType:     detectMIME(fileName),
		IsActive:     true,
	}
	return s.repo.UpsertBySource(doc)
}

func (s *repositoryDocumentService) SyncChronologyDocument(chronID string, caseID string, filePath string) error {
	if filePath == "" {
		return nil
	}
	doc := entity.RepositoryDocument{
		Title:        filepath.Base(filePath),
		FeatureCode:  "case_management",
		SourceType:   "case_chronologies",
		SourceID:     chronID,
		UploaderType: "requester",
		FilePath:     filePath,
		FileName:     filepath.Base(filePath),
		MIMEType:     detectMIME(filePath),
		IsActive:     true,
	}
	return s.repo.UpsertBySource(doc)
}

func (s *repositoryDocumentService) RemoveCaseDocuments(caseID string) error {
	if err := s.repo.MarkInactiveBySource("legal_cases", []string{caseID}); err != nil {
		return err
	}
	return s.repo.MarkInactiveBySource("legal_cases_photo", []string{caseID})
}

func (s *repositoryDocumentService) RemoveChronologyDocuments(chronID string) error {
	return s.repo.MarkInactiveBySource("case_chronologies", []string{chronID})
}

func (s *repositoryDocumentService) BackfillFromExistingData(db *gorm.DB) error {
	docs := make([]entity.RepositoryDocument, 0)

	var adminID uuid.UUID
	if err := db.Raw(`SELECT id FROM users WHERE role = 'ADMIN' LIMIT 1`).Scan(&adminID).Error; err != nil {
		adminID = uuid.Nil
	}
	if adminID == uuid.Nil {
		adminID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}

	loAttachments, err := s.backfillLegalOpinionAttachments(db)
	if err != nil {
		return err
	}
	docs = append(docs, loAttachments...)

	loResults, err := s.backfillLegalOpinionResults(db)
	if err != nil {
		return err
	}
	docs = append(docs, loResults...)

	drAttachments, err := s.backfillDocumentReviewAttachments(db)
	if err != nil {
		return err
	}
	docs = append(docs, drAttachments...)

	drResults, err := s.backfillDocumentReviewResults(db)
	if err != nil {
		return err
	}
	docs = append(docs, drResults...)

	agAttachments, err := s.backfillAgreementAttachments(db)
	if err != nil {
		return err
	}
	docs = append(docs, agAttachments...)

	agGenerated, err := s.backfillAgreementGeneratedFiles(db)
	if err != nil {
		return err
	}
	docs = append(docs, agGenerated...)

	caseDocs, err := s.backfillLegalCaseDocuments(db)
	if err != nil {
		return err
	}
	docs = append(docs, caseDocs...)

	casePhotos, err := s.backfillLegalCasePhotos(db)
	if err != nil {
		return err
	}
	docs = append(docs, casePhotos...)

	chronologyDocs, err := s.backfillChronologyDocuments(db)
	if err != nil {
		return err
	}
	docs = append(docs, chronologyDocs...)

	if len(docs) == 0 {
		return nil
	}

	now := time.Now()
	for i := range docs {
		docs[i].CreatedAt = now
		docs[i].UpdatedAt = now
		docs[i].CreatedBy = adminID
		docs[i].UpdatedBy = adminID
		docs[i].IsActive = true
	}

	if err := s.repo.MarkInactiveBySource("all", nil); err != nil {
		return err
	}

	sourceTypes := make(map[string][]string)
	for _, doc := range docs {
		sourceTypes[doc.SourceType] = append(sourceTypes[doc.SourceType], doc.SourceID)
	}
	for sourceType, sourceIDs := range sourceTypes {
		uniqueIDs := uniqueStrings(sourceIDs)
		if err := s.repo.MarkInactiveBySource(sourceType, uniqueIDs); err != nil {
			return err
		}
	}

	if err := s.repo.BatchUpsert(docs); err != nil {
		return err
	}

	return nil
}

func (s *repositoryDocumentService) backfillLegalOpinionAttachments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID             uuid.UUID
		LegalOpinionID uuid.UUID
		FileName       string
		FilePath       string
		FileSize       int64
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, legal_opinion_id, file_name, file_path, file_size
		FROM legal_opinion_attachments
		WHERE file_path IS NOT NULL AND file_path != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, entity.RepositoryDocument{
			Title:        r.FileName,
			FeatureCode:  "legal_opinion",
			SourceType:   "legal_opinion_attachments",
			SourceID:     r.LegalOpinionID.String(),
			UploaderType: "requester",
			FilePath:     r.FilePath,
			FileName:     r.FileName,
			FileSize:     r.FileSize,
			MIMEType:     detectMIME(r.FileName),
		})
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillLegalOpinionResults(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID             uuid.UUID
		LegalOpinionID uuid.UUID
		FileName       string
		FilePath       string
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, legal_opinion_id, file_name, file_path
		FROM legal_opinion_results
		WHERE file_path IS NOT NULL AND file_path != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, entity.RepositoryDocument{
			Title:        r.FileName,
			FeatureCode:  "legal_opinion",
			SourceType:   "legal_opinion_results",
			SourceID:     r.LegalOpinionID.String(),
			UploaderType: "approver",
			FilePath:     r.FilePath,
			FileName:     r.FileName,
			MIMEType:     detectMIME(r.FileName),
		})
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillDocumentReviewAttachments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID               uuid.UUID
		DocumentReviewID uuid.UUID
		FileName         string
		FilePath         string
		FileSize         int64
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, document_review_id, file_name, file_path, file_size
		FROM document_review_attachments
		WHERE file_path IS NOT NULL AND file_path != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, entity.RepositoryDocument{
			Title:        r.FileName,
			FeatureCode:  "document_review",
			SourceType:   "document_review_attachments",
			SourceID:     r.DocumentReviewID.String(),
			UploaderType: "requester",
			FilePath:     r.FilePath,
			FileName:     r.FileName,
			FileSize:     r.FileSize,
			MIMEType:     detectMIME(r.FileName),
		})
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillDocumentReviewResults(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID               uuid.UUID
		DocumentReviewID uuid.UUID
		FileName         string
		FilePath         string
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, document_review_id, file_name, file_path
		FROM document_review_results
		WHERE file_path IS NOT NULL AND file_path != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, entity.RepositoryDocument{
			Title:        r.FileName,
			FeatureCode:  "document_review",
			SourceType:   "document_review_results",
			SourceID:     r.DocumentReviewID.String(),
			UploaderType: "approver",
			FilePath:     r.FilePath,
			FileName:     r.FileName,
			MIMEType:     detectMIME(r.FileName),
		})
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillAgreementAttachments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID                 uuid.UUID
		AgreementDocumentID uuid.UUID
		FileName           string
		FilePath           string
		FileSize           int64
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, agreement_document_id, file_name, file_path, file_size
		FROM agreement_attachments
		WHERE file_path IS NOT NULL AND file_path != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, entity.RepositoryDocument{
			Title:        r.FileName,
			FeatureCode:  "agreement_document",
			SourceType:   "agreement_attachments",
			SourceID:     r.AgreementDocumentID.String(),
			UploaderType: "requester",
			FilePath:     r.FilePath,
			FileName:     r.FileName,
			FileSize:     r.FileSize,
			MIMEType:     detectMIME(r.FileName),
		})
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillAgreementGeneratedFiles(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID                uuid.UUID
		GeneratedDocxPath string
		GeneratedPdfPath  string
		GeneratedFileName string
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, generated_docx_path, generated_pdf_path, generated_file_name
		FROM agreement_documents
		WHERE (generated_docx_path IS NOT NULL AND generated_docx_path != '')
		   OR (generated_pdf_path IS NOT NULL AND generated_pdf_path != '')
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows)*2)
	for _, r := range rows {
		if r.GeneratedDocxPath != "" {
			docs = append(docs, entity.RepositoryDocument{
				Title:        r.GeneratedFileName + ".docx",
				FeatureCode:  "agreement_document",
				SourceType:   "agreement_documents",
				SourceID:     r.ID.String(),
				UploaderType: "approver",
				FilePath:     r.GeneratedDocxPath,
				FileName:     r.GeneratedFileName + ".docx",
				MIMEType:     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			})
		}
		if r.GeneratedPdfPath != "" {
			docs = append(docs, entity.RepositoryDocument{
				Title:        r.GeneratedFileName + ".pdf",
				FeatureCode:  "agreement_document",
				SourceType:   "agreement_documents",
				SourceID:     r.ID.String(),
				UploaderType: "approver",
				FilePath:     r.GeneratedPdfPath,
				FileName:     r.GeneratedFileName + ".pdf",
				MIMEType:     "application/pdf",
			})
		}
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillLegalCaseDocuments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID           uuid.UUID
		DocumentLink string
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, document_link
		FROM legal_cases
		WHERE document_link IS NOT NULL AND document_link != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, entity.RepositoryDocument{
			Title:        filepath.Base(r.DocumentLink),
			FeatureCode:  "case_management",
			SourceType:   "legal_cases",
			SourceID:     r.ID.String(),
			UploaderType: "requester",
			FilePath:     r.DocumentLink,
			FileName:     filepath.Base(r.DocumentLink),
			MIMEType:     detectMIME(r.DocumentLink),
		})
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillLegalCasePhotos(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID    uuid.UUID
		Photo string
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, photo
		FROM legal_cases
		WHERE photo IS NOT NULL AND photo != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, entity.RepositoryDocument{
			Title:        filepath.Base(r.Photo),
			FeatureCode:  "case_management",
			SourceType:   "legal_cases",
			SourceID:     r.ID.String(),
			UploaderType: "requester",
			FilePath:     r.Photo,
			FileName:     filepath.Base(r.Photo),
			MIMEType:     detectMIME(r.Photo),
		})
	}
	return docs, nil
}

func (s *repositoryDocumentService) backfillChronologyDocuments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
	type row struct {
		ID        uuid.UUID
		Documents string
	}
	var rows []row
	if err := db.Raw(`
		SELECT id, documents
		FROM case_chronologies
		WHERE documents IS NOT NULL AND documents != ''
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]entity.RepositoryDocument, 0)
	for _, r := range rows {
		var paths []string
		if err := json.Unmarshal([]byte(r.Documents), &paths); err != nil {
			continue
		}
		for _, p := range paths {
			if p == "" {
				continue
			}
			docs = append(docs, entity.RepositoryDocument{
				Title:        filepath.Base(p),
				FeatureCode:  "case_management",
				SourceType:   "case_chronologies",
				SourceID:     r.ID.String(),
				UploaderType: "requester",
				FilePath:     p,
				FileName:     filepath.Base(p),
				MIMEType:     detectMIME(p),
			})
		}
	}
	return docs, nil
}

func toRepositoryDocumentResponse(d entity.RepositoryDocument) dto.RepositoryDocumentResponse {
	return dto.RepositoryDocumentResponse{
		ID:           d.ID.String(),
		Title:        d.Title,
		FeatureCode:  d.FeatureCode,
		SourceType:   d.SourceType,
		SourceID:     d.SourceID,
		UploaderType: d.UploaderType,
		FileName:     d.FileName,
		MIMEType:     d.MIMEType,
		FileSize:     d.FileSize,
		IsActive:     d.IsActive,
		CreatedBy:    d.CreatedBy.String(),
		UpdatedBy:    d.UpdatedBy.String(),
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func toRepositoryDocumentResponses(docs []entity.RepositoryDocument) []dto.RepositoryDocumentResponse {
	result := make([]dto.RepositoryDocumentResponse, 0, len(docs))
	for i := range docs {
		result = append(result, toRepositoryDocumentResponse(docs[i]))
	}
	return result
}

func detectMIME(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

func uniqueStrings(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	result := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}