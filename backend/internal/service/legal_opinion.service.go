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

type LegalOpinionService interface {
	Create(userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error)
	GetByID(id string, userID string, role string) (*entity.LegalOpinion, error)
	GetAll(userID string, role string, query dto.LegalOpinionListQuery) ([]entity.LegalOpinion, int64, error)
	Update(id string, userID string, req dto.UpdateLegalOpinionRequest) (*entity.LegalOpinion, error)
	Delete(id string, userID string) error
	Resubmit(id string, userID string, files []*multipart.FileHeader) (*entity.LegalOpinion, error)
	UpdateStatus(id string, req dto.UpdateStatusRequest) error
	UploadResult(id string, adminID string, req dto.UploadResultRequest, file *multipart.FileHeader) error
	GetPresignedURL(filePath string) (string, error)
	DownloadFile(filePath string) (*minio.Object, error)
	GeneratePDF(id string) ([]byte, error)
}

type legalOpinionService struct {
	repo    repository.LegalOpinionRepository
	storage *storage.MinIOClient
	pdfSvc  PDFService
}

func NewLegalOpinionService(repo repository.LegalOpinionRepository, s *storage.MinIOClient) LegalOpinionService {
	return &legalOpinionService{repo: repo, storage: s, pdfSvc: NewPDFService()}
}

func (s *legalOpinionService) Create(userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	// Generate ticket number
	count, err := s.repo.CountByMonthAndPrefix("LO")
	if err != nil {
		return nil, errors.New("gagal generate ticket number")
	}
	ticket := utils.GenerateTicketNumber(utils.PrefixLegalOpinion, int(count)+1)

	lo := &entity.LegalOpinion{
		TicketNumber:      ticket,
		UserID:            uid,
		RequestorName:     req.RequestorName,
		RequestorPosition: req.RequestorPosition,
		RequestorDivision: req.RequestorDivision,
		RequestorEmail:    req.RequestorEmail,
		RequestorPhone:    req.RequestorPhone,
		LegalType:         req.LegalType,
		LegalTypeOther:    req.LegalTypeOther,
		Title:             req.Title,
		Chronology:        req.Chronology,
		Question:          req.Question,
		Status:            entity.StatusSubmitted,
	}

	if err := s.repo.Create(lo); err != nil {
		return nil, errors.New("gagal membuat pengajuan")
	}

	// Upload attachments
	if len(files) > 0 {
		if err := s.uploadAttachments(lo.ID, files, 1); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(lo.ID)
}

func (s *legalOpinionService) GetByID(id string, userID string, role string) (*entity.LegalOpinion, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	lo, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	// Regular users can only see their own submissions (BR-001).
	if !canAccessAllSubmissions(role) && lo.UserID.String() != userID {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	return lo, nil
}

func (s *legalOpinionService) GetAll(userID string, role string, query dto.LegalOpinionListQuery) ([]entity.LegalOpinion, int64, error) {
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

func (s *legalOpinionService) Update(id string, userID string, req dto.UpdateLegalOpinionRequest) (*entity.LegalOpinion, error) {
	lo, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return nil, err
	}

	// Only editable when SUBMITTED (BR-005)
	if lo.Status != entity.StatusSubmitted {
		return nil, errors.New("pengajuan hanya dapat diedit saat berstatus SUBMITTED")
	}

	lo.RequestorName = req.RequestorName
	lo.RequestorPosition = req.RequestorPosition
	lo.RequestorDivision = req.RequestorDivision
	lo.RequestorEmail = req.RequestorEmail
	lo.RequestorPhone = req.RequestorPhone
	lo.LegalType = req.LegalType
	lo.LegalTypeOther = req.LegalTypeOther
	lo.Title = req.Title
	lo.Chronology = req.Chronology
	lo.Question = req.Question

	if err := s.repo.Update(lo); err != nil {
		return nil, errors.New("gagal mengupdate pengajuan")
	}
	return s.repo.FindByID(lo.ID)
}

func (s *legalOpinionService) Delete(id string, userID string) error {
	lo, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return err
	}

	// Only deletable when SUBMITTED (BR-005)
	if lo.Status != entity.StatusSubmitted {
		return errors.New("pengajuan hanya dapat dihapus saat berstatus SUBMITTED")
	}

	return s.repo.Delete(lo.ID)
}

func (s *legalOpinionService) Resubmit(id string, userID string, files []*multipart.FileHeader) (*entity.LegalOpinion, error) {
	lo, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return nil, err
	}

	// Only resubmittable from NEED_REVISION or REJECTED (BR-006, BR-007)
	if lo.Status != entity.StatusNeedRevision && lo.Status != entity.StatusRejected {
		return nil, errors.New("pengajuan hanya dapat diajukan ulang dari status NEED_REVISION atau REJECTED")
	}

	// Upload new attachments as next round (append — file lama tetap)
	if len(files) > 0 {
		round, _ := s.repo.GetLatestUploadRound(lo.ID)
		if err := s.uploadAttachments(lo.ID, files, round+1); err != nil {
			return nil, err
		}
	}

	// Change status to RESUBMITTED
	if err := s.repo.UpdateStatus(lo.ID, entity.StatusResubmitted, ""); err != nil {
		return nil, errors.New("gagal mengubah status")
	}

	return s.repo.FindByID(lo.ID)
}

func (s *legalOpinionService) UpdateStatus(id string, req dto.UpdateStatusRequest) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}

	lo, err := s.repo.FindByID(uid)
	if err != nil {
		return errors.New("pengajuan tidak ditemukan")
	}

	newStatus := entity.SubmissionStatus(req.Status)
	if !isValidStatusTransition(lo.Status, newStatus) {
		return errors.New("perubahan status tidak valid")
	}

	return s.repo.UpdateStatus(uid, newStatus, req.AdminNote)
}

func (s *legalOpinionService) UploadResult(id string, adminID string, req dto.UploadResultRequest, file *multipart.FileHeader) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	adminUID, err := parseUUID(adminID)
	if err != nil {
		return errors.New("admin tidak valid")
	}

	// Get current legal opinion to check status
	lo, err := s.repo.FindByID(uid)
	if err != nil {
		return errors.New("pengajuan tidak ditemukan")
	}

	ctx := context.Background()
	objectPath, fileName, err := s.storage.UploadFile(ctx, "legal-opinions/results", file, fmt.Sprintf("opinion-result-%s", lo.ID.String()))
	if err != nil {
		return errors.New("gagal mengupload hasil kajian")
	}

	result := &entity.LegalOpinionResult{
		LegalOpinionID: uid,
		UploadedBy:     adminUID,
		FileName:       fileName,
		FilePath:       objectPath,
		Notes:          req.Notes,
	}
	if err := s.repo.AddResult(result); err != nil {
		return errors.New("gagal menyimpan hasil kajian")
	}

	// Auto-complete: If status is UNDER_REVIEW, automatically set to COMPLETED
	if lo.Status == entity.StatusUnderReview {
		if err := s.repo.UpdateStatus(uid, entity.StatusCompleted, ""); err != nil {
			return errors.New("gagal mengupdate status ke COMPLETED")
		}
	}

	return nil
}

func (s *legalOpinionService) GetPresignedURL(filePath string) (string, error) {
	return s.storage.GetPresignedURL(context.Background(), filePath)
}

func (s *legalOpinionService) DownloadFile(filePath string) (*minio.Object, error) {
	return s.storage.GetFileObject(context.Background(), filePath)
}

func (s *legalOpinionService) GeneratePDF(id string) ([]byte, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	lo, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	return s.pdfSvc.GenerateLegalOpinionPDF(lo)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *legalOpinionService) uploadAttachments(loID uuid.UUID, files []*multipart.FileHeader, round int) error {
	ctx := context.Background()
	for _, file := range files {
		objectPath, fileName, err := s.storage.UploadFile(ctx, "legal-opinions/attachments", file, fmt.Sprintf("opinion-att-%s", loID.String()))
		if err != nil {
			return errors.New("gagal mengupload file: " + file.Filename)
		}
		att := &entity.LegalOpinionAttachment{
			LegalOpinionID: loID,
			FileName:       fileName,
			FilePath:       objectPath,
			FileSize:       file.Size,
			UploadRound:    round,
		}
		if err := s.repo.AddAttachment(att); err != nil {
			return errors.New("gagal menyimpan metadata file")
		}
	}
	return nil
}

// Status machine — valid transitions
func isValidStatusTransition(current, next entity.SubmissionStatus) bool {
	transitions := map[entity.SubmissionStatus][]entity.SubmissionStatus{
		entity.StatusSubmitted:    {entity.StatusUnderReview},
		entity.StatusUnderReview:  {entity.StatusNeedRevision, entity.StatusRejected, entity.StatusCompleted},
		entity.StatusNeedRevision: {entity.StatusUnderReview},
		entity.StatusRejected:     {entity.StatusUnderReview},
		entity.StatusResubmitted:  {entity.StatusUnderReview},
		entity.StatusCompleted:    {},
	}
	for _, valid := range transitions[current] {
		if valid == next {
			return true
		}
	}
	return false
}
