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

type LegalOpinionService interface {
	Create(userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error)
	GetByID(id string, userID string, role string) (*entity.LegalOpinion, error)
	GetAll(userID string, role string, query dto.LegalOpinionListQuery) ([]entity.LegalOpinion, int64, error)
	Update(id string, userID string, req dto.UpdateLegalOpinionRequest) (*entity.LegalOpinion, error)
	Delete(id string, userID string) error
	Resubmit(id string, userID string, files []*multipart.FileHeader) (*entity.LegalOpinion, error)
	UpdateStatus(id string, req dto.UpdateStatusRequest) error
	UploadResult(id string, adminID string, req dto.UploadResultRequest, file *multipart.FileHeader) error
	DownloadFile(filePath, userID, role string) (*minio.Object, string, error)
	GeneratePDF(id string) ([]byte, error)
	SaveDraft(userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error)
	UpdateDraft(id, userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error)
	DeleteDraftAttachment(id, attachmentID, userID string) error
	SubmitDraft(id, userID string) (*entity.LegalOpinion, error)
}

type legalOpinionService struct {
	repo    repository.LegalOpinionRepository
	storage *storage.MinIOClient
	pdfSvc  PDFService
	repoDoc RepositoryDocumentService
}

func NewLegalOpinionService(repo repository.LegalOpinionRepository, s *storage.MinIOClient, repoDoc RepositoryDocumentService) LegalOpinionService {
	return &legalOpinionService{repo: repo, storage: s, pdfSvc: NewPDFService(), repoDoc: repoDoc}
}

func (s *legalOpinionService) Create(userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	// Generate ticket number
	if err := validateFiles(files); err != nil {
		return nil, err
	}

	lo := &entity.LegalOpinion{UserID: uid, Status: entity.StatusSubmitted}
	isiLegalOpinion(lo, req)
	if err := denganNomorTiket(utils.PrefixLegalOpinion, s.repo.NextTicketSequence, func(nomor string) error {
		lo.TicketNumber = nomor
		return s.repo.Create(lo)
	}); err != nil {
		return nil, errors.New("gagal membuat pengajuan")
	}

	// Upload attachments
	if len(files) > 0 {
		if err := s.uploadAttachments(lo.ID, files, 1); err != nil {
			// Unggahan gagal di tengah (bukan karena file ditolak): pengajuan yang
			// baru dibuat dibatalkan, supaya user tidak mendapat pengajuan ganda
			// saat mengulang.
			_ = s.repo.Delete(lo.ID)
			return nil, err
		}
	}

	return s.muat(lo.ID)
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

	// Regular users can only see their own submissions (BR-001). Draf hanya untuk
	// pemiliknya, juga bagi yang boleh melihat semua pengajuan (PR Draf).
	if (!canAccessAllSubmissions(role) || lo.Status == entity.StatusDraft) && lo.UserID.String() != userID {
		return nil, errors.New("pengajuan tidak ditemukan")
	}

	return lo, nil
}

// GetAll: ?status=DRAFT selalu hanya memuat draf milik sendiri (PR Draf).
func (s *legalOpinionService) GetAll(userID string, role string, query dto.LegalOpinionListQuery) ([]entity.LegalOpinion, int64, error) {
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

	// Draf dihapus permanen beserta lampirannya (PR Draf).
	if lo.Status == entity.StatusDraft {
		paths := make([]string, 0, len(lo.Attachments))
		for _, a := range lo.Attachments {
			paths = append(paths, a.FilePath)
		}
		if err := s.repo.HardDelete(lo.ID); err != nil {
			return errors.New("gagal menghapus draf")
		}
		hapusFileDraf(s.storage, paths)
		return nil
	}

	// Only deletable when SUBMITTED (BR-005)
	if lo.Status != entity.StatusSubmitted {
		return errors.New("pengajuan hanya dapat dihapus saat berstatus SUBMITTED")
	}

	if err := s.repo.Delete(lo.ID); err != nil {
		return err
	}
	catatRepositori(lo.ID, s.repoDoc.RemoveSubmission(lo.ID.String()))
	return nil
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
	if err := validateFiles(files); err != nil {
		return nil, err
	}
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

	return s.muat(lo.ID)
}

func (s *legalOpinionService) UpdateStatus(id string, req dto.UpdateStatusRequest) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}

	lo, err := s.repo.FindByID(uid)
	if err != nil || lo.Status == entity.StatusDraft {
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
	if err != nil || lo.Status == entity.StatusDraft {
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

	s.muat(uid) // hasil kajian langsung masuk Repositori Dokumen
	return nil
}

// DownloadFile hanya melayani file yang terdaftar sebagai lampiran atau hasil
// kajian pengajuan modul ini, dan hanya untuk yang boleh melihat pengajuannya.
// Path lain, termasuk file modul lain, dijawab ErrFileNotFound. Nama unduhannya
// <tiket>_Lampiran_<nama asli> atau <tiket>_Hasil_<nama asli>.
func (s *legalOpinionService) DownloadFile(filePath, userID, role string) (*minio.Object, string, error) {
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

// ─── Draf (PR Draf) ───────────────────────────────────────────────────────────

// SaveDraft menyimpan draf baru. Isian boleh belum lengkap; lampiran tetap diperiksa.
func (s *legalOpinionService) SaveDraft(userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}
	if err := validateFiles(files); err != nil {
		return nil, err
	}
	lo := &entity.LegalOpinion{TicketNumber: nomorDraf(), UserID: uid, Status: entity.StatusDraft}
	isiLegalOpinion(lo, req)
	if err := s.repo.Create(lo); err != nil {
		return nil, errors.New("gagal menyimpan draf")
	}
	if err := s.uploadAttachments(lo.ID, files, 1); err != nil {
		_ = s.repo.HardDelete(lo.ID)
		return nil, err
	}
	return s.repo.FindByID(lo.ID)
}

// UpdateDraft mengganti isian draf dan menambah lampiran baru.
func (s *legalOpinionService) UpdateDraft(id, userID string, req dto.CreateLegalOpinionRequest, files []*multipart.FileHeader) (*entity.LegalOpinion, error) {
	lo, err := s.drafMilik(id, userID)
	if err != nil {
		return nil, err
	}
	if err := validateFiles(files); err != nil {
		return nil, err
	}
	isiLegalOpinion(lo, req)
	if err := s.repo.Update(lo); err != nil {
		return nil, errors.New("gagal menyimpan draf")
	}
	if err := s.uploadAttachments(lo.ID, files, 1); err != nil {
		return nil, err
	}
	return s.repo.FindByID(lo.ID)
}

// DeleteDraftAttachment menghapus permanen satu lampiran draf beserta filenya.
func (s *legalOpinionService) DeleteDraftAttachment(id, attachmentID, userID string) error {
	lo, err := s.drafMilik(id, userID)
	if err != nil {
		return err
	}
	for _, a := range lo.Attachments {
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
func (s *legalOpinionService) SubmitDraft(id, userID string) (*entity.LegalOpinion, error) {
	lo, err := s.drafMilik(id, userID)
	if err != nil {
		return nil, err
	}
	if !permintaanLegalOpinion(lo).Complete() {
		return nil, ErrIsianBelumLengkap
	}
	now := time.Now()
	lo.Status = entity.StatusSubmitted
	lo.CreatedAt = now
	lo.StatusUpdatedAt = &now
	if err := denganNomorTiket(utils.PrefixLegalOpinion, s.repo.NextTicketSequence, func(nomor string) error {
		lo.TicketNumber = nomor
		return s.repo.Update(lo)
	}); err != nil {
		return nil, errors.New("gagal mengirim draf")
	}
	return s.muat(lo.ID)
}

// muat membaca ulang pengajuan sesudah lampiran atau hasilnya bertambah, lalu mencatat
// semua filenya di Repositori Dokumen (PR A tahap A.4).
func (s *legalOpinionService) muat(id uuid.UUID) (*entity.LegalOpinion, error) {
	lo, err := s.repo.FindByID(id)
	if err == nil {
		catatRepositori(id, s.repoDoc.SyncLegalOpinion(lo))
	}
	return lo, err
}

// drafMilik mengambil draf milik pemanggil; selain itu ErrDrafTidakAda.
func (s *legalOpinionService) drafMilik(id, userID string) (*entity.LegalOpinion, error) {
	lo, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil || lo.Status != entity.StatusDraft {
		return nil, ErrDrafTidakAda
	}
	return lo, nil
}

func isiLegalOpinion(lo *entity.LegalOpinion, req dto.CreateLegalOpinionRequest) {
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
}

func permintaanLegalOpinion(lo *entity.LegalOpinion) dto.CreateLegalOpinionRequest {
	return dto.CreateLegalOpinionRequest{
		RequestorName: lo.RequestorName, RequestorPosition: lo.RequestorPosition, RequestorDivision: lo.RequestorDivision,
		RequestorEmail: lo.RequestorEmail, RequestorPhone: lo.RequestorPhone, LegalType: lo.LegalType,
		LegalTypeOther: lo.LegalTypeOther, Title: lo.Title, Chronology: lo.Chronology, Question: lo.Question,
	}
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
