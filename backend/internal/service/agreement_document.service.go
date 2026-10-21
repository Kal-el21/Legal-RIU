package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"slices"
	"strings"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type AgreementDocumentService interface {
	ListTypes() []dto.AgreementTypeResponse
	GetType(string) (*dto.AgreementTypeResponse, error)
	DocumentSchema(id, userID string, all bool) (*dto.AgreementTypeResponse, error)
	Create(string, dto.CreateAgreementDocumentRequest, []*multipart.FileHeader) (*entity.AgreementDocument, error)
	GetAll(string, bool, dto.AgreementListQuery) ([]entity.AgreementDocument, int64, error)
	GetByID(string, string, bool) (*entity.AgreementDocument, error)
	Detail(string, string, bool) (*entity.AgreementDocument, error)
	Update(string, string, dto.UpdateAgreementDocumentRequest) (*entity.AgreementDocument, error)
	Delete(string, string) error
	Resubmit(string, string, []*multipart.FileHeader) (*entity.AgreementDocument, error)
	UpdateMeta(string, dto.AgreementMetaRequest) (*entity.AgreementDocument, error)
	UpdateStatus(context.Context, string, string, dto.AgreementStatusRequest) (*entity.AgreementDocument, error)
	Preview(context.Context, string, string, bool) ([]byte, error)
	PreviewDraft(context.Context, string, dto.PreviewDraftRequest) ([]byte, error)
	GetGeneratedFile(string, string, bool, string) (*minio.Object, string, error)
	GetAttachment(string, string, string, bool) (*minio.Object, string, string, error)
	GetMaster() (*entity.AgreementCompanyMaster, error)
	UpdateMaster(dto.AgreementCompanyMasterRequest) (*entity.AgreementCompanyMaster, error)
	ListSignatories() ([]entity.AgreementSignatory, error)
	CreateSignatory(dto.AgreementSignatoryRequest) (*entity.AgreementSignatory, error)
	UpdateSignatory(id string, req dto.AgreementSignatoryRequest) (lama, baru *entity.AgreementSignatory, e error)
	DeleteSignatory(id string) (*entity.AgreementSignatory, error)
}

type agreementDocumentService struct {
	repo      repository.AgreementDocumentRepository
	storage   *storage.MinIOClient
	templates AgreementTemplateService
	generator *AgreementGenerator
	converter DOCXConverter
}

func NewAgreementDocumentService(r repository.AgreementDocumentRepository, s *storage.MinIOClient, templates AgreementTemplateService) AgreementDocumentService {
	return &agreementDocumentService{r, s, templates, NewAgreementGenerator(), DOCXConverter{}}
}

// ErrPengajuanTidakAda: pengajuan tidak ada atau bukan milik pengguna.
var ErrPengajuanTidakAda = errors.New("pengajuan tidak ditemukan")

// ErrTemplateUsang: form dikirim dengan versi template yang sudah bukan versi
// aktif jenis dokumennya. Dijawab 409 supaya frontend memuat field versi baru.
var ErrTemplateUsang = errors.New("Template sudah diperbarui Admin. Periksa kembali isian Anda.")

// templateFor selalu mengutamakan versi yang terkunci di dokumen. Pengajuan yang
// sedang mengantre review karena itu tidak ikut berubah ketika admin
// mengaktifkan versi template baru.
func (s *agreementDocumentService) templateFor(ctx context.Context, d *entity.AgreementDocument) (*entity.AgreementTemplate, []byte, error) {
	if d.TemplateID != nil {
		return s.templates.TemplateByID(ctx, *d.TemplateID)
	}
	return s.templates.ActiveTemplate(ctx, strings.ToUpper(strings.TrimSpace(d.DocumentTypeCode)))
}

// sectionsFor membaca definisi field dari versi template yang dikunci dokumen.
func (s *agreementDocumentService) sectionsFor(d *entity.AgreementDocument) ([]AgreementSectionDefinition, error) {
	if d.TemplateID == nil {
		return nil, errors.New("pengajuan tidak terhubung ke versi template")
	}
	tpl, e := s.templates.Get(d.TemplateID.String())
	if e != nil {
		return nil, e
	}
	return templateSections(tpl)
}

// ListTypes hanya memuat jenis yang punya template aktif, karena hanya jenis
// itu yang bisa diajukan.
func (s *agreementDocumentService) ListTypes() []dto.AgreementTypeResponse {
	types, _ := s.templates.ActiveTypes()
	out := make([]dto.AgreementTypeResponse, 0, len(types))
	for _, t := range types {
		if _, _, e := s.templates.ActiveType(t.Code); e == nil {
			out = append(out, dto.AgreementTypeResponse{Code: t.Code, Name: t.Name})
		}
	}
	return out
}

// GetType mengembalikan form jenis dokumen dari versi template aktif. Versi itu
// (template_id) dikirim balik saat pengajuan dibuat.
func (s *agreementDocumentService) GetType(code string) (*dto.AgreementTypeResponse, error) {
	typ, tpl, e := s.templates.ActiveType(code)
	if e != nil {
		return nil, errors.New("tipe dokumen tidak ditemukan")
	}
	return skemaVersi(typ.Code, typ.Name, tpl)
}

// DocumentSchema: definisi field versi template yang dikunci pengajuan, karena
// pengajuan itu juga divalidasi dengan versi tersebut. Dipakai form revisi dan
// halaman detail; all untuk approver (semua pengajuan), selain itu hanya milik sendiri.
func (s *agreementDocumentService) DocumentSchema(id, userID string, all bool) (*dto.AgreementTypeResponse, error) {
	d, e := s.GetByID(id, userID, all)
	if e != nil {
		return nil, e
	}
	if d.TemplateID == nil {
		return nil, errors.New("pengajuan tidak terhubung ke versi template")
	}
	tpl, e := s.templates.Get(d.TemplateID.String())
	if e != nil {
		return nil, e
	}
	name := d.DocumentTypeCode
	if types, err := s.templates.ActiveTypes(); err == nil {
		for _, t := range types {
			if t.Code == d.DocumentTypeCode {
				name = t.Name
			}
		}
	}
	return skemaVersi(d.DocumentTypeCode, name, tpl)
}

// skemaVersi: form satu versi template, ditambah isian sistem di bagian pertama.
func skemaVersi(code, name string, tpl *entity.AgreementTemplate) (*dto.AgreementTypeResponse, error) {
	sections, e := templateSections(tpl)
	if e != nil {
		return nil, e
	}
	sections[0].Fields = append(sections[0].Fields, isianSistemPengaju...)
	schema, e := json.Marshal(map[string]interface{}{"code": code, "name": name, "template_id": tpl.ID, "template_version": tpl.Version, "sections": sections})
	if e != nil {
		return nil, e
	}
	return &dto.AgreementTypeResponse{Code: code, Name: name, Schema: schema}, nil
}

func (s *agreementDocumentService) Create(userID string, req dto.CreateAgreementDocumentRequest, files []*multipart.FileHeader) (*entity.AgreementDocument, error) {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return nil, errors.New("user tidak valid")
	}
	typ, tpl, e := s.templates.ActiveType(req.DocumentTypeCode)
	if e != nil {
		return nil, e
	}
	if strings.TrimSpace(req.TemplateID) == "" {
		return nil, errors.New("versi template wajib disertakan; muat ulang halaman pengajuan")
	}
	if req.TemplateID != tpl.ID.String() {
		return nil, ErrTemplateUsang
	}
	sections, e := templateSections(tpl)
	if e != nil {
		return nil, e
	}
	if e = validateAgreementForm(sections, req.FormData); e != nil {
		return nil, e
	}
	// Semua lampiran diperiksa sebelum pengajuan dibuat, supaya file yang ditolak
	// tidak meninggalkan pengajuan terhapus, nomor tiket hangus, atau lampiran separuh.
	if e = validateFiles(files); e != nil {
		return nil, e
	}
	raw, e := json.Marshal(req.FormData)
	if e != nil {
		return nil, e
	}
	seq, e := s.repo.NextTicketSequence("PK")
	if e != nil {
		return nil, e
	}
	doc := &entity.AgreementDocument{TicketNumber: utils.GenerateTicketNumber("PK", seq), UserID: uid, DocumentTypeCode: typ.Code, FormData: raw, AgreementNumber: nil, Status: entity.StatusSubmitted, TemplateID: &tpl.ID, TemplateChecksum: tpl.Checksum}
	if e = s.repo.Create(doc); e != nil {
		return nil, errors.New("gagal membuat pengajuan")
	}
	if e = s.uploadAttachments(doc, uid, files, 1); e != nil {
		_ = s.repo.Delete(doc.ID)
		return nil, errors.New("gagal mengunggah lampiran: " + e.Error())
	}
	return s.repo.FindByID(doc.ID)
}
func (s *agreementDocumentService) GetAll(userID string, all bool, q dto.AgreementListQuery) ([]entity.AgreementDocument, int64, error) {
	var owner *uuid.UUID
	if !all {
		id, e := uuid.Parse(userID)
		if e != nil {
			return nil, 0, errors.New("user tidak valid")
		}
		owner = &id
	}
	items, total, e := s.repo.FindAll(owner, q.Status, q.DateFrom, q.Search, q.Page, q.Limit)
	if e == nil && all {
		s.tandaiTemplateLama(items)
	}
	return items, total, e
}

// Detail: GetByID untuk halaman detail; approver ikut melihat penanda template lama.
func (s *agreementDocumentService) Detail(id, userID string, all bool) (*entity.AgreementDocument, error) {
	d, e := s.GetByID(id, userID, all)
	if e == nil && all {
		docs := []entity.AgreementDocument{*d}
		s.tandaiTemplateLama(docs)
		d.OldTemplate = docs[0].OldTemplate
	}
	return d, e
}

// tandaiTemplateLama mengisi OldTemplate pada pengajuan yang belum selesai dan
// masih dikunci ke versi selain versi aktif jenisnya. Persetujuannya membentuk
// dokumen final dengan versi lama itu, jadi approver perlu tahu.
func (s *agreementDocumentService) tandaiTemplateLama(docs []entity.AgreementDocument) {
	aktif := map[string]*entity.AgreementTemplate{}
	versi := map[uuid.UUID]*entity.AgreementTemplate{}
	for i := range docs {
		d := &docs[i]
		if d.TemplateID == nil || d.Status == entity.StatusCompleted || d.Status == entity.StatusRejected {
			continue
		}
		a, ada := aktif[d.DocumentTypeCode]
		if !ada {
			_, a, _ = s.templates.ActiveType(d.DocumentTypeCode)
			aktif[d.DocumentTypeCode] = a
		}
		if a == nil || a.ID == *d.TemplateID {
			continue
		}
		v, ada := versi[*d.TemplateID]
		if !ada {
			v, _ = s.templates.Get(d.TemplateID.String())
			versi[*d.TemplateID] = v
		}
		if v == nil {
			continue
		}
		oleh := ""
		if a.Activator != nil {
			oleh = a.Activator.FullName
		}
		d.OldTemplate = &entity.TemplateLama{Version: v.Version, ActiveVersion: a.Version, ActiveNote: a.Note, ActivatedAt: a.ActivatedAt, ActivatedBy: oleh}
	}
}
func (s *agreementDocumentService) GetByID(id, userID string, all bool) (*entity.AgreementDocument, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, errors.New("ID tidak valid")
	}
	d, e := s.repo.FindByID(uid)
	if e != nil {
		return nil, ErrPengajuanTidakAda
	}
	if !all && d.UserID.String() != userID {
		return nil, ErrPengajuanTidakAda
	}
	return d, nil
}
func (s *agreementDocumentService) Update(id, userID string, req dto.UpdateAgreementDocumentRequest) (*entity.AgreementDocument, error) {
	d, e := s.GetByID(id, userID, false)
	if e != nil {
		return nil, e
	}
	if d.Status != entity.StatusSubmitted && d.Status != entity.StatusNeedRevision {
		return nil, errors.New("pengajuan tidak dapat diubah pada status ini")
	}
	sections, e := s.sectionsFor(d)
	if e != nil {
		return nil, e
	}
	if e = validateAgreementForm(sections, req.FormData); e != nil {
		return nil, e
	}
	d.FormData, _ = json.Marshal(req.FormData)
	if e = s.repo.Save(d); e != nil {
		return nil, e
	}
	return s.repo.FindByID(d.ID)
}
func (s *agreementDocumentService) Delete(id, userID string) error {
	d, e := s.GetByID(id, userID, false)
	if e != nil {
		return e
	}
	if d.Status != entity.StatusSubmitted {
		return errors.New("hanya pengajuan SUBMITTED yang dapat dihapus")
	}
	return s.repo.Delete(d.ID)
}
func (s *agreementDocumentService) Resubmit(id, userID string, files []*multipart.FileHeader) (*entity.AgreementDocument, error) {
	d, e := s.GetByID(id, userID, false)
	if e != nil {
		return nil, e
	}
	if d.Status != entity.StatusNeedRevision {
		return nil, errors.New("pengajuan tidak dalam status NEED_REVISION")
	}
	if e = validateFiles(files); e != nil {
		return nil, e
	}
	round, _ := s.repo.LatestUploadRound(d.ID)
	if e = s.uploadAttachments(d, d.UserID, files, round+1); e != nil {
		return nil, e
	}
	d.Status = entity.StatusResubmitted
	now := time.Now()
	d.StatusUpdatedAt = &now
	d.ApproverNote = ""
	if e = s.repo.Save(d); e != nil {
		return nil, e
	}
	return s.repo.FindByID(d.ID)
}
func (s *agreementDocumentService) uploadAttachments(d *entity.AgreementDocument, u uuid.UUID, files []*multipart.FileHeader, round int) error {
	for _, f := range files {
		path, name, e := s.storage.UploadFile(context.Background(), "agreement-documents/"+d.ID.String()+"/attachments", f, "attachment")
		if e != nil {
			return e
		}
		a := &entity.AgreementAttachment{AgreementDocumentID: d.ID, FileName: name, FilePath: path, MIMEType: f.Header.Get("Content-Type"), FileSize: f.Size, UploadRound: round, UploadedBy: u}
		if e = s.repo.AddAttachment(a); e != nil {
			return e
		}
	}
	return nil
}

func (s *agreementDocumentService) UpdateMeta(id string, req dto.AgreementMetaRequest) (*entity.AgreementDocument, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, errors.New("ID tidak valid")
	}
	d, e := s.repo.FindByID(uid)
	if e != nil {
		return nil, errors.New("pengajuan tidak ditemukan")
	}
	if d.Status == entity.StatusCompleted || d.Status == entity.StatusRejected {
		return nil, errors.New("pengajuan sudah final")
	}
	data, e := decodeForm(d.FormData)
	if e != nil {
		return nil, e
	}
	if req.AgreementNumber != nil && strings.TrimSpace(*req.AgreementNumber) != "" {
		num := strings.TrimSpace(*req.AgreementNumber)
		taken, e := s.repo.AgreementNumberTaken(num, d.ID)
		if e != nil {
			return nil, errors.New("gagal memeriksa nomor perjanjian")
		}
		if taken {
			return nil, repository.ErrAgreementNumberTaken
		}
		d.AgreementNumber = &num
	}
	setOptional(data, "tempat_ttd", req.SigningPlace)
	setOptional(data, "tanggal_ttd", req.SigningDate)
	setOptional(data, "pihak_pertama_pejabat", req.PartyOneSignatoryName)
	setOptional(data, "pihak_pertama_jabatan", req.PartyOneSignatoryPosition)
	setOptional(data, "pihak_pertama_pic", req.Pic)
	setOptional(data, "pihak_pertama_telepon", req.Phone)
	setOptional(data, "pihak_pertama_email", req.Email)
	d.FormData, _ = json.Marshal(data)
	if e = s.repo.Save(d); e != nil {
		// Pesan database tidak diteruskan ke pengguna.
		if errors.Is(e, repository.ErrAgreementNumberTaken) {
			return nil, e
		}
		return nil, errors.New("gagal menyimpan metadata")
	}
	return s.repo.FindByID(d.ID)
}
func setOptional(m map[string]interface{}, k string, v *string) {
	if v != nil {
		m[k] = strings.TrimSpace(*v)
	}
}

func strOrDash(v *string) string {
	if v == nil || strings.TrimSpace(*v) == "" {
		return "-"
	}
	return strings.TrimSpace(*v)
}

func (s *agreementDocumentService) UpdateStatus(ctx context.Context, id, approverID string, req dto.AgreementStatusRequest) (*entity.AgreementDocument, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, errors.New("ID tidak valid")
	}
	aid, e := uuid.Parse(approverID)
	if e != nil {
		return nil, errors.New("approver tidak valid")
	}
	d, e := s.repo.FindByID(uid)
	if e != nil {
		return nil, errors.New("pengajuan tidak ditemukan")
	}
	target := entity.SubmissionStatus(req.Status)
	switch target {
	case entity.StatusUnderReview:
		if d.Status != entity.StatusSubmitted && d.Status != entity.StatusResubmitted {
			return nil, errors.New("transisi status tidak valid")
		}
	case entity.StatusNeedRevision, entity.StatusRejected:
		if d.Status != entity.StatusUnderReview {
			return nil, errors.New("transisi status tidak valid")
		}
		if strings.TrimSpace(req.Note) == "" {
			return nil, errors.New("catatan approver wajib diisi")
		}
	case entity.StatusCompleted:
		return s.approve(ctx, d, aid)
	default:
		return nil, errors.New("status tidak didukung")
	}
	d.Status = target
	d.ApproverNote = strings.TrimSpace(req.Note)
	now := time.Now()
	d.StatusUpdatedAt = &now
	if e = s.repo.Save(d); e != nil {
		return nil, e
	}
	return s.repo.FindByID(d.ID)
}
func (s *agreementDocumentService) approve(ctx context.Context, d *entity.AgreementDocument, aid uuid.UUID) (*entity.AgreementDocument, error) {
	if d.Status != entity.StatusUnderReview {
		return nil, errors.New("pengajuan belum dalam review")
	}
	sections, e := s.sectionsFor(d)
	if e != nil {
		return nil, e
	}
	data, e := decodeForm(d.FormData)
	if e != nil {
		return nil, e
	}
	// Isian pengajuan hanya bisa diubah pengaju, jadi approver diarahkan ke revisi.
	if e = validateAgreementForm(sections, data); e != nil {
		return nil, fmt.Errorf("%w. Kembalikan pengajuan ke pengaju untuk direvisi", e)
	}
	master, e := s.repo.GetActiveMaster()
	if e != nil {
		return nil, errors.New("Master Pihak Pertama belum tersedia")
	}
	snapshot := partyOneSnapshot(master, data)
	if valueString(data["tanggal_ttd"]) == "" {
		return nil, errors.New("tanggal tanda tangan wajib diisi approver")
	}
	// Tempat tidak wajib bagi pengaju, tetapi dokumen final tidak boleh tanpa tempat (keputusan G).
	if strings.TrimSpace(valueString(snapshot["signing_place"])) == "" {
		return nil, errors.New("Tempat penandatanganan wajib diisi: isi di data persetujuan, atau isi tempat bawaan di master Pihak Pertama")
	}
	tpl, source, e := s.templateFor(ctx, d)
	if e != nil {
		return nil, e
	}
	values := placeholderValues(d, sections, data, snapshot)
	docx, checksum, e := s.generator.Generate(source, values, GenerateOptions{})
	if e != nil {
		return nil, e
	}
	pdf, e := s.converter.ToPDF(ctx, docx)
	if e != nil {
		return nil, e
	}
	// Path unik per percobaan: kalau dua persetujuan berjalan bersamaan, yang kalah
	// hanya menghapus filenya sendiri, bukan file milik persetujuan yang berhasil.
	base := "agreement-documents/" + d.ID.String() + "/final/agreement-" + uuid.NewString()[:8]
	if e = s.storage.UploadBytes(ctx, base+".docx", docx, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"); e != nil {
		return nil, e
	}
	if e = s.storage.UploadBytes(ctx, base+".pdf", pdf, "application/pdf"); e != nil {
		s.storage.DeleteFile(ctx, base+".docx")
		return nil, e
	}
	snap, _ := json.Marshal(snapshot)
	now := time.Now()
	ok, e := s.repo.Complete(d.ID, entity.StatusUnderReview, map[string]interface{}{"status": entity.StatusCompleted, "status_updated_at": now, "party_one_snapshot": snap, "generated_docx_path": base + ".docx", "generated_pdf_path": base + ".pdf", "generated_file_name": "perjanjian-" + d.TicketNumber, "template_id": tpl.ID, "template_checksum": checksum, "approved_by": aid, "approved_at": now, "approver_note": ""})
	if e != nil || !ok {
		s.storage.DeleteFile(ctx, base+".docx")
		s.storage.DeleteFile(ctx, base+".pdf")
		if e != nil {
			return nil, e
		}
		return nil, errors.New("pengajuan sudah diproses approver lain")
	}
	return s.repo.FindByID(d.ID)
}

func (s *agreementDocumentService) Preview(ctx context.Context, id, userID string, all bool) ([]byte, error) {
	d, e := s.GetByID(id, userID, all)
	if e != nil {
		return nil, e
	}
	// Perjanjian Selesai selalu dipratinjau dari file finalnya. Kalau file itu hilang,
	// dokumen tidak dibentuk ulang: hasilnya bukan dokumen yang ditandatangani
	// (pilihan membentuk ulang sebagai salinan pembanding dicatat di
	// .docs/future-implementation/bentuk-ulang-perjanjian-final.md).
	if d.Status == entity.StatusCompleted {
		if d.GeneratedPDFPath == "" {
			return nil, ErrFileNotFound
		}
		obj, err := s.storage.GetFileObject(ctx, d.GeneratedPDFPath)
		if err != nil {
			return nil, err
		}
		defer obj.Close()
		return io.ReadAll(obj)
	}
	data, e := decodeForm(d.FormData)
	if e != nil {
		return nil, e
	}
	master, e := s.repo.GetActiveMaster()
	if e != nil {
		return nil, errors.New("Master Pihak Pertama belum tersedia")
	}
	_, source, e := s.templateFor(ctx, d)
	if e != nil {
		return nil, e
	}
	sections, e := s.sectionsFor(d)
	if e != nil {
		return nil, e
	}
	docx, _, e := s.generator.Generate(source, placeholderValues(d, sections, data, partyOneSnapshot(master, data)), GenerateOptions{Draft: true})
	if e != nil {
		return nil, e
	}
	return s.converter.ToPDF(ctx, docx)
}
// PreviewDraft membentuk PDF pratinjau dari isian yang belum dikirim (langkah
// Tinjau di wizard) tanpa menyimpan apa pun. Validasinya sama dengan saat Kirim,
// jadi yang terlihat di pratinjau sama dengan yang akan dikirim.
func (s *agreementDocumentService) PreviewDraft(ctx context.Context, userID string, req dto.PreviewDraftRequest) ([]byte, error) {
	if strings.TrimSpace(req.TemplateID) == "" {
		return nil, errors.New("versi template wajib disertakan; muat ulang halaman pengajuan")
	}
	doc := &entity.AgreementDocument{}
	var tpl *entity.AgreementTemplate
	if req.DocumentID != "" {
		d, e := s.GetByID(req.DocumentID, userID, false)
		if e != nil {
			return nil, e
		}
		if d.Status != entity.StatusSubmitted && d.Status != entity.StatusNeedRevision {
			return nil, errors.New("pengajuan tidak dapat diubah pada status ini")
		}
		// Revisi tetap memakai versi yang dikunci pengajuan itu.
		if d.TemplateID == nil || d.TemplateID.String() != req.TemplateID {
			return nil, ErrTemplateUsang
		}
		if tpl, e = s.templates.Get(req.TemplateID); e != nil {
			return nil, e
		}
		doc = d
	} else {
		_, aktif, e := s.templates.ActiveType(req.DocumentTypeCode)
		if e != nil {
			return nil, e
		}
		if aktif.ID.String() != req.TemplateID {
			return nil, ErrTemplateUsang
		}
		tpl = aktif
	}
	sections, e := templateSections(tpl)
	if e != nil {
		return nil, e
	}
	if e = validateAgreementForm(sections, req.FormData); e != nil {
		return nil, e
	}
	master, e := s.repo.GetActiveMaster()
	if e != nil {
		return nil, errors.New("Master Pihak Pertama belum tersedia")
	}
	_, source, e := s.templates.TemplateByID(ctx, tpl.ID)
	if e != nil {
		return nil, e
	}
	draf := *doc
	draf.Attachments = slices.Clone(doc.Attachments)
	for _, nama := range req.AttachmentNames {
		draf.Attachments = append(draf.Attachments, entity.AgreementAttachment{FileName: nama})
	}
	docx, _, e := s.generator.Generate(source, placeholderValues(&draf, sections, req.FormData, partyOneSnapshot(master, req.FormData)), GenerateOptions{Draft: true})
	if e != nil {
		return nil, e
	}
	return s.converter.ToPDF(ctx, docx)
}

func (s *agreementDocumentService) GetGeneratedFile(id, userID string, all bool, kind string) (*minio.Object, string, error) {
	d, e := s.GetByID(id, userID, all)
	if e != nil {
		return nil, "", e
	}
	if d.Status != entity.StatusCompleted {
		return nil, "", errors.New("dokumen final belum tersedia")
	}
	path, ext := d.GeneratedPDFPath, ".pdf"
	if kind == "docx" {
		if !all {
			return nil, "", errors.New("akses DOCX tidak diizinkan")
		}
		path, ext = d.GeneratedDOCXPath, ".docx"
	}
	obj, e := s.storage.GetFileObject(context.Background(), path)
	if errors.Is(e, ErrFileNotFound) {
		return nil, "", e
	}
	if e != nil {
		return nil, "", errors.New("dokumen final tidak tersedia")
	}
	return obj, d.GeneratedFileName + ext, e
}
func (s *agreementDocumentService) GetAttachment(id, attachmentID, userID string, all bool) (*minio.Object, string, string, error) {
	d, e := s.GetByID(id, userID, all)
	if e != nil {
		return nil, "", "", e
	}
	aid, e := uuid.Parse(attachmentID)
	if e != nil {
		return nil, "", "", errors.New("attachment tidak valid")
	}
	a, e := s.repo.FindAttachment(d.ID, aid)
	if e != nil {
		return nil, "", "", errors.New("attachment tidak ditemukan")
	}
	obj, e := s.storage.GetFileObject(context.Background(), a.FilePath)
	return obj, a.FileName, a.MIMEType, e
}
func (s *agreementDocumentService) GetMaster() (*entity.AgreementCompanyMaster, error) {
	return s.repo.GetActiveMaster()
}
func (s *agreementDocumentService) UpdateMaster(req dto.AgreementCompanyMasterRequest) (*entity.AgreementCompanyMaster, error) {
	m, e := s.repo.GetActiveMaster()
	if e != nil {
		m = &entity.AgreementCompanyMaster{IsActive: true}
	}
	m.Name = req.Name
	m.Address = req.Address
	m.NPWP = req.NPWP
	m.Phone = req.Phone
	m.Email = req.Email
	m.PIC = req.PIC
	m.DefaultSignatoryName = req.DefaultSignatoryName
	m.DefaultSignatoryPosition = req.DefaultSignatoryPosition
	m.DefaultSigningPlace = req.DefaultSigningPlace
	m.DefaultAgreementNumber = req.DefaultAgreementNumber
	if e = s.repo.SaveMaster(m); e != nil {
		return nil, e
	}
	return m, nil
}

func decodeForm(raw json.RawMessage) (map[string]interface{}, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var m map[string]interface{}
	e := dec.Decode(&m)
	return m, e
}
func partyOneSnapshot(m *entity.AgreementCompanyMaster, data map[string]interface{}) map[string]interface{} {
	name := valueString(data["pihak_pertama_pejabat"])
	if name == "" {
		name = m.DefaultSignatoryName
	}
	pos := valueString(data["pihak_pertama_jabatan"])
	if pos == "" {
		pos = m.DefaultSignatoryPosition
	}
	place := valueString(data["tempat_ttd"])
	if place == "" {
		place = m.DefaultSigningPlace
	}
	pic := valueString(data["pihak_pertama_pic"])
	if pic == "" {
		pic = m.PIC
	}
	phone := valueString(data["pihak_pertama_telepon"])
	if phone == "" {
		phone = m.Phone
	}
	email := valueString(data["pihak_pertama_email"])
	if email == "" {
		email = m.Email
	}
	return map[string]interface{}{"name": m.Name, "address": m.Address, "phone": phone, "email": email, "pic": pic, "signatory_name": name, "signatory_position": pos, "signing_place": place}
}
