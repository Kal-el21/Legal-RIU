package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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
	Create(string, dto.CreateAgreementDocumentRequest, []*multipart.FileHeader) (*entity.AgreementDocument, error)
	GetAll(string, bool, dto.AgreementListQuery) ([]entity.AgreementDocument, int64, error)
	GetByID(string, string, bool) (*entity.AgreementDocument, error)
	Update(string, string, dto.UpdateAgreementDocumentRequest) (*entity.AgreementDocument, error)
	Delete(string, string) error
	Resubmit(string, string, []*multipart.FileHeader) (*entity.AgreementDocument, error)
	UpdateMeta(string, dto.AgreementMetaRequest) (*entity.AgreementDocument, error)
	UpdateStatus(context.Context, string, string, dto.AgreementStatusRequest) (*entity.AgreementDocument, error)
	Preview(context.Context, string, string, bool) ([]byte, error)
	GetGeneratedFile(string, string, bool, string) (*minio.Object, string, error)
	GetAttachment(string, string, string, bool) (*minio.Object, string, string, error)
	GetMaster() (*entity.AgreementCompanyMaster, error)
	UpdateMaster(dto.AgreementCompanyMasterRequest) (*entity.AgreementCompanyMaster, error)
}

type agreementDocumentService struct {
	repo      repository.AgreementDocumentRepository
	storage   *storage.MinIOClient
	registry  *AgreementRegistry
	generator *AgreementGenerator
	converter DOCXConverter
}

func NewAgreementDocumentService(r repository.AgreementDocumentRepository, s *storage.MinIOClient, registry *AgreementRegistry) AgreementDocumentService {
	return &agreementDocumentService{r, s, registry, NewAgreementGenerator(), DOCXConverter{}}
}

func (s *agreementDocumentService) ListTypes() []dto.AgreementTypeResponse {
	defs := s.registry.List()
	out := make([]dto.AgreementTypeResponse, 0, len(defs))
	for _, d := range defs {
		out = append(out, dto.AgreementTypeResponse{Code: d.Code, Name: d.Name})
	}
	return out
}
func (s *agreementDocumentService) GetType(code string) (*dto.AgreementTypeResponse, error) {
	d, ok := s.registry.Get(code)
	if !ok {
		return nil, errors.New("tipe dokumen tidak ditemukan")
	}
	return &dto.AgreementTypeResponse{Code: d.Code, Name: d.Name, Schema: d.RawSchema}, nil
}

func (s *agreementDocumentService) definitionFor(code string) (AgreementTypeDefinition, error) {
	def, ok := s.registry.Get(code)
	if !ok {
		return AgreementTypeDefinition{}, fmt.Errorf("tipe dokumen %q tidak terdaftar", strings.TrimSpace(code))
	}
	if len(def.Template) == 0 {
		return AgreementTypeDefinition{}, fmt.Errorf("template untuk tipe dokumen %q tidak tersedia", def.Code)
	}
	return def, nil
}

func (s *agreementDocumentService) Create(userID string, req dto.CreateAgreementDocumentRequest, files []*multipart.FileHeader) (*entity.AgreementDocument, error) {
	uid, e := uuid.Parse(userID)
	if e != nil {
		return nil, errors.New("user tidak valid")
	}
	def, ok := s.registry.Get(req.DocumentTypeCode)
	if !ok {
		return nil, errors.New("tipe dokumen tidak valid")
	}
	if e = validateAgreementForm(def, req.FormData); e != nil {
		return nil, e
	}
	raw, e := json.Marshal(req.FormData)
	if e != nil {
		return nil, e
	}
	seq, e := s.repo.NextAgreementSequence(time.Now().Year())
	if e != nil {
		return nil, e
	}
	doc := &entity.AgreementDocument{TicketNumber: utils.GenerateTicketNumber("PK", int(seq)), RequesterID: uid, DocumentTypeCode: def.Code, FormData: raw, AgreementNumber: fmt.Sprintf("%03d/RM.01.01/HR/IndonesiaRe/%d", seq, time.Now().Year()), Status: entity.StatusSubmitted}
	if e = s.repo.Create(doc); e != nil {
		return nil, errors.New("gagal membuat pengajuan")
	}
	if e = s.uploadAttachments(doc, uid, files, 1); e != nil {
		return nil, e
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
	return s.repo.FindAll(owner, q.Status, q.DateFrom, q.Search, q.Page, q.Limit)
}
func (s *agreementDocumentService) GetByID(id, userID string, all bool) (*entity.AgreementDocument, error) {
	uid, e := uuid.Parse(id)
	if e != nil {
		return nil, errors.New("ID tidak valid")
	}
	d, e := s.repo.FindByID(uid)
	if e != nil {
		return nil, errors.New("pengajuan tidak ditemukan")
	}
	if !all && d.RequesterID.String() != userID {
		return nil, errors.New("pengajuan tidak ditemukan")
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
	def, e := s.definitionFor(d.DocumentTypeCode)
	if e != nil {
		return nil, e
	}
	if e = validateAgreementForm(def, req.FormData); e != nil {
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
	round, _ := s.repo.LatestUploadRound(d.ID)
	if e = s.uploadAttachments(d, d.RequesterID, files, round+1); e != nil {
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
		d.AgreementNumber = strings.TrimSpace(*req.AgreementNumber)
	}
	setOptional(data, "tempat_ttd", req.SigningPlace)
	setOptional(data, "tanggal_ttd", req.SigningDate)
	setOptional(data, "pihak_pertama_pejabat", req.PartyOneSignatoryName)
	setOptional(data, "pihak_pertama_jabatan", req.PartyOneSignatoryPosition)
	d.FormData, _ = json.Marshal(data)
	if e = s.repo.Save(d); e != nil {
		return nil, e
	}
	return s.repo.FindByID(d.ID)
}
func setOptional(m map[string]interface{}, k string, v *string) {
	if v != nil {
		m[k] = strings.TrimSpace(*v)
	}
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
	def, e := s.definitionFor(d.DocumentTypeCode)
	if e != nil {
		return nil, e
	}
	data, e := decodeForm(d.FormData)
	if e != nil {
		return nil, e
	}
	if e = validateAgreementForm(def, data); e != nil {
		return nil, e
	}
	master, e := s.repo.GetActiveMaster()
	if e != nil {
		return nil, errors.New("Master Pihak Pertama belum tersedia")
	}
	snapshot := partyOneSnapshot(master, data)
	if valueString(data["tanggal_ttd"]) == "" {
		return nil, errors.New("tanggal tanda tangan wajib diisi approver")
	}
	values := placeholderValues(d, data, snapshot)
	docx, checksum, e := s.generator.Generate(def.Template, values, false)
	if e != nil {
		return nil, e
	}
	pdf, e := s.converter.ToPDF(ctx, docx)
	if e != nil {
		return nil, e
	}
	base := "agreement-documents/" + d.ID.String() + "/final/agreement"
	if e = s.storage.UploadBytes(ctx, base+".docx", docx, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"); e != nil {
		return nil, e
	}
	if e = s.storage.UploadBytes(ctx, base+".pdf", pdf, "application/pdf"); e != nil {
		s.storage.DeleteFile(ctx, base+".docx")
		return nil, e
	}
	snap, _ := json.Marshal(snapshot)
	now := time.Now()
	ok, e := s.repo.Complete(d.ID, entity.StatusUnderReview, map[string]interface{}{"status": entity.StatusCompleted, "status_updated_at": now, "party_one_snapshot": snap, "generated_docx_path": base + ".docx", "generated_pdf_path": base + ".pdf", "generated_file_name": "perjanjian-" + d.TicketNumber, "template_checksum": checksum, "approved_by": aid, "approved_at": now, "approver_note": ""})
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
	if d.Status == entity.StatusCompleted && d.GeneratedPDFPath != "" {
		obj, err := s.storage.GetFileObject(ctx, d.GeneratedPDFPath)
		if err != nil {
			return nil, err
		}
		defer obj.Close()
		return io.ReadAll(obj)
	}
	def, e := s.definitionFor(d.DocumentTypeCode)
	if e != nil {
		return nil, e
	}
	data, e := decodeForm(d.FormData)
	if e != nil {
		return nil, e
	}
	master, e := s.repo.GetActiveMaster()
	if e != nil {
		return nil, errors.New("Master Pihak Pertama belum tersedia")
	}
	docx, _, e := s.generator.Generate(def.Template, placeholderValues(d, data, partyOneSnapshot(master, data)), true)
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
	return map[string]interface{}{"name": m.Name, "address": m.Address, "phone": m.Phone, "email": m.Email, "pic": m.PIC, "signatory_name": name, "signatory_position": pos, "signing_place": place}
}
func placeholderValues(d *entity.AgreementDocument, data, snap map[string]interface{}) map[string]string {
	get := func(k string) string {
		v := strings.TrimSpace(valueString(data[k]))
		if v == "" {
			return "-"
		}
		return v
	}
	date := valueString(data["tanggal_ttd"])
	dt, _ := parseDate(date)
	atts := make([]string, 0, len(d.Attachments))
	for i, a := range d.Attachments {
		atts = append(atts, fmt.Sprintf("%d. %s", i+1, a.FileName))
	}
	if len(atts) == 0 {
		atts = append(atts, "-")
	}
	money := func(k string) int64 { return valueInt64(data[k]) }
	pct := func(k string) int64 { return valueInt64(data[k]) }
	return map[string]string{"NOMOR_PIHAK_PERTAMA": d.AgreementNumber, "NOMOR_PIHAK_KEDUA": get("nomor_pihak_kedua"), "HARI_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return indoDays[dt.Weekday()]
	}(), "TANGGAL_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return fmt.Sprint(dt.Day())
	}(), "BULAN_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return indoMonths[dt.Month()]
	}(), "TAHUN_TTD": func() string {
		if dt.IsZero() {
			return "-"
		}
		return fmt.Sprint(dt.Year())
	}(), "TANGGAL_TTD_LENGKAP": indoDate(date), "TEMPAT_TTD": valueString(snap["signing_place"]), "PIHAK_PERTAMA_NAMA": valueString(snap["name"]), "PIHAK_PERTAMA_ALAMAT": valueString(snap["address"]), "PIHAK_PERTAMA_TELEPON": valueString(snap["phone"]), "PIHAK_PERTAMA_EMAIL": valueString(snap["email"]), "PIHAK_PERTAMA_PIC": valueString(snap["pic"]), "PIHAK_PERTAMA_PEJABAT": valueString(snap["signatory_name"]), "PIHAK_PERTAMA_JABATAN": valueString(snap["signatory_position"]), "PIHAK_KEDUA_NAMA": get("pihak_kedua_nama"), "PIHAK_KEDUA_BIDANG": get("pihak_kedua_bidang"), "PIHAK_KEDUA_ALAMAT": get("pihak_kedua_alamat"), "PIHAK_KEDUA_TELEPON": get("pihak_kedua_telepon"), "PIHAK_KEDUA_EMAIL": get("pihak_kedua_email"), "PIHAK_KEDUA_PIC": get("pihak_kedua_pic"), "PIHAK_KEDUA_PEJABAT": get("pihak_kedua_pejabat"), "PIHAK_KEDUA_JABATAN": get("pihak_kedua_jabatan"), "JENIS_PEKERJAAN": get("jenis_pekerjaan"), "RUANG_LINGKUP": get("ruang_lingkup"), "SURAT_PENAWARAN_NOMOR": get("surat_penawaran_nomor"), "SURAT_PENAWARAN_PERIHAL": get("surat_penawaran_perihal"), "SURAT_PENAWARAN_TANGGAL": indoDate(valueString(data["surat_penawaran_tanggal"])), "SURAT_PENUNJUKAN_NOMOR": get("surat_penunjukan_nomor"), "SURAT_PENUNJUKAN_PERIHAL": get("surat_penunjukan_perihal"), "SURAT_PENUNJUKAN_TANGGAL": indoDate(valueString(data["surat_penunjukan_tanggal"])), "JANGKA_WAKTU_MULAI": indoDate(valueString(data["jangka_waktu_mulai"])), "JANGKA_WAKTU_SELESAI": indoDate(valueString(data["jangka_waktu_selesai"])), "NILAI_KONTRAK": rupiah(money("nilai_kontrak")), "NILAI_KONTRAK_TERBILANG": terbilang(money("nilai_kontrak")), "TERMIN_1_PERSEN": fmt.Sprint(pct("termin_1_persen")), "TERMIN_1_PERSEN_TERBILANG": terbilang(pct("termin_1_persen")), "TERMIN_1_NILAI": rupiah(money("termin_1_nilai")), "TERMIN_1_NILAI_TERBILANG": terbilang(money("termin_1_nilai")), "TERMIN_2_PERSEN": fmt.Sprint(pct("termin_2_persen")), "TERMIN_2_PERSEN_TERBILANG": terbilang(pct("termin_2_persen")), "TERMIN_2_NILAI": rupiah(money("termin_2_nilai")), "TERMIN_2_NILAI_TERBILANG": terbilang(money("termin_2_nilai")), "BANK": get("bank"), "NOMOR_REKENING": get("nomor_rekening"), "ATAS_NAMA": get("atas_nama"), "DAFTAR_LAMPIRAN": strings.Join(atts, "\n")}
}
