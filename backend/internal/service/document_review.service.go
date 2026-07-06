package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type DocumentReviewService interface {
	Create(userID string, req dto.CreateDocumentReviewRequest, files []*multipart.FileHeader) (*entity.DocumentReview, error)
	GetByID(id string, userID string, role string) (*entity.DocumentReview, error)
	GetAll(userID string, role string, query dto.DocumentReviewListQuery) ([]entity.DocumentReview, int64, error)
	Update(id string, userID string, req dto.UpdateDocumentReviewRequest) (*entity.DocumentReview, error)
	Delete(id string, userID string) error
	Resubmit(id string, userID string, files []*multipart.FileHeader) (*entity.DocumentReview, error)
	UpdateStatus(id string, req dto.UpdateStatusRequest) error
	UploadResult(id string, adminID string, req dto.UploadResultRequest, file *multipart.FileHeader) error
	GetPresignedURL(filePath string) (string, error)
	DownloadFile(filePath string) (*minio.Object, error)

	GeneratePDF(id string) ([]byte, error)
}

type documentReviewService struct {
	repo    repository.DocumentReviewRepository
	storage *storage.MinIOClient
	pdfSvc  PDFService
}

func NewDocumentReviewService(repo repository.DocumentReviewRepository, s *storage.MinIOClient) DocumentReviewService {
	return &documentReviewService{repo: repo, storage: s, pdfSvc: NewPDFService()}
}

func (s *documentReviewService) Create(userID string, req dto.CreateDocumentReviewRequest, files []*multipart.FileHeader) (*entity.DocumentReview, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	count, err := s.repo.CountByMonthAndPrefix("RD")
	if err != nil {
		return nil, errors.New("gagal generate ticket number")
	}
	ticket := utils.GenerateTicketNumber(utils.PrefixDocumentReview, int(count)+1)

	dr := &entity.DocumentReview{
		TicketNumber:      ticket,
		UserID:            uid,
		RequestorName:     req.RequestorName,
		RequestorPosition: req.RequestorPosition,
		RequestorDivision: req.RequestorDivision,
		RequestorEmail:    req.RequestorEmail,
		RequestorPhone:    req.RequestorPhone,
		DocumentName:      req.DocumentName,
		SecondParty:       req.SecondParty,
		ThirdParty:        req.ThirdParty,
		DocumentType:      req.DocumentType,
		DocumentTypeOther: req.DocumentTypeOther,
		AdditionalNote:    req.AdditionalNote,
		Status:            entity.StatusSubmitted,
	}

	if err := s.repo.Create(dr); err != nil {
		return nil, errors.New("gagal membuat pengajuan")
	}

	if len(files) > 0 {
		if err := s.uploadAttachments(dr.ID, files, 1); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(dr.ID)
}

func (s *documentReviewService) GetByID(id string, userID string, role string) (*entity.DocumentReview, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	dr, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	if !canAccessAllSubmissions(role) && dr.UserID.String() != userID {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	return dr, nil
}

func (s *documentReviewService) GetAll(userID string, role string, query dto.DocumentReviewListQuery) ([]entity.DocumentReview, int64, error) {
	var filterUserID *uuid.UUID
	if !canAccessAllSubmissions(role) {
		uid, err := parseUUID(userID)
		if err != nil {
			return nil, 0, errors.New("user tidak valid")
		}
		filterUserID = &uid
	}
	return s.repo.FindAll(filterUserID, query.Status, query.Page, query.Limit)
}

func (s *documentReviewService) Update(id string, userID string, req dto.UpdateDocumentReviewRequest) (*entity.DocumentReview, error) {
	dr, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return nil, err
	}

	if dr.Status != entity.StatusSubmitted {
		return nil, errors.New("pengajuan hanya dapat diedit saat berstatus SUBMITTED")
	}

	dr.RequestorName = req.RequestorName
	dr.RequestorPosition = req.RequestorPosition
	dr.RequestorDivision = req.RequestorDivision
	dr.RequestorEmail = req.RequestorEmail
	dr.RequestorPhone = req.RequestorPhone
	dr.DocumentName = req.DocumentName
	dr.SecondParty = req.SecondParty
	dr.ThirdParty = req.ThirdParty
	dr.DocumentType = req.DocumentType
	dr.DocumentTypeOther = req.DocumentTypeOther
	dr.AdditionalNote = req.AdditionalNote

	if err := s.repo.Update(dr); err != nil {
		return nil, errors.New("gagal mengupdate pengajuan")
	}
	return s.repo.FindByID(dr.ID)
}

func (s *documentReviewService) Delete(id string, userID string) error {
	dr, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return err
	}

	if dr.Status != entity.StatusSubmitted {
		return errors.New("pengajuan hanya dapat dihapus saat berstatus SUBMITTED")
	}

	return s.repo.Delete(dr.ID)
}

func (s *documentReviewService) Resubmit(id string, userID string, files []*multipart.FileHeader) (*entity.DocumentReview, error) {
	dr, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return nil, err
	}

	if dr.Status != entity.StatusNeedRevision && dr.Status != entity.StatusRejected {
		return nil, errors.New("pengajuan hanya dapat diajukan ulang dari status NEED_REVISION atau REJECTED")
	}

	if len(files) > 0 {
		round, _ := s.repo.GetLatestUploadRound(dr.ID)
		if err := s.uploadAttachments(dr.ID, files, round+1); err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateStatus(dr.ID, entity.StatusResubmitted, ""); err != nil {
		return nil, errors.New("gagal mengubah status")
	}

	return s.repo.FindByID(dr.ID)
}

func (s *documentReviewService) UpdateStatus(id string, req dto.UpdateStatusRequest) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}

	dr, err := s.repo.FindByID(uid)
	if err != nil {
		return errors.New("pengajuan tidak ditemukan")
	}

	newStatus := entity.SubmissionStatus(req.Status)
	if !isValidStatusTransition(dr.Status, newStatus) {
		return errors.New("perubahan status tidak valid")
	}

	return s.repo.UpdateStatus(uid, newStatus, req.AdminNote)
}

func (s *documentReviewService) UploadResult(id string, adminID string, req dto.UploadResultRequest, file *multipart.FileHeader) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	adminUID, err := parseUUID(adminID)
	if err != nil {
		return errors.New("admin tidak valid")
	}

	ctx := context.Background()
	objectPath, fileName, err := s.storage.UploadFile(ctx, "document-reviews/results", file, fmt.Sprintf("review-result-%s", uid.String()))
	if err != nil {
		return errors.New("gagal mengupload hasil review")
	}

	result := &entity.DocumentReviewResult{
		DocumentReviewID: uid,
		UploadedBy:       adminUID,
		FileName:         fileName,
		FilePath:         objectPath,
		Notes:            req.Notes,
	}
	return s.repo.AddResult(result)
}

func (s *documentReviewService) GetPresignedURL(filePath string) (string, error) {
	return s.storage.GetPresignedURL(context.Background(), filePath)
}

func (s *documentReviewService) DownloadFile(filePath string) (*minio.Object, error) {
	return s.storage.GetFileObject(context.Background(), filePath)
}

func (s *documentReviewService) uploadAttachments(drID uuid.UUID, files []*multipart.FileHeader, round int) error {
	ctx := context.Background()
	for _, file := range files {
		objectPath, fileName, err := s.storage.UploadFile(ctx, "document-reviews/attachments", file, fmt.Sprintf("review-att-%s", drID.String()))
		if err != nil {
			return errors.New("gagal mengupload file: " + file.Filename)
		}
		att := &entity.DocumentReviewAttachment{
			DocumentReviewID: drID,
			FileName:         fileName,
			FilePath:         objectPath,
			FileSize:         file.Size,
			UploadRound:      round,
		}
		if err := s.repo.AddAttachment(att); err != nil {
			return errors.New("gagal menyimpan metadata file")
		}
	}
	return nil
}

func (s *documentReviewService) GeneratePDF(id string) ([]byte, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	dr, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	return s.pdfSvc.GenerateDocumentReviewPDF(dr)
}
