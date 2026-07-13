package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
)

type AgreementDocumentService interface {
	Create(userID string, req dto.CreateAgreementRequest, files []*multipart.FileHeader) (*entity.AgreementDocument, error)
	GetByID(id string, userID string, role string) (*entity.AgreementDocument, error)
	GetAll(userID string, role string, query dto.AgreementListQuery) ([]entity.AgreementDocument, int64, error)
	Update(id string, userID string, req dto.UpdateAgreementRequest) (*entity.AgreementDocument, error)
	Delete(id string, userID string) error
	Resubmit(id string, userID string, files []*multipart.FileHeader) (*entity.AgreementDocument, error)
	UpdatePihakPertama(id string, req dto.UpdatePihakPertamaRequest) (*entity.AgreementDocument, error)
	Approve(id string) (*entity.AgreementDocument, error)
	ReturnForRevision(id string, adminNote string) (*entity.AgreementDocument, error)
	Reject(id string, adminNote string) (*entity.AgreementDocument, error)
	UpdateMeta(id string, meta map[string]string) (*entity.AgreementDocument, error)
	GeneratePreview(ctx context.Context, id string, userID string, role string) ([]byte, error)
	GetDownloadURL(id string) (string, error)
	UploadAttachment(docID uuid.UUID, files []*multipart.FileHeader, round int) error
}

type agreementDocumentService struct {
	repo        repository.AgreementDocumentRepository
	storage     *storage.MinIOClient
	pdfSvc      PDFService
	templateSvc TemplateConversionService
}

func NewAgreementDocumentService(repo repository.AgreementDocumentRepository, s *storage.MinIOClient, fieldPositionRepo repository.TemplateFieldPositionRepository) AgreementDocumentService {
	return &agreementDocumentService{repo: repo, storage: s, pdfSvc: NewPDFService(NewTemplateConversionService(s), fieldPositionRepo), templateSvc: NewTemplateConversionService(s)}
}
func toFormData(req dto.CreateAgreementRequest) map[string]interface{} {
	return map[string]interface{}{
		"nomor_pihak_kedua":        req.NomorPihakKedua,
		"tempat_ttd":               req.TempatTtd,
		"tanggal_ttd":              req.TanggalTtd,
		"pihak_kedua_nama":         req.PihakKeduaNama,
		"pihak_kedua_bidang":       req.PihakKeduaBidang,
		"pihak_kedua_alamat":       req.PihakKeduaAlamat,
		"pihak_kedua_telepon":      req.PihakKeduaTelepon,
		"pihak_kedua_email":        req.PihakKeduaEmail,
		"pihak_kedua_pic":          req.PihakKeduaPic,
		"pihak_kedua_pejabat":      req.PihakKeduaPejabat,
		"pihak_kedua_jabatan":      req.PihakKeduaJabatan,
		"jenis_pekerjaan":          req.JenisPekerjaan,
		"surat_penawaran_nomor":    req.SuratPenawaranNomor,
		"surat_penawaran_perihal":  req.SuratPenawaranPerihal,
		"surat_penawaran_tanggal":  req.SuratPenawaranTanggal,
		"surat_penunjukan_nomor":   req.SuratPenunjukanNomor,
		"surat_penunjukan_perihal": req.SuratPenunjukanPerihal,
		"surat_penunjukan_tanggal": req.SuratPenunjukanTanggal,
		"ruang_lingkup":            req.RuangLingkup,
		"jangka_waktu_mulai":       req.JangkaWaktuMulai,
		"jangka_waktu_selesai":     req.JangkaWaktuSelesai,
		"nilai_kontrak":            req.NilaiKontrak,
		"termin1_persen":           req.Termin1Persen,
		"termin1_nilai":            req.Termin1Nilai,
		"termin2_persen":           req.Termin2Persen,
		"termin2_nilai":            req.Termin2Nilai,
		"bank":                     req.Bank,
		"nomor_rekening":           req.NomorRekening,
		"atas_nama":                req.AtasNama,
		"lampiran":                 req.Lampiran,
	}
}

func marshalFormData(m map[string]interface{}) (json.RawMessage, error) {
	if m == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(m)
}

func unmarshalFormData(raw json.RawMessage) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	return m, nil
}

func formDataString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return formatNumber(val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatNumber(n float64) string {
	return strings.ReplaceAll(fmt.Sprintf("%.2f", n), ".", ",")
}

func formatRupiah(n float64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%.2f", n)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	rev := []rune{}
	for i, c := range reverseString(intPart) {
		if i > 0 && i%3 == 0 {
			rev = append(rev, '.')
		}
		rev = append(rev, c)
	}
	intFormatted := reverseString(string(rev))
	result := "Rp. " + intFormatted
	if len(parts) > 1 {
		result += "," + parts[1]
	}
	if neg {
		return "-" + result
	}
	return result
}

func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func terbilang(n float64) string {
	units := []string{"", "satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan"}
	scales := []string{"", "ribu", "juta", "miliar", "triliun"}

	if n == 0 {
		return "nol"
	}

	remaining := int64(n)
	if remaining < 0 {
		remaining = -remaining
	}

	var groups [5]int64
	for i := 0; remaining > 0 && i < 5; i++ {
		groups[i] = remaining % 1000
		remaining /= 1000
	}

	var words []string
	for i := 4; i >= 0; i-- {
		g := groups[i]
		if g == 0 {
			continue
		}
		hundreds := g / 100
		tens := (g % 100) / 10
		ones := g % 10

		var groupWords []string
		if hundreds > 0 {
			if hundreds == 1 {
				groupWords = append(groupWords, "seratus")
			} else {
				groupWords = append(groupWords, units[hundreds], "ratus")
			}
		}
		if tens > 1 {
			groupWords = append(groupWords, units[tens]+" puluh")
			if ones > 0 {
				groupWords = append(groupWords, units[ones])
			}
		} else if tens == 1 {
			switch ones {
			case 0:
				groupWords = append(groupWords, "sepuluh")
			case 1:
				groupWords = append(groupWords, "sebelas")
			case 2:
				groupWords = append(groupWords, "dua belas")
			case 3:
				groupWords = append(groupWords, "tiga belas")
			case 4:
				groupWords = append(groupWords, "empat belas")
			case 5:
				groupWords = append(groupWords, "lima belas")
			case 6:
				groupWords = append(groupWords, "enam belas")
			case 7:
				groupWords = append(groupWords, "tujuh belas")
			case 8:
				groupWords = append(groupWords, "delapan belas")
			case 9:
				groupWords = append(groupWords, "sembilan belas")
			}
		} else if ones > 0 {
			groupWords = append(groupWords, units[ones])
		}

		groupStr := strings.Join(groupWords, " ")
		if i > 0 && scales[i] != "" {
			groupStr += " " + scales[i]
		}
		words = append(words, groupStr)
	}
	return strings.Join(words, " ")
}

func (s *agreementDocumentService) Create(userID string, req dto.CreateAgreementRequest, files []*multipart.FileHeader) (*entity.AgreementDocument, error) {
	uid, err := parseUUID(userID)
	if err != nil {
		return nil, errors.New("user tidak valid")
	}

	master, err := s.repo.GetFirstActiveCompanyMaster()
	if err != nil {
		return nil, errors.New("data perusahaan (Pihak Pertama) belum tersedia")
	}

	count, err := s.repo.CountByMonthAndPrefix("PK")
	if err != nil {
		return nil, errors.New("gagal generate ticket number")
	}
	ticket := utils.GenerateTicketNumber(utils.PrefixAgreement, int(count)+1)

	formData := toFormData(req)
	if formData["tempat_ttd"] == nil || formData["tempat_ttd"] == "" {
		formData["tempat_ttd"] = master.DefaultTempatTtd
	}

	agreementNumber, err := s.generateAgreementNumber()
	if err != nil {
		return nil, errors.New("gagal generate nomor perjanjian")
	}
	formData["nomor_pihak_pertama"] = agreementNumber
	templateVersion := "1"
	formData["template_version"] = templateVersion

	formDataJSON, err := marshalFormData(formData)
	if err != nil {
		return nil, errors.New("gagal menyimpan data form")
	}

	doc := &entity.AgreementDocument{
		TicketNumber:   ticket,
		UserID:         uid,
		PihakPertamaID: master.ID,
		FormData:       formDataJSON,
		Status:         entity.StatusSubmitted,
	}

	if err := s.repo.Create(doc); err != nil {
		return nil, errors.New("gagal membuat pengajuan")
	}

	if len(files) > 0 {
		if err := s.UploadAttachment(doc.ID, files, 1); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(doc.ID)
}

func (s *agreementDocumentService) generateAgreementNumber() (string, error) {
	count, err := s.repo.CountByMonthAndPrefix("PK")
	if err != nil {
		return "", err
	}
	seq := int(count) + 1
	now := time.Now()
	return fmt.Sprintf("%03d/RM.01.01/HR/IndonesiaRe/%d", seq, now.Year()), nil
}

func (s *agreementDocumentService) GetByID(id string, userID string, role string) (*entity.AgreementDocument, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}

	if !canAccessAllSubmissions(role) && doc.UserID.String() != userID {
		return nil, errors.New("dokumen tidak ditemukan")
	}

	// Confidentiality: Pihak Pertama master data must not leak to the requester
	// before the agreement is approved/finalized.
	if !canAccessAllSubmissions(role) && doc.Status != entity.StatusCompleted {
		doc.PihakPertama = entity.CompanyMaster{}
		doc.PihakPertamaPejabat = ""
		doc.PihakPertamaJabatan = ""
	}

	return doc, nil
}

func (s *agreementDocumentService) GetAll(userID string, role string, query dto.AgreementListQuery) ([]entity.AgreementDocument, int64, error) {
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

func (s *agreementDocumentService) Update(id string, userID string, req dto.UpdateAgreementRequest) (*entity.AgreementDocument, error) {
	doc, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return nil, err
	}
	if doc.Status != entity.StatusSubmitted && doc.Status != entity.StatusNeedRevision && doc.Status != entity.StatusRejected {
		return nil, errors.New("dokumen hanya dapat diedit saat berstatus SUBMITTED, NEED_REVISION, atau REJECTED")
	}
	formData := toFormData(req)
	formDataJSON, err := marshalFormData(formData)
	if err != nil {
		return nil, errors.New("gagal menyimpan data form")
	}
	doc.FormData = formDataJSON
	if err := s.repo.Update(doc); err != nil {
		return nil, errors.New("gagal mengupdate dokumen")
	}
	return s.repo.FindByID(doc.ID)
}

func (s *agreementDocumentService) Delete(id string, userID string) error {
	doc, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return err
	}
	if doc.Status != entity.StatusSubmitted && doc.Status != entity.StatusRejected {
		return errors.New("dokumen hanya dapat dihapus saat berstatus SUBMITTED atau REJECTED")
	}
	return s.repo.Delete(doc.ID)
}

func (s *agreementDocumentService) Resubmit(id string, userID string, files []*multipart.FileHeader) (*entity.AgreementDocument, error) {
	doc, err := s.GetByID(id, userID, string(entity.RoleUser))
	if err != nil {
		return nil, err
	}
	if doc.Status != entity.StatusNeedRevision && doc.Status != entity.StatusRejected {
		return nil, errors.New("dokumen hanya dapat diajukan ulang dari status NEED_REVISION atau REJECTED")
	}
	if len(files) > 0 {
		round, _ := s.repo.GetLatestUploadRound(doc.ID)
		if err := s.UploadAttachment(doc.ID, files, round+1); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateStatus(doc.ID, entity.StatusResubmitted, ""); err != nil {
		return nil, errors.New("gagal mengubah status")
	}
	return s.repo.FindByID(doc.ID)
}

func (s *agreementDocumentService) UpdatePihakPertama(id string, req dto.UpdatePihakPertamaRequest) (*entity.AgreementDocument, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	doc.PihakPertamaPejabat = req.PihakPertamaPejabat
	doc.PihakPertamaJabatan = req.PihakPertamaJabatan
	if err := s.repo.Update(doc); err != nil {
		return nil, errors.New("gagal mengupdate data Pihak Pertama")
	}
	return s.repo.FindByID(doc.ID)
}

func (s *agreementDocumentService) Approve(id string) (*entity.AgreementDocument, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if doc.Status != entity.StatusSubmitted && doc.Status != entity.StatusResubmitted && doc.Status != entity.StatusUnderReview {
		return nil, errors.New("dokumen tidak dalam status yang dapat disetujui")
	}

	pdfBytes, err := s.pdfSvc.GenerateFinalAgreementPDF(context.Background(), doc)
	if err != nil {
		return nil, errors.New("gagal generate dokumen PDF")
	}

	ctx := context.Background()
	objectPath, fileName, err := s.storage.UploadBytes(ctx, "agreement-documents", pdfBytes, "application/pdf", "perjanjian-"+doc.TicketNumber)
	if err != nil {
		return nil, errors.New("gagal mengupload dokumen final")
	}

	doc.GeneratedPDFPath = objectPath
	doc.GeneratedFileName = fileName
	if err := s.repo.UpdateStatus(uid, entity.StatusCompleted, ""); err != nil {
		return nil, errors.New("gagal mengubah status")
	}
	// Persist the PDF path on the document.
	upd := map[string]interface{}{"generated_pdf_path": objectPath, "generated_file_name": fileName}
	if err := s.repo.UpdateFields(uid, upd); err != nil {
		return nil, errors.New("gagal menyimpan dokumen final")
	}
	return s.repo.FindByID(uid)
}

func (s *agreementDocumentService) ReturnForRevision(id string, adminNote string) (*entity.AgreementDocument, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	if _, err := s.repo.FindByID(uid); err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if err := s.repo.UpdateStatus(uid, entity.StatusNeedRevision, adminNote); err != nil {
		return nil, errors.New("gagal mengubah status")
	}
	return s.repo.FindByID(uid)
}

func (s *agreementDocumentService) Reject(id string, adminNote string) (*entity.AgreementDocument, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	if _, err := s.repo.FindByID(uid); err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if err := s.repo.UpdateStatus(uid, entity.StatusRejected, adminNote); err != nil {
		return nil, errors.New("gagal mengubah status")
	}
	return s.repo.FindByID(uid)
}

func (s *agreementDocumentService) UpdateMeta(id string, meta map[string]string) (*entity.AgreementDocument, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if doc.Status == entity.StatusCompleted || doc.Status == entity.StatusRejected {
		return nil, errors.New("dokumen tidak dapat diubah pada status ini")
	}

	formData, err := unmarshalFormData(doc.FormData)
	if err != nil {
		return nil, errors.New("gagal membaca data form")
	}

	if v, ok := meta["nomor_pihak_pertama"]; ok && v != "" {
		formData["nomor_pihak_pertama"] = v
	}
	if v, ok := meta["tempat_ttd"]; ok {
		formData["tempat_ttd"] = v
	}
	if v, ok := meta["tanggal_ttd"]; ok {
		formData["tanggal_ttd"] = v
	}
	if v, ok := meta["pihak_pertama_pejabat"]; ok {
		doc.PihakPertamaPejabat = v
	}
	if v, ok := meta["pihak_pertama_jabatan"]; ok {
		doc.PihakPertamaJabatan = v
	}

	formDataJSON, err := marshalFormData(formData)
	if err != nil {
		return nil, errors.New("gagal menyimpan data form")
	}
	doc.FormData = formDataJSON

	if err := s.repo.Update(doc); err != nil {
		return nil, errors.New("gagal mengupdate data")
	}
	return s.repo.FindByID(doc.ID)
}

func (s *agreementDocumentService) GeneratePreview(ctx context.Context, id string, userID string, role string) ([]byte, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	if !canAccessAllSubmissions(role) && doc.UserID.String() != userID {
		return nil, errors.New("dokumen tidak ditemukan")
	}
	return s.pdfSvc.GenerateAgreementPreview(ctx, doc)
}

func (s *agreementDocumentService) GetDownloadURL(id string) (string, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return "", errors.New("ID tidak valid")
	}
	doc, err := s.repo.FindByID(uid)
	if err != nil {
		return "", errors.New("dokumen tidak ditemukan")
	}
	if doc.Status != entity.StatusCompleted || doc.GeneratedPDFPath == "" {
		return "", errors.New("dokumen final belum tersedia")
	}
	return s.storage.GetPresignedURL(context.Background(), doc.GeneratedPDFPath)
}

func (s *agreementDocumentService) UploadAttachment(docID uuid.UUID, files []*multipart.FileHeader, round int) error {
	ctx := context.Background()
	for _, file := range files {
		objectPath, fileName, err := s.storage.UploadFile(ctx, "agreement-documents/attachments", file, "agreement-att-"+docID.String())
		if err != nil {
			return errors.New("gagal mengupload file: " + file.Filename)
		}
		att := &entity.AgreementAttachment{
			AgreementID: docID,
			FileName:    fileName,
			FilePath:    objectPath,
			FileSize:    file.Size,
			UploadRound: round,
		}
		if err := s.repo.AddAttachment(att); err != nil {
			return errors.New("gagal menyimpan metadata file")
		}
	}
	return nil
}
