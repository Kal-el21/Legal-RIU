package service

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"

	"github.com/minio/minio-go/v7"
)

type RepositoryDocumentService interface {
	GetAllDocuments(access repository.DocumentAccess, featureCode string, search string) ([]dto.RepositoryDocumentResponse, error)
	GetDocumentByID(access repository.DocumentAccess, id string) (*dto.RepositoryDocumentResponse, error)
	DownloadDocument(access repository.DocumentAccess, id string) (*minio.Object, string, error)
	SyncCaseDocument(caseID string, filePath string, fileName string, fileSize int64) error
	SyncCasePhoto(caseID string, filePath string, fileName string, fileSize int64) error
	SyncChronologyDocument(chronID string, caseID string, filePath string, fileName string) error
	RemoveCaseDocuments(caseID string) error
	RemoveChronologyDocuments(chronID string) error
	SyncLegalOpinion(lo *entity.LegalOpinion) error
	SyncDocumentReview(dr *entity.DocumentReview) error
	SyncAgreement(d *entity.AgreementDocument) error
	RemoveSubmission(id string) error
}

type repositoryDocumentService struct {
	repo    repository.RepositoryDocumentRepository
	storage *storage.MinIOClient
}

func NewRepositoryDocumentService(repo repository.RepositoryDocumentRepository, s *storage.MinIOClient) RepositoryDocumentService {
	return &repositoryDocumentService{repo: repo, storage: s}
}

func (s *repositoryDocumentService) GetAllDocuments(access repository.DocumentAccess, featureCode string, search string) ([]dto.RepositoryDocumentResponse, error) {
	docs, err := s.repo.FindAll(access, featureCode, search)
	if err != nil {
		return nil, err
	}
	return toRepositoryDocumentResponses(docs), nil
}

func (s *repositoryDocumentService) GetDocumentByID(access repository.DocumentAccess, id string) (*dto.RepositoryDocumentResponse, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(access, uid)
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

func (s *repositoryDocumentService) DownloadDocument(access repository.DocumentAccess, id string) (*minio.Object, string, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, "", errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(access, uid)
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

// SyncChronologyDocument mencatat satu lampiran kronologi. fileName berisi nama asli untuk
// unggahan baru (PR A tahap A.4b). Untuk lampiran yang sudah ada fileName kosong, dan barisnya
// hanya diaktifkan lagi supaya nama aslinya tidak tertimpa nama objek.
func (s *repositoryDocumentService) SyncChronologyDocument(chronID string, caseID string, filePath string, fileName string) error {
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
	if fileName == "" {
		return s.repo.BatchUpsert([]entity.RepositoryDocument{doc})
	}
	doc.Title, doc.FileName = fileName, fileName
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

// Sinkron pengajuan (PR A tahap A.4): setiap kali lampiran atau hasil bertambah, semua file
// pengajuan dicatat ulang. Isi dan kuncinya sama dengan backfill saat server menyala
// (seed.BackfillRepositoryDocuments), jadi aman dipanggil berulang tanpa baris ganda.
// Dipanggil sesudah pengajuan bukan draf lagi; draf tidak masuk Repositori Dokumen.

func (s *repositoryDocumentService) SyncLegalOpinion(lo *entity.LegalOpinion) error {
	id := lo.ID.String()
	var docs []entity.RepositoryDocument
	for _, a := range lo.Attachments {
		docs = append(docs, berkasPengajuan("legal_opinion", "legal_opinion_attachments", id, "requester", a.FilePath, a.FileName, a.FileSize))
	}
	for _, r := range lo.Results {
		docs = append(docs, berkasPengajuan("legal_opinion", "legal_opinion_results", id, "approver", r.FilePath, r.FileName, 0))
	}
	return s.repo.BatchUpsert(docs)
}

func (s *repositoryDocumentService) SyncDocumentReview(dr *entity.DocumentReview) error {
	id := dr.ID.String()
	var docs []entity.RepositoryDocument
	for _, a := range dr.Attachments {
		docs = append(docs, berkasPengajuan("document_review", "document_review_attachments", id, "requester", a.FilePath, a.FileName, a.FileSize))
	}
	for _, r := range dr.Results {
		docs = append(docs, berkasPengajuan("document_review", "document_review_results", id, "approver", r.FilePath, r.FileName, 0))
	}
	return s.repo.BatchUpsert(docs)
}

// SyncAgreement mencatat lampiran perjanjian, dan DOCX serta PDF final kalau sudah disetujui.
func (s *repositoryDocumentService) SyncAgreement(d *entity.AgreementDocument) error {
	id := d.ID.String()
	var docs []entity.RepositoryDocument
	for _, a := range d.Attachments {
		docs = append(docs, berkasPengajuan("agreement_document", "agreement_attachments", id, "requester", a.FilePath, a.FileName, a.FileSize))
	}
	for _, p := range []string{d.GeneratedDOCXPath, d.GeneratedPDFPath} {
		if p != "" {
			docs = append(docs, berkasPengajuan("agreement_document", "agreement_documents", id, "approver", p, d.GeneratedFileName+filepath.Ext(p), 0))
		}
	}
	return s.repo.BatchUpsert(docs)
}

// RemoveSubmission menonaktifkan semua file pengajuan LO, RD, atau perjanjian yang dihapus.
// ID pengajuan berupa UUID, jadi tidak tertukar antarjenis sumber.
func (s *repositoryDocumentService) RemoveSubmission(id string) error {
	for _, sumber := range []string{"legal_opinion_attachments", "legal_opinion_results", "document_review_attachments",
		"document_review_results", "agreement_attachments", "agreement_documents"} {
		if err := s.repo.MarkInactiveBySource(sumber, []string{id}); err != nil {
			return err
		}
	}
	return nil
}

func berkasPengajuan(fitur, sumber, id, pengunggah, path, nama string, ukuran int64) entity.RepositoryDocument {
	return entity.RepositoryDocument{Title: nama, FeatureCode: fitur, SourceType: sumber, SourceID: id, UploaderType: pengunggah,
		FilePath: path, FileName: nama, FileSize: ukuran, MIMEType: detectMIME(nama), IsActive: true}
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
