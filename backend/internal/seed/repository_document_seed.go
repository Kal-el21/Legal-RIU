package seed

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"time"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func BackfillRepositoryDocuments(db *gorm.DB) error {
	repo := repository.NewRepositoryDocumentRepository(db)

	docs := make([]entity.RepositoryDocument, 0)

	loAtts, err := backfillLegalOpinionAttachments(db)
	if err != nil {
		return err
	}
	docs = append(docs, loAtts...)

	loResults, err := backfillLegalOpinionResults(db)
	if err != nil {
		return err
	}
	docs = append(docs, loResults...)

	drAtts, err := backfillDocumentReviewAttachments(db)
	if err != nil {
		return err
	}
	docs = append(docs, drAtts...)

	drResults, err := backfillDocumentReviewResults(db)
	if err != nil {
		return err
	}
	docs = append(docs, drResults...)

	agAtts, err := backfillAgreementAttachments(db)
	if err != nil {
		return err
	}
	docs = append(docs, agAtts...)

	agGenerated, err := backfillAgreementGeneratedFiles(db)
	if err != nil {
		return err
	}
	docs = append(docs, agGenerated...)

	caseDocs, err := backfillLegalCaseDocuments(db)
	if err != nil {
		return err
	}
	docs = append(docs, caseDocs...)

	casePhotos, err := backfillLegalCasePhotos(db)
	if err != nil {
		return err
	}
	docs = append(docs, casePhotos...)

	chronDocs, err := backfillChronologyDocuments(db)
	if err != nil {
		return err
	}
	docs = append(docs, chronDocs...)

	if len(docs) == 0 {
		return nil
	}

	now := time.Now()
	for i := range docs {
		docs[i].UpdatedAt = now
	}

	if err := repo.MarkInactiveBySource("all", nil); err != nil {
		return err
	}

	sourceTypes := make(map[string][]string)
	for _, doc := range docs {
		sourceTypes[doc.SourceType] = append(sourceTypes[doc.SourceType], doc.SourceID)
	}
	for sourceType, sourceIDs := range sourceTypes {
		uniqueIDs := uniqueStrings(sourceIDs)
		if err := repo.MarkInactiveBySource(sourceType, uniqueIDs); err != nil {
			return err
		}
	}

	if err := repo.BatchUpsert(docs); err != nil {
		return err
	}

	return nil
}

func backfillLegalOpinionAttachments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillLegalOpinionResults(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillDocumentReviewAttachments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillDocumentReviewResults(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillAgreementAttachments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillAgreementGeneratedFiles(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillLegalCaseDocuments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillLegalCasePhotos(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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

func backfillChronologyDocuments(db *gorm.DB) ([]entity.RepositoryDocument, error) {
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