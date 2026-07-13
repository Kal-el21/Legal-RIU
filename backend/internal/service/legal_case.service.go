package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/storage"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/xuri/excelize/v2"
)

type LegalCaseService interface {
	Create(companyID *uuid.UUID, req dto.CreateLegalCaseRequest) (*dto.LegalCaseResponse, error)
	GetAll(companyID *uuid.UUID, query dto.LegalCaseListQuery) ([]dto.LegalCaseResponse, int64, error)
	GetLatest(companyID *uuid.UUID) (*dto.LegalCaseResponse, error)
	GetByID(companyID *uuid.UUID, id string) (*dto.LegalCaseResponse, error)
	Update(id string, req dto.UpdateLegalCaseRequest) (*dto.LegalCaseResponse, error)
	UpdateStatus(id string, status string) error
	Delete(id string) error

	ListChronologies(caseID string) ([]dto.CaseChronologyResponse, error)
	CreateChronology(caseID string, req dto.CreateCaseChronologyRequest, files []*multipart.FileHeader) (*dto.CaseChronologyResponse, error)
	UpdateChronology(caseID string, chronologyID string, req dto.UpdateCaseChronologyRequest, files []*multipart.FileHeader) (*dto.CaseChronologyResponse, error)
	DeleteChronology(caseID string, chronologyID string) error
	DownloadFile(filePath string) (*minio.Object, error)
	UploadDocument(caseID string, file *multipart.FileHeader) (*dto.LegalCaseResponse, error)
	DeleteDocument(caseID string) (*dto.LegalCaseResponse, error)
	UploadPhoto(caseID string, file *multipart.FileHeader) (*dto.LegalCaseResponse, error)
	DeletePhoto(caseID string) (*dto.LegalCaseResponse, error)
	ImportChronologies(caseID string, file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateChronologyTemplate() (*bytes.Buffer, error)
	ViewFile(filePath string) (*minio.Object, string, error)

	ListRegencies(search string, limit int) ([]dto.RegencyResponse, error)
	ListCedants(search string, limit int) ([]dto.CedantResponse, error)
	CreateCedant(req dto.CreateCedantRequest) (*dto.CedantResponse, error)
	UpdateCedant(id string, req dto.UpdateCedantRequest) (*dto.CedantResponse, error)
	DeleteCedant(id string) error
	ImportCedants(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateCedantTemplate() (*bytes.Buffer, error)
	ImportRegencies(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateRegencyTemplate() (*bytes.Buffer, error)

	GeneratePDF(id string) ([]byte, error)
}

type legalCaseService struct {
	repo    repository.LegalCaseRepository
	storage *storage.MinIOClient
	pdfSvc  PDFService
}

func NewLegalCaseService(repo repository.LegalCaseRepository, s *storage.MinIOClient) LegalCaseService {
	return &legalCaseService{repo: repo, storage: s, pdfSvc: NewPDFService(NewNoOpTemplateConversionService(), nil)}
}

func (s *legalCaseService) Create(companyID *uuid.UUID, req dto.CreateLegalCaseRequest) (*dto.LegalCaseResponse, error) {
	count, err := s.repo.CountByMonthAndPrefix("LC")
	if err != nil {
		return nil, errors.New("gagal generate ticket number")
	}
	ticket := utils.GenerateTicketNumber(utils.PrefixLegalCase, int(count)+1)

	legalCase, err := s.buildLegalCase(companyID, req)
	if err != nil {
		return nil, err
	}
	legalCase.TicketNumber = ticket

	if err := s.repo.Create(legalCase); err != nil {
		return nil, errors.New("gagal membuat kasus hukum")
	}

	created, err := s.repo.FindByID(legalCase.ID)
	if err != nil {
		return nil, errors.New("kasus berhasil dibuat, tetapi gagal mengambil detail")
	}
	response := toLegalCaseResponse(created, true)
	return &response, nil
}

func (s *legalCaseService) GetAll(companyID *uuid.UUID, query dto.LegalCaseListQuery) ([]dto.LegalCaseResponse, int64, error) {
	dateFrom, err := parseOptionalDate(query.DateFrom)
	if err != nil {
		return nil, 0, errors.New("tanggal awal tidak valid")
	}
	dateTo, err := parseOptionalDate(query.DateTo)
	if err != nil {
		return nil, 0, errors.New("tanggal akhir tidak valid")
	}

	filter := repository.LegalCaseFilter{
		Search:    query.Search,
		Status:    query.Status,
		CaseType:  query.CaseType,
		Level:     query.Level,
		CompanyID: companyID,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Page:      query.Page,
		Limit:     query.Limit,
	}

	items, total, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.LegalCaseResponse, 0, len(items))
	for i := range items {
		responses = append(responses, toLegalCaseResponse(&items[i], false))
	}

	return responses, total, nil
}

func (s *legalCaseService) GetLatest(companyID *uuid.UUID) (*dto.LegalCaseResponse, error) {
	filter := repository.LegalCaseFilter{CompanyID: companyID, Limit: 1}
	items, _, err := s.repo.FindAll(filter)
	if err != nil || len(items) == 0 {
		return nil, errors.New("kasus terbaru tidak ditemukan")
	}
	response := toLegalCaseResponse(&items[0], false)
	return &response, nil
}

func (s *legalCaseService) GetByID(companyID *uuid.UUID, id string) (*dto.LegalCaseResponse, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	legalCase, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus hukum tidak ditemukan")
	}

	response := toLegalCaseResponse(legalCase, true)
	return &response, nil
}

func (s *legalCaseService) Update(id string, req dto.UpdateLegalCaseRequest) (*dto.LegalCaseResponse, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	existing, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus hukum tidak ditemukan")
	}

	updated, err := s.buildLegalCase(existing.CompanyID, req)
	if err != nil {
		return nil, err
	}

	existing.CaseName = updated.CaseName
	existing.CaseSummary = updated.CaseSummary
	existing.RelatedPartyID = updated.RelatedPartyID
	existing.CategoryID = updated.CategoryID
	existing.Specification = updated.Specification
	existing.CaseTypeID = updated.CaseTypeID
	existing.TechnicalReserve = updated.TechnicalReserve
	existing.CaseValue = updated.CaseValue
	existing.PIC = updated.PIC
	existing.DocumentLink = updated.DocumentLink
	existing.Photo = updated.Photo
	existing.CurrentStatus = updated.CurrentStatus
	existing.CaseDate = updated.CaseDate
	existing.Level = updated.Level
	existing.AdditionalNotes = updated.AdditionalNotes
	existing.LocationRegencyID = updated.LocationRegencyID

	if err := s.repo.Update(existing); err != nil {
		return nil, errors.New("gagal mengupdate kasus hukum")
	}

	legalCase, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus berhasil diupdate, tetapi gagal mengambil detail")
	}
	response := toLegalCaseResponse(legalCase, true)
	return &response, nil
}

func (s *legalCaseService) UpdateStatus(id string, status string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}

	now := time.Now()
	return s.repo.UpdateStatus(uid, status, &now)
}

func (s *legalCaseService) Delete(id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if _, err := s.repo.FindByID(uid); err != nil {
		return errors.New("kasus hukum tidak ditemukan")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus kasus hukum")
	}
	return nil
}

func (s *legalCaseService) ListChronologies(caseID string) ([]dto.CaseChronologyResponse, error) {
	uid, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}
	if _, err := s.repo.FindByID(uid); err != nil {
		return nil, errors.New("kasus hukum tidak ditemukan")
	}

	items, err := s.repo.ListChronologies(uid)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CaseChronologyResponse, 0, len(items))
	for i := range items {
		responses = append(responses, toCaseChronologyResponse(&items[i]))
	}
	return responses, nil
}

func (s *legalCaseService) CreateChronology(caseID string, req dto.CreateCaseChronologyRequest, files []*multipart.FileHeader) (*dto.CaseChronologyResponse, error) {
	uid, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}
	if _, err := s.repo.FindByID(uid); err != nil {
		return nil, errors.New("kasus hukum tidak ditemukan")
	}

	agendaDate, err := parseRequiredDate(req.AgendaDate)
	if err != nil {
		return nil, errors.New("tanggal agenda tidak valid")
	}

	documents := append([]string{}, req.Documents...)
	uploaded, err := s.uploadChronologyDocuments(caseID, files)
	if err != nil {
		return nil, err
	}
	documents = append(documents, uploaded...)

	chronology := &entity.CaseChronology{
		CaseID:      uid,
		AgendaDate:  agendaDate,
		Agenda:      strings.TrimSpace(req.Agenda),
		Description: req.Description,
		Documents:   encodeDocuments(documents),
	}
	chronology.ID = uuid.New()

	if chronology.Agenda == "" {
		return nil, errors.New("agenda wajib diisi")
	}

	if err := s.repo.CreateChronology(chronology); err != nil {
		return nil, errors.New("gagal menambahkan kronologi kasus")
	}

	response := toCaseChronologyResponse(chronology)
	return &response, nil
}

func (s *legalCaseService) UpdateChronology(caseID string, chronologyID string, req dto.UpdateCaseChronologyRequest, files []*multipart.FileHeader) (*dto.CaseChronologyResponse, error) {
	caseUID, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}
	chronologyUID, err := parseUUID(chronologyID)
	if err != nil {
		return nil, errors.New("ID kronologi tidak valid")
	}

	chronology, err := s.repo.FindChronology(caseUID, chronologyUID)
	if err != nil {
		return nil, errors.New("kronologi kasus tidak ditemukan")
	}

	agendaDate, err := parseRequiredDate(req.AgendaDate)
	if err != nil {
		return nil, errors.New("tanggal agenda tidak valid")
	}

	documents := append([]string{}, req.Documents...)
	uploaded, err := s.uploadChronologyDocuments(caseID, files)
	if err != nil {
		return nil, err
	}
	documents = append(documents, uploaded...)

	chronology.AgendaDate = agendaDate
	chronology.Agenda = strings.TrimSpace(req.Agenda)
	chronology.Description = req.Description
	chronology.Documents = encodeDocuments(documents)

	if chronology.Agenda == "" {
		return nil, errors.New("agenda wajib diisi")
	}

	if err := s.repo.UpdateChronology(chronology); err != nil {
		return nil, errors.New("gagal mengupdate kronologi kasus")
	}

	response := toCaseChronologyResponse(chronology)
	return &response, nil
}

func (s *legalCaseService) DeleteChronology(caseID string, chronologyID string) error {
	caseUID, err := parseUUID(caseID)
	if err != nil {
		return errors.New("ID kasus tidak valid")
	}
	chronologyUID, err := parseUUID(chronologyID)
	if err != nil {
		return errors.New("ID kronologi tidak valid")
	}
	if _, err := s.repo.FindChronology(caseUID, chronologyUID); err != nil {
		return errors.New("kronologi kasus tidak ditemukan")
	}
	if err := s.repo.DeleteChronology(caseUID, chronologyUID); err != nil {
		return errors.New("gagal menghapus kronologi kasus")
	}
	return nil
}

func (s *legalCaseService) DownloadFile(filePath string) (*minio.Object, error) {
	if !strings.HasPrefix(filePath, "legal-cases/documents/") && !strings.HasPrefix(filePath, "legal-cases/chronologies/") {
		return nil, errors.New("path file tidak valid")
	}
	return s.storage.GetFileObject(context.Background(), filePath)
}

func (s *legalCaseService) ListRegencies(search string, limit int) ([]dto.RegencyResponse, error) {
	items, err := s.repo.ListRegencies(search, limit)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.RegencyResponse, 0, len(items))
	for i := range items {
		responses = append(responses, toRegencyResponse(&items[i]))
	}
	return responses, nil
}

func (s *legalCaseService) ListCedants(search string, limit int) ([]dto.CedantResponse, error) {
	items, err := s.repo.ListCedants(search, limit)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.CedantResponse, 0, len(items))
	for i := range items {
		responses = append(responses, toCedantResponse(&items[i]))
	}
	return responses, nil
}

func (s *legalCaseService) CreateCedant(req dto.CreateCedantRequest) (*dto.CedantResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("nama cedant wajib diisi")
	}

	cedant := &entity.Cedant{
		Name:        name,
		Description: req.Description,
	}
	if err := s.repo.CreateCedant(cedant); err != nil {
		return nil, errors.New("gagal membuat cedant")
	}
	response := toCedantResponse(cedant)
	return &response, nil
}

func (s *legalCaseService) UpdateCedant(id string, req dto.UpdateCedantRequest) (*dto.CedantResponse, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID cedant tidak valid")
	}
	cedant, err := s.repo.FindCedantByID(uid)
	if err != nil {
		return nil, errors.New("cedant tidak ditemukan")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("nama cedant wajib diisi")
	}

	cedant.Name = name
	cedant.Description = req.Description
	if err := s.repo.UpdateCedant(cedant); err != nil {
		return nil, errors.New("gagal mengupdate cedant")
	}

	response := toCedantResponse(cedant)
	return &response, nil
}

func (s *legalCaseService) DeleteCedant(id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return errors.New("ID cedant tidak valid")
	}
	if _, err := s.repo.FindCedantByID(uid); err != nil {
		return errors.New("cedant tidak ditemukan")
	}
	if err := s.repo.DeleteCedant(uid); err != nil {
		return errors.New("gagal menghapus cedant")
	}
	return nil
}

func (s *legalCaseService) ImportCedants(file *multipart.FileHeader) (*dto.ImportResult, error) {
	rows, err := utils.ReadSheet(file, 0)
	if err != nil {
		return nil, errors.New("gagal membaca file Excel: " + err.Error())
	}

	result := &dto.ImportResult{Errors: []dto.ImportRowError{}}
	if len(rows) < 2 {
		return result, nil
	}

	header := utils.NormalizeHeaders(rows[0])
	colName := utils.IndexOfHeader(header, "name")
	colDescription := utils.IndexOfHeader(header, "description")

	for i, row := range rows[1:] {
		rowNumber := i + 2
		if utils.IsEmptyRow(row) {
			continue
		}

		name := utils.CellValue(row, colName)
		description := utils.CellValue(row, colDescription)

		if name == "" {
			utils.AppendRowError(result, rowNumber, "name", "nama wajib diisi")
			continue
		}

		_, err := s.repo.FindCedantByName(name)
		if err == nil {
			utils.AppendRowError(result, rowNumber, "name", "nama sudah ada")
			continue
		}

		cedant := &entity.Cedant{Name: name, Description: description}
		if err := s.repo.CreateCedant(cedant); err != nil {
			utils.AppendRowError(result, rowNumber, "name", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *legalCaseService) GenerateCedantTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"name", "description"}
	examples := [][]string{
		{"Cedant A", "Deskripsi cedant A"},
		{"Cedant B", "Deskripsi cedant B"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}

func (s *legalCaseService) ImportRegencies(file *multipart.FileHeader) (*dto.ImportResult, error) {
	rows, err := utils.ReadSheet(file, 0)
	if err != nil {
		return nil, errors.New("gagal membaca file Excel: " + err.Error())
	}

	result := &dto.ImportResult{Errors: []dto.ImportRowError{}}
	if len(rows) < 2 {
		return result, nil
	}

	header := utils.NormalizeHeaders(rows[0])
	colName := utils.IndexOfHeader(header, "name")
	colProvince := utils.IndexOfHeader(header, "province")
	colType := utils.IndexOfHeader(header, "type")

	for i, row := range rows[1:] {
		rowNumber := i + 2
		if utils.IsEmptyRow(row) {
			continue
		}

		name := utils.CellValue(row, colName)
		province := utils.CellValue(row, colProvince)
		regencyType := utils.CellValue(row, colType)

		if name == "" {
			utils.AppendRowError(result, rowNumber, "name", "nama wajib diisi")
			continue
		}
		if province == "" {
			utils.AppendRowError(result, rowNumber, "province", "provinsi wajib diisi")
			continue
		}

		_, err := s.repo.FindRegencyByNameAndProvince(name, province)
		if err == nil {
			utils.AppendRowError(result, rowNumber, "name", "kabupaten/kota sudah ada di provinsi tersebut")
			continue
		}

		regency := &entity.Regency{Name: name, Province: province, Type: regencyType}
		if err := s.repo.CreateRegency(regency); err != nil {
			utils.AppendRowError(result, rowNumber, "name", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *legalCaseService) GenerateRegencyTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"name", "province", "type"}
	examples := [][]string{
		{"Jakarta Selatan", "DKI Jakarta", "Kota"},
		{"Bandung", "Jawa Barat", "Kabupaten"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}

func (s *legalCaseService) UploadDocument(caseID string, file *multipart.FileHeader) (*dto.LegalCaseResponse, error) {
	uid, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}

	legalCase, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus tidak ditemukan")
	}

	oldDocumentLink := legalCase.DocumentLink

	objectName, _, err := s.storage.UploadFile(context.Background(), "legal-cases/documents", file, fmt.Sprintf("case-%s", uid.String()))
	if err != nil {
		return nil, errors.New("gagal upload dokumen: " + err.Error())
	}

	legalCase.DocumentLink = objectName
	if err := s.repo.Update(legalCase); err != nil {
		_ = s.storage.DeleteFile(context.Background(), objectName)
		return nil, errors.New("gagal menyimpan dokumen")
	}

	if oldDocumentLink != "" {
		_ = s.storage.DeleteFile(context.Background(), oldDocumentLink)
	}

	response := toLegalCaseResponse(legalCase, true)
	return &response, nil
}

func (s *legalCaseService) DeleteDocument(caseID string) (*dto.LegalCaseResponse, error) {
	uid, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}

	legalCase, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus tidak ditemukan")
	}

	if legalCase.DocumentLink == "" {
		response := toLegalCaseResponse(legalCase, true)
		return &response, nil
	}

	oldDocumentLink := legalCase.DocumentLink
	legalCase.DocumentLink = ""
	if err := s.repo.Update(legalCase); err != nil {
		return nil, errors.New("gagal memperbarui kasus")
	}

	_ = s.storage.DeleteFile(context.Background(), oldDocumentLink)

	response := toLegalCaseResponse(legalCase, true)
	return &response, nil
}

func (s *legalCaseService) UploadPhoto(caseID string, file *multipart.FileHeader) (*dto.LegalCaseResponse, error) {
	uid, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}

	legalCase, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus tidak ditemukan")
	}

	oldPhoto := legalCase.Photo

	objectName, _, err := s.storage.UploadFile(context.Background(), "legal-cases/photos", file, fmt.Sprintf("case-%s", uid.String()))
	if err != nil {
		return nil, errors.New("gagal upload foto: " + err.Error())
	}

	legalCase.Photo = objectName
	if err := s.repo.Update(legalCase); err != nil {
		_ = s.storage.DeleteFile(context.Background(), objectName)
		return nil, errors.New("gagal menyimpan foto")
	}

	if oldPhoto != "" {
		_ = s.storage.DeleteFile(context.Background(), oldPhoto)
	}

	response := toLegalCaseResponse(legalCase, true)
	return &response, nil
}

func (s *legalCaseService) DeletePhoto(caseID string) (*dto.LegalCaseResponse, error) {
	uid, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}

	legalCase, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus tidak ditemukan")
	}

	if legalCase.Photo == "" {
		response := toLegalCaseResponse(legalCase, true)
		return &response, nil
	}

	oldPhoto := legalCase.Photo
	legalCase.Photo = ""
	if err := s.repo.Update(legalCase); err != nil {
		return nil, errors.New("gagal memperbarui kasus")
	}

	_ = s.storage.DeleteFile(context.Background(), oldPhoto)

	response := toLegalCaseResponse(legalCase, true)
	return &response, nil
}

func (s *legalCaseService) buildLegalCase(companyID *uuid.UUID, req dto.CreateLegalCaseRequest) (*entity.LegalCase, error) {
	caseDate, err := parseRequiredDate(req.CaseDate)
	if err != nil {
		return nil, errors.New("tanggal kasus tidak valid")
	}

	relatedPartyID, err := parseUUID(req.RelatedPartyID)
	if err != nil {
		return nil, errors.New("pihak terkait tidak valid")
	}
	if _, err := s.repo.FindCedantByID(relatedPartyID); err != nil {
		return nil, errors.New("pihak terkait tidak ditemukan")
	}

	locationRegencyID, err := parseUUID(req.LocationRegencyID)
	if err != nil {
		return nil, errors.New("lokasi kabupaten/kota tidak valid")
	}
	if _, err := s.repo.FindRegencyByID(locationRegencyID); err != nil {
		return nil, errors.New("lokasi kabupaten/kota tidak ditemukan")
	}

	categoryID, err := parseUUID(req.CategoryID)
	if err != nil {
		return nil, errors.New("kategori tidak valid")
	}
	if _, err := s.repo.FindCaseCategoryByID(categoryID); err != nil {
		return nil, errors.New("kategori tidak ditemukan")
	}

	caseTypeID, err := parseUUID(req.CaseTypeID)
	if err != nil {
		return nil, errors.New("jenis kasus tidak valid")
	}
	if _, err := s.repo.FindCaseTypeByID(caseTypeID); err != nil {
		return nil, errors.New("jenis kasus tidak ditemukan")
	}

	picID, err := parseUUID(req.PIC)
	if err != nil {
		return nil, errors.New("PIC tidak valid")
	}
	if _, err := s.repo.FindDivisionByID(picID); err != nil {
		return nil, errors.New("PIC tidak ditemukan")
	}

	legalCase := &entity.LegalCase{
		CaseName:          strings.TrimSpace(req.CaseName),
		CaseSummary:       req.CaseSummary,
		RelatedPartyID:    relatedPartyID,
		CategoryID:        &categoryID,
		Specification:     req.Specification,
		CaseTypeID:        &caseTypeID,
		TechnicalReserve:  req.TechnicalReserve,
		CaseValue:         req.CaseValue,
		PIC:               picID,
		DocumentLink:      req.DocumentLink,
		Photo:             req.Photo,
		CurrentStatus:     strings.TrimSpace(req.CurrentStatus),
		CaseDate:          caseDate,
		Level:             strings.TrimSpace(req.Level),
		AdditionalNotes:   req.AdditionalNotes,
		LocationRegencyID: locationRegencyID,
		CompanyID:         companyID,
	}

	if legalCase.CaseName == "" || legalCase.PIC == uuid.Nil || legalCase.Level == "" {
		return nil, errors.New("field wajib belum lengkap")
	}

	return legalCase, nil
}

func (s *legalCaseService) uploadChronologyDocuments(caseID string, files []*multipart.FileHeader) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}

	paths := make([]string, 0, len(files))
	for _, file := range files {
		objectPath, _, err := s.storage.UploadFile(context.Background(), "legal-cases/chronologies", file, fmt.Sprintf("case-%s-chron", caseID))
		if err != nil {
			for _, path := range paths {
				_ = s.storage.DeleteFile(context.Background(), path)
			}
			return nil, errors.New("gagal mengupload dokumen: " + file.Filename)
		}
		paths = append(paths, objectPath)
	}
	return paths, nil
}

func parseRequiredDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("tanggal wajib diisi")
	}
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse("02/01/2006", value); err == nil {
		return parsed, nil
	}
	if serial, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && serial > 0 {
		return time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, serial), nil
	}
	return time.Time{}, errors.New("format tanggal tidak valid")
}

func parseFlexibleDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("tanggal wajib diisi")
	}

	// Layouts tried in order. Day-month-year (Indonesian) is preferred and
	// tried before month-day-year. Any common separator is covered.
	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
		time.RFC3339,
		"02/01/2006", // DD/MM/YYYY
		"02-01-2006",
		"02.01.2006",
		"02 01 2006",
		"02-Jan-2006",
		"02 Jan 2006",
		"2 January 2006",
		"01/02/2006", // MM/DD/YYYY fallback
		"01-02-2006",
		"02/01/2006 15:04",
		"2006-01-02 15:04",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}

	// Excel serial date (numeric, e.g. 45123). Floor at 36526 (~ year 2000)
	// so a bare 4-digit year like "2024" is not mistaken for a serial.
	if serial, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64); err == nil && serial >= 36526 {
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		days := int(serial)
		frac := serial - float64(days)
		return base.AddDate(0, 0, days).Add(time.Duration(frac * 24 * float64(time.Hour))), nil
	}

	// Fallback: split by any non-digit separator, expect day-month-year.
	// Also supports 2-digit years (e.g. 15/02/24 -> 2024) and 8 digits (DDMMYYYY).
	parts := strings.FieldsFunc(value, func(r rune) bool { return r < '0' || r > '9' })
	if len(parts) == 3 {
		d, de := strconv.Atoi(parts[0])
		m, me := strconv.Atoi(parts[1])
		y, ye := strconv.Atoi(parts[2])
		if de == nil && me == nil && ye == nil {
			if y < 100 {
				y += 2000
			}
			if d >= 1 && d <= 31 && m >= 1 && m <= 12 && y > 1900 {
				return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC), nil
			}
		}
	} else if len(parts) == 1 && len(parts[0]) == 8 {
		s := parts[0]
		d, _ := strconv.Atoi(s[0:2])
		m, _ := strconv.Atoi(s[2:4])
		y, _ := strconv.Atoi(s[4:8])
		if d >= 1 && d <= 31 && m >= 1 && m <= 12 {
			return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC), nil
		}
	}

	return time.Time{}, errors.New("format tanggal tidak didukung")
}

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := parseRequiredDate(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func encodeDocuments(documents []string) string {
	if len(documents) == 0 {
		return "[]"
	}
	data, err := json.Marshal(documents)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeDocuments(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	var documents []string
	if err := json.Unmarshal([]byte(value), &documents); err != nil {
		return []string{}
	}
	return documents
}

func (s *legalCaseService) ImportChronologies(caseID string, file *multipart.FileHeader) (*dto.ImportResult, error) {
	uid, err := parseUUID(caseID)
	if err != nil {
		return nil, errors.New("ID kasus tidak valid")
	}
	if _, err := s.repo.FindByID(uid); err != nil {
		return nil, errors.New("kasus hukum tidak ditemukan")
	}

	src, err := file.Open()
	if err != nil {
		return nil, errors.New("gagal membaca file: " + err.Error())
	}
	defer src.Close()

	wb, err := excelize.OpenReader(src)
	if err != nil {
		return nil, errors.New("gagal membaca file Excel: " + err.Error())
	}
	defer wb.Close()

	sheetName := wb.GetSheetName(0)
	rows, err := wb.GetRows(sheetName)
	if err != nil {
		return nil, errors.New("gagal membaca sheet Excel: " + err.Error())
	}

	result := &dto.ImportResult{Errors: []dto.ImportRowError{}}
	if len(rows) < 2 {
		return result, nil
	}

	header := utils.NormalizeHeaders(rows[0])
	colAgendaDate := utils.IndexOfHeader(header, "agenda_date")
	colAgenda := utils.IndexOfHeader(header, "agenda")
	colDescription := utils.IndexOfHeader(header, "description")

	for i, row := range rows[1:] {
		rowNumber := i + 2
		if utils.IsEmptyRow(row) {
			continue
		}

		agendaDateRaw := utils.CellValue(row, colAgendaDate)
		agenda := utils.CellValue(row, colAgenda)
		description := utils.CellValue(row, colDescription)

		if agenda == "" {
			utils.AppendRowError(result, rowNumber, "agenda", "agenda wajib diisi")
			continue
		}

		agendaDate, err := parseFlexibleDate(agendaDateRaw)
		if err != nil {
			utils.AppendRowError(result, rowNumber, "agenda_date", "tanggal agenda tidak valid")
			continue
		}

		chronology := &entity.CaseChronology{
			CaseID:      uid,
			AgendaDate:  agendaDate,
			Agenda:      agenda,
			Description: description,
			Documents:   encodeDocuments(nil),
		}
		chronology.ID = uuid.New()

		if err := s.repo.CreateChronology(chronology); err != nil {
			utils.AppendRowError(result, rowNumber, "agenda", "gagal menyimpan kronologi")
			continue
		}

		result.Imported++
	}

	return result, nil
}

func (s *legalCaseService) GenerateChronologyTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	sheet := "Template"
	headers := []string{"agenda_date", "agenda", "description"}
	examples := [][]string{
		{"01/31/2024", "Sidang Pertama", "Menghadirkan saksi dan bukti dokumen"},
		{"02/15/2024", "Mediasi", "Mencoba penyelesaian damai antar pihak"},
		{"03/20/2024", "Putusan", "Menunggu salinan putusan dari kepaniteraan"},
	}
	utils.GenerateTemplate(wb, sheet, headers, examples)

	guide := "Petunjuk"
	wb.NewSheet(guide)
	guideLines := []string{
		"FORMAT TANGGAL: MM/DD/YYYY (bulan/hari/tahun)",
		"Contoh: 01/31/2024 berarti 31 Januari 2024.",
		"Juga mendukung: YYYY-MM-DD (mis. 2024-01-31) dan nomor serial Excel.",
		"Kolom agenda wajib diisi, kolom description opsional.",
		"Baris dengan agenda atau tanggal kosong/format salah akan dilewati.",
	}
	for i, line := range guideLines {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		wb.SetCellValue(guide, cell, line)
	}

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}

func (s *legalCaseService) ViewFile(filePath string) (*minio.Object, string, error) {
	if !strings.HasPrefix(filePath, "legal-cases/documents/") &&
		!strings.HasPrefix(filePath, "legal-cases/chronologies/") &&
		!strings.HasPrefix(filePath, "legal-cases/photos/") {
		return nil, "", errors.New("path file tidak valid")
	}

	obj, err := s.storage.GetFileObject(context.Background(), filePath)
	if err != nil {
		return nil, "", err
	}

	// Prefer MinIO's stored content type, fall back to extension detection.
	contentType := s.storage.GetFileContentType(context.Background(), filePath)
	if contentType == "" || contentType == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".webp":
			contentType = "image/webp"
		case ".pdf":
			contentType = "application/pdf"
		default:
			contentType = "application/octet-stream"
		}
	}

	return obj, contentType, nil
}

func toLegalCaseResponse(legalCase *entity.LegalCase, includeChronologies bool) dto.LegalCaseResponse {
	categoryID := ""
	if legalCase.CategoryID != nil {
		categoryID = legalCase.CategoryID.String()
	}
	caseTypeID := ""
	if legalCase.CaseTypeID != nil {
		caseTypeID = legalCase.CaseTypeID.String()
	}
	companyID := ""
	if legalCase.CompanyID != nil {
		companyID = legalCase.CompanyID.String()
	}

	response := dto.LegalCaseResponse{
		ID:                legalCase.ID.String(),
		CaseName:          legalCase.CaseName,
		CaseSummary:       legalCase.CaseSummary,
		TicketNumber:      legalCase.TicketNumber,
		RelatedPartyID:    legalCase.RelatedPartyID.String(),
		CategoryID:        categoryID,
		Specification:     legalCase.Specification,
		CaseTypeID:        caseTypeID,
		TechnicalReserve:  legalCase.TechnicalReserve,
		CaseValue:         legalCase.CaseValue,
		PIC:               legalCase.PIC.String(),
		DocumentLink:      legalCase.DocumentLink,
		Photo:             legalCase.Photo,
		CurrentStatus:     legalCase.CurrentStatus,
		CaseDate:          legalCase.CaseDate,
		Level:             legalCase.Level,
		AdditionalNotes:   legalCase.AdditionalNotes,
		LocationRegencyID: legalCase.LocationRegencyID.String(),
		CompanyID:         companyID,
		CreatedAt:         legalCase.CreatedAt,
		UpdatedAt:         legalCase.UpdatedAt,
		PICDivision:       toDivisionResponsePointer(&legalCase.PICDivision),
	}

	if legalCase.RelatedParty.ID != uuid.Nil {
		relatedParty := toCedantResponse(&legalCase.RelatedParty)
		response.RelatedParty = &relatedParty
	}
	if legalCase.LocationRegency.ID != uuid.Nil {
		location := toRegencyResponse(&legalCase.LocationRegency)
		response.LocationRegency = &location
	}
	if legalCase.CaseTypeRef.ID != uuid.Nil {
		caseType := toCaseTypeResponse(&legalCase.CaseTypeRef)
		response.CaseType = &caseType
	}
	if legalCase.CategoryRef.ID != uuid.Nil {
		category := toCaseCategoryResponse(&legalCase.CategoryRef)
		response.Category = &category
	}
	if legalCase.Company.ID != uuid.Nil {
		company := toCompanyResponse(&legalCase.Company)
		response.Company = &company
	}
	if includeChronologies {
		response.Chronologies = make([]dto.CaseChronologyResponse, 0, len(legalCase.Chronologies))
		for i := range legalCase.Chronologies {
			response.Chronologies = append(response.Chronologies, toCaseChronologyResponse(&legalCase.Chronologies[i]))
		}
	}

	return response
}

func toCaseChronologyResponse(chronology *entity.CaseChronology) dto.CaseChronologyResponse {
	return dto.CaseChronologyResponse{
		ID:          chronology.ID.String(),
		CaseID:      chronology.CaseID.String(),
		AgendaDate:  chronology.AgendaDate,
		Agenda:      chronology.Agenda,
		Description: chronology.Description,
		Documents:   decodeDocuments(chronology.Documents),
		CreatedAt:   chronology.CreatedAt,
		UpdatedAt:   chronology.UpdatedAt,
	}
}

func toRegencyResponse(regency *entity.Regency) dto.RegencyResponse {
	return dto.RegencyResponse{
		ID:       regency.ID.String(),
		Name:     regency.Name,
		Province: regency.Province,
		Type:     regency.Type,
		Label:    regency.Name + " - " + regency.Province,
	}
}

func toCedantResponse(cedant *entity.Cedant) dto.CedantResponse {
	return dto.CedantResponse{
		ID:          cedant.ID.String(),
		Name:        cedant.Name,
		Description: cedant.Description,
		CreatedAt:   cedant.CreatedAt,
		UpdatedAt:   cedant.UpdatedAt,
	}
}

func FileNameFromPath(filePath string) string {
	return filepath.Base(filePath)
}

func (s *legalCaseService) GeneratePDF(id string) ([]byte, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	lc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kasus hukum tidak ditemukan")
	}

	return s.pdfSvc.GenerateLegalCasePDF(lc)
}
