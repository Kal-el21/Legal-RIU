package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"time"

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
	DownloadFile(filePath, userID, role string) (*minio.Object, string, error)

	GeneratePDF(id string) ([]byte, error)
	SaveDraft(userID string, req dto.CreateDocumentReviewRequest, files []*multipart.FileHeader) (*entity.DocumentReview, error)
	UpdateDraft(id, userID string, req dto.CreateDocumentReviewRequest, files []*multipart.FileHeader) (*entity.DocumentReview, error)
	DeleteDraftAttachment(id, attachmentID, userID string) error
	SubmitDraft(id, userID string) (*entity.DocumentReview, error)
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

	if err := validateFiles(files); err != nil {
		return nil, err
	}

	dr := &entity.DocumentReview{UserID: uid, Status: entity.StatusSubmitted}
	isiDocumentReview(dr, req)
	if err := denganNomorTiket(utils.PrefixDocumentReview, s.repo.NextTicketSequence, func(nomor string) error {
		dr.TicketNumber = nomor
		return s.repo.Create(dr)
	}); err != nil {
		return nil, errors.New("gagal membuat pengajuan")
	}

	if len(files) > 0 {
		if err := s.uploadAttachments(dr.ID, files, 1); err != nil {
			// Unggahan gagal di tengah (bukan karena file ditolak): pengajuan yang
			// baru dibuat dibatalkan, supaya user tidak mendapat pengajuan ganda
			// saat mengulang.
			_ = s.repo.Delete(dr.ID)
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

	// Draf hanya untuk pemiliknya, juga bagi yang boleh melihat semua pengajuan (PR Draf).
	if (!canAccessAllSubmissions(role) || dr.Status == entity.StatusDraft) && dr.UserID.String() != userID {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	return dr, nil
}

// GetAll: ?status=DRAFT selalu hanya memuat draf milik sendiri (PR Draf).
func (s *documentReviewService) GetAll(userID string, role string, query dto.DocumentReviewListQuery) ([]entity.DocumentReview, int64, error) {
	var filterUserID *uuid.UUID
	if !canAccessAllSubmissions(role) || query.Status == string(entity.StatusDraft) {
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

	// Draf dihapus permanen beserta lampirannya (PR Draf).
	if dr.Status == entity.StatusDraft {
		paths := make([]string, 0, len(dr.Attachments))
		for _, a := range dr.Attachments {
			paths = append(paths, a.FilePath)
		}
		if err := s.repo.HardDelete(dr.ID); err != nil {
			return errors.New("gagal menghapus draf")
		}
		hapusFileDraf(s.storage, paths)
		return nil
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

	if err := validateFiles(files); err != nil {
		return nil, err
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
	if err != nil || dr.Status == entity.StatusDraft {
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

	// Get current document review to check status
	dr, err := s.repo.FindByID(uid)
	if err != nil || dr.Status == entity.StatusDraft {
		return errors.New("pengajuan tidak ditemukan")
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
	if err := s.repo.AddResult(result); err != nil {
		return errors.New("gagal menyimpan hasil review")
	}

	// Auto-complete: If status is UNDER_REVIEW, automatically set to COMPLETED
	if dr.Status == entity.StatusUnderReview {
		if err := s.repo.UpdateStatus(uid, entity.StatusCompleted, ""); err != nil {
			return errors.New("gagal mengupdate status ke COMPLETED")
		}
	}

	return nil
}

// DownloadFile hanya melayani file yang terdaftar sebagai lampiran atau hasil
// kajian pengajuan modul ini, dan hanya untuk yang boleh melihat pengajuannya.
// Path lain, termasuk file modul lain, dijawab ErrFileNotFound. Nama unduhannya
// <tiket>_Lampiran_<nama asli> atau <tiket>_Hasil_<nama asli>.
func (s *documentReviewService) DownloadFile(filePath, userID, role string) (*minio.Object, string, error) {
	id, nama, hasil, err := s.repo.FindFileByPath(filePath)
	if err != nil {
		return nil, "", ErrFileNotFound
	}
	p, err := s.GetByID(id.String(), userID, role)
	if err != nil {
		return nil, "", ErrFileNotFound
	}
	jenis := "Lampiran"
	if hasil {
		jenis = "Hasil"
	}
	if nama == "" {
		nama = path.Base(filePath)
	}
	obj, err := s.storage.GetFileObject(context.Background(), filePath)
	return obj, utils.NamaUnduhan(p.TicketNumber, jenis, nama), err
}

// ─── Draf (PR Draf) ───────────────────────────────────────────────────────────
// Sama dengan draf Legal Opinion; pembantu bersamanya ada di helpers.go.

// SaveDraft menyimpan draf baru. Isian boleh belum lengkap; lampiran tetap diperiksa.
func (s *documentReviewService) SaveDraft(userID string, req dto.CreateDocumentReviewRequest, files []*multipart.FileHeader) (*entity.DocumentReview, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}
	if err := validateFiles(files); err != nil {
		return nil, err
	}
	dr := &entity.DocumentReview{TicketNumber: nomorDraf(), UserID: uid, Status: entity.StatusDraft}
	isiDocumentReview(dr, req)
	if err := s.repo.Create(dr); err != nil {
		return nil, errors.New("gagal menyimpan draf")
	}
	if err := s.uploadAttachments(dr.ID, files, 1); err != nil {
		_ = s.repo.HardDelete(dr.ID)
		return nil, err
	}
	return s.repo.FindByID(dr.ID)
}

// UpdateDraft mengganti isian draf dan menambah lampiran baru.
func (s *documentReviewService) UpdateDraft(id, userID string, req dto.CreateDocumentReviewRequest, files []*multipart.FileHeader) (*entity.DocumentReview, error) {
	dr, err := s.drafMilik(id, userID)
	if err != nil {
		return nil, err
	}
	if err := validateFiles(files); err != nil {
		return nil, err
	}
	isiDocumentReview(dr, req)
	if err := s.repo.Update(dr); err != nil {
		return nil, errors.New("gagal menyimpan draf")
	}
	if err := s.uploadAttachments(dr.ID, files, 1); err != nil {
		return nil, err
	}
	return s.repo.FindByID(dr.ID)
}

// DeleteDraftAttachment menghapus permanen satu lampiran draf beserta filenya.
func (s *documentReviewService) DeleteDraftAttachment(id, attachmentID, userID string) error {
	dr, err := s.drafMilik(id, userID)
	if err != nil {
		return err
	}
	for _, a := range dr.Attachments {
		if a.ID.String() == attachmentID {
			if err := s.repo.DeleteAttachment(a.ID); err != nil {
				return errors.New("gagal menghapus lampiran")
			}
			hapusFileDraf(s.storage, []string{a.FilePath})
			return nil
		}
	}
	return ErrLampiranTidakAda
}

// SubmitDraft mengirim draf: isian wajib diperiksa seperti pengajuan langsung, lalu
// draf mendapat nomor tiket, status SUBMITTED, dan tanggal pengajuan saat ini (K2).
func (s *documentReviewService) SubmitDraft(id, userID string) (*entity.DocumentReview, error) {
	dr, err := s.drafMilik(id, userID)
	if err != nil {
		return nil, err
	}
	if !permintaanDocumentReview(dr).Complete() {
		return nil, ErrIsianBelumLengkap
	}
	now := time.Now()
	dr.Status = entity.StatusSubmitted
	dr.CreatedAt = now
	dr.StatusUpdatedAt = &now
	if err := denganNomorTiket(utils.PrefixDocumentReview, s.repo.NextTicketSequence, func(nomor string) error {
		dr.TicketNumber = nomor
		return s.repo.Update(dr)
	}); err != nil {
		return nil, errors.New("gagal mengirim draf")
	}
	return s.repo.FindByID(dr.ID)
}

// drafMilik mengambil draf milik pemanggil; selain itu ErrDrafTidakAda.
func (s *documentReviewService) drafMilik(id, userID string) (*entity.DocumentReview, error) {
	dr, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil || dr.Status != entity.StatusDraft {
		return nil, ErrDrafTidakAda
	}
	return dr, nil
}

func isiDocumentReview(dr *entity.DocumentReview, req dto.CreateDocumentReviewRequest) {
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
}

func permintaanDocumentReview(dr *entity.DocumentReview) dto.CreateDocumentReviewRequest {
	return dto.CreateDocumentReviewRequest{
		RequestorName: dr.RequestorName, RequestorPosition: dr.RequestorPosition, RequestorDivision: dr.RequestorDivision,
		RequestorEmail: dr.RequestorEmail, RequestorPhone: dr.RequestorPhone, DocumentName: dr.DocumentName,
		SecondParty: dr.SecondParty, ThirdParty: dr.ThirdParty, DocumentType: dr.DocumentType,
		DocumentTypeOther: dr.DocumentTypeOther, AdditionalNote: dr.AdditionalNote,
	}
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
