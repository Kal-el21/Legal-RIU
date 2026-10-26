package service

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"
	"legal-riu-portal/internal/utils"

	"github.com/minio/minio-go/v7"
)

type RepositoryDocumentService interface {
	GetAllDocuments(access repository.DocumentAccess, filters dto.RepositoryDocumentFilters) (dto.RepositoryDocumentListResponse, error)
	GetDocumentByID(access repository.DocumentAccess, id string) (*dto.RepositoryDocumentResponse, error)
	DownloadDocument(access repository.DocumentAccess, id string) (*minio.Object, string, error)
	SyncCaseDocument(legalCase *entity.LegalCase, filePath string, fileName string, fileSize int64, uploadedAt time.Time) error
	SyncCasePhoto(legalCase *entity.LegalCase, filePath string, fileName string, fileSize int64, uploadedAt time.Time) error
	SyncCaseMetadata(legalCase *entity.LegalCase) error
	SyncChronologyDocument(legalCase *entity.LegalCase, chronology *entity.CaseChronology, filePath string, fileName string, uploadedAt *time.Time) error
	RemoveCaseDocuments(caseID string) error
	RemoveChronologyDocuments(chronID string) error
	SyncLegalOpinion(lo *entity.LegalOpinion) error
	SyncDocumentReview(dr *entity.DocumentReview) error
	SyncAgreement(d *entity.AgreementDocument) error
	RemoveSubmission(id string) error
}

type repositoryDocumentService struct {
	repo     repository.RepositoryDocumentRepository
	caseRepo repository.LegalCaseRepository
	storage  *storage.MinIOClient
}

func NewRepositoryDocumentService(repo repository.RepositoryDocumentRepository, caseRepo repository.LegalCaseRepository, s *storage.MinIOClient) RepositoryDocumentService {
	return &repositoryDocumentService{repo: repo, caseRepo: caseRepo, storage: s}
}

func (s *repositoryDocumentService) GetAllDocuments(access repository.DocumentAccess, filters dto.RepositoryDocumentFilters) (dto.RepositoryDocumentListResponse, error) {
	docs, total, err := s.repo.FindAll(access, filters)
	if err != nil {
		return dto.RepositoryDocumentListResponse{}, err
	}
	return dto.RepositoryDocumentListResponse{
		Items:      toRepositoryDocumentResponses(docs),
		Total:      total,
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: int((total + int64(filters.Limit) - 1) / int64(filters.Limit)),
	}, nil
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
		SourceNumber: doc.SourceNumber,
		SourceTitle:  doc.SourceTitle,
		UploadedAt:   doc.UploadedAt,
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
	return obj, s.namaUnduhan(doc), nil
}

func (s *repositoryDocumentService) namaUnduhan(doc *entity.RepositoryDocument) string {
	switch doc.SourceType {
	case "legal_opinion_attachments", "document_review_attachments", "agreement_attachments":
		return utils.NamaUnduhan(stringPointer(doc.SourceNumber), "Lampiran", doc.FileName)
	case "legal_opinion_results", "document_review_results":
		return utils.NamaUnduhan(stringPointer(doc.SourceNumber), "Hasil", doc.FileName)
	case "agreement_documents":
		return doc.FileName
	case "legal_cases", "legal_cases_photo", "case_chronologies":
		if tiket, jenis, nama := s.caseRepo.NamaFilePerkara(doc.FilePath); jenis != "" {
			if nama == "" {
				nama = filepath.Base(doc.FilePath)
			}
			return utils.NamaUnduhan(tiket, jenis, nama)
		}
	}
	return doc.FileName
}

func (s *repositoryDocumentService) SyncCaseDocument(legalCase *entity.LegalCase, filePath string, fileName string, fileSize int64, uploadedAt time.Time) error {
	if filePath == "" {
		return nil
	}
	doc := entity.RepositoryDocument{
		Title:        fileName,
		FeatureCode:  "case_management",
		SourceType:   "legal_cases",
		SourceID:     legalCase.ID.String(),
		SourceNumber: pointer(legalCase.TicketNumber),
		SourceTitle:  pointer(legalCase.CaseName),
		UploadedAt:   &uploadedAt,
		UploaderType: "requester",
		FilePath:     filePath,
		FileName:     fileName,
		FileSize:     fileSize,
		MIMEType:     detectMIME(fileName),
		IsActive:     true,
	}
	return s.repo.UpsertBySource(doc)
}

func (s *repositoryDocumentService) SyncCasePhoto(legalCase *entity.LegalCase, filePath string, fileName string, fileSize int64, uploadedAt time.Time) error {
	if filePath == "" {
		return nil
	}
	doc := entity.RepositoryDocument{
		Title:        fileName,
		FeatureCode:  "case_management",
		SourceType:   "legal_cases_photo",
		SourceID:     legalCase.ID.String(),
		SourceNumber: pointer(legalCase.TicketNumber),
		SourceTitle:  pointer(legalCase.CaseName),
		UploadedAt:   &uploadedAt,
		UploaderType: "requester",
		FilePath:     filePath,
		FileName:     fileName,
		FileSize:     fileSize,
		MIMEType:     detectMIME(fileName),
		IsActive:     true,
	}
	return s.repo.UpsertBySource(doc)
}

func (s *repositoryDocumentService) SyncCaseMetadata(legalCase *entity.LegalCase) error {
	var docs []entity.RepositoryDocument
	for _, source := range []struct {
		sourceType string
		filePath   string
	}{
		{"legal_cases", legalCase.DocumentLink},
		{"legal_cases_photo", legalCase.Photo},
	} {
		if source.filePath == "" {
			continue
		}
		fileName := filepath.Base(source.filePath)
		docs = append(docs, entity.RepositoryDocument{
			Title:        fileName,
			FeatureCode:  "case_management",
			SourceType:   source.sourceType,
			SourceID:     legalCase.ID.String(),
			SourceNumber: pointer(legalCase.TicketNumber),
			SourceTitle:  pointer(legalCase.CaseName),
			UploaderType: "requester",
			FilePath:     source.filePath,
			FileName:     fileName,
			MIMEType:     detectMIME(fileName),
			IsActive:     true,
		})
	}
	if err := s.repo.BatchUpsert(docs); err != nil {
		return err
	}
	return s.repo.UpdateCaseMetadata(legalCase.ID.String(), legalCase.TicketNumber, legalCase.CaseName)
}

// SyncChronologyDocument preserves existing file names/timestamps when reactivating old paths.
func (s *repositoryDocumentService) SyncChronologyDocument(legalCase *entity.LegalCase, chronology *entity.CaseChronology, filePath string, fileName string, uploadedAt *time.Time) error {
	if filePath == "" {
		return nil
	}
	doc := entity.RepositoryDocument{
		Title:        filepath.Base(filePath),
		FeatureCode:  "case_management",
		SourceType:   "case_chronologies",
		SourceID:     chronology.ID.String(),
		SourceNumber: pointer(legalCase.TicketNumber),
		SourceTitle:  pointer(legalCase.CaseName),
		UploadedAt:   uploadedAt,
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
	nomor, judul := lo.TicketNumber, lo.Title
	var docs []entity.RepositoryDocument
	for _, a := range lo.Attachments {
		docs = append(docs, berkasPengajuan("legal_opinion", "legal_opinion_attachments", id, "requester", a.FilePath, a.FileName, a.FileSize, nomor, judul, &a.CreatedAt))
	}
	for _, r := range lo.Results {
		docs = append(docs, berkasPengajuan("legal_opinion", "legal_opinion_results", id, "approver", r.FilePath, r.FileName, 0, nomor, judul, &r.CreatedAt))
	}
	return s.repo.BatchUpsert(docs)
}

func (s *repositoryDocumentService) SyncDocumentReview(dr *entity.DocumentReview) error {
	id := dr.ID.String()
	nomor, judul := dr.TicketNumber, dr.DocumentName
	var docs []entity.RepositoryDocument
	for _, a := range dr.Attachments {
		docs = append(docs, berkasPengajuan("document_review", "document_review_attachments", id, "requester", a.FilePath, a.FileName, a.FileSize, nomor, judul, &a.CreatedAt))
	}
	for _, r := range dr.Results {
		docs = append(docs, berkasPengajuan("document_review", "document_review_results", id, "approver", r.FilePath, r.FileName, 0, nomor, judul, &r.CreatedAt))
	}
	return s.repo.BatchUpsert(docs)
}

// SyncAgreement mencatat lampiran perjanjian, dan DOCX serta PDF final kalau sudah disetujui.
func (s *repositoryDocumentService) SyncAgreement(d *entity.AgreementDocument) error {
	id := d.ID.String()
	data, err := decodeForm(d.FormData)
	if err != nil {
		return err
	}
	nomor := d.TicketNumber
	judul, _ := data["pihak_kedua_nama"].(string)
	judul = strings.TrimSpace(judul)
	var docs []entity.RepositoryDocument
	for _, a := range d.Attachments {
		docs = append(docs, berkasPengajuan("agreement_document", "agreement_attachments", id, "requester", a.FilePath, a.FileName, a.FileSize, nomor, judul, &a.CreatedAt))
	}
	for _, p := range []string{d.GeneratedDOCXPath, d.GeneratedPDFPath} {
		if p != "" {
			docs = append(docs, berkasPengajuan("agreement_document", "agreement_documents", id, "approver", p, d.GeneratedFileName+filepath.Ext(p), 0, nomor, judul, d.ApprovedAt))
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

func berkasPengajuan(fitur, sumber, id, pengunggah, path, nama string, ukuran int64, nomor, judul string, uploadedAt *time.Time) entity.RepositoryDocument {
	return entity.RepositoryDocument{Title: nama, FeatureCode: fitur, SourceType: sumber, SourceID: id, UploaderType: pengunggah,
		SourceNumber: pointer(nomor), SourceTitle: pointer(judul), UploadedAt: uploadedAt,
		FilePath: path, FileName: nama, FileSize: ukuran, MIMEType: detectMIME(nama), IsActive: true}
}

func toRepositoryDocumentResponse(d entity.RepositoryDocument) dto.RepositoryDocumentResponse {
	return dto.RepositoryDocumentResponse{
		ID:           d.ID.String(),
		Title:        d.Title,
		FeatureCode:  d.FeatureCode,
		SourceType:   d.SourceType,
		SourceID:     d.SourceID,
		SourceNumber: d.SourceNumber,
		SourceTitle:  d.SourceTitle,
		UploadedAt:   d.UploadedAt,
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

func pointer[T any](value T) *T {
	return &value
}

func stringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
