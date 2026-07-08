package service

import (
	"bytes"
	"errors"
	"mime/multipart"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"
	"legal-riu-portal/internal/utils"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type CaseTypeService interface {
	GetAll() ([]entity.CaseType, error)
	GetByID(id string) (*entity.CaseType, error)
	GetByCode(code string) (*entity.CaseType, error)
	Create(code, label string) (*entity.CaseType, error)
	Update(id, code, label string, isActive bool) (*entity.CaseType, error)
	Delete(id string) error
	ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateImportTemplate() (*bytes.Buffer, error)
}

type caseTypeService struct {
	repo repository.CaseTypeRepository
}

func NewCaseTypeService(repo repository.CaseTypeRepository) CaseTypeService {
	return &caseTypeService{repo: repo}
}

func (s *caseTypeService) GetAll() ([]entity.CaseType, error) {
	return s.repo.FindAll()
}

func (s *caseTypeService) GetByID(id string) (*entity.CaseType, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	return s.repo.FindByID(uid)
}

func (s *caseTypeService) GetByCode(code string) (*entity.CaseType, error) {
	return s.repo.FindByCode(code)
}

func (s *caseTypeService) Create(code, label string) (*entity.CaseType, error) {
	if code == "" || label == "" {
		return nil, errors.New("kode dan label jenis kasus wajib diisi")
	}
	ct := &entity.CaseType{
		Code:     code,
		Label:    label,
		IsActive: true,
	}
	if err := s.repo.Create(ct); err != nil {
		return nil, errors.New("gagal membuat jenis kasus")
	}
	return ct, nil
}

func (s *caseTypeService) Update(id, code, label string, isActive bool) (*entity.CaseType, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	ct, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("jenis kasus tidak ditemukan")
	}
	ct.Code = code
	ct.Label = label
	ct.IsActive = isActive
	if err := s.repo.Update(ct); err != nil {
		return nil, errors.New("gagal mengupdate jenis kasus")
	}
	return ct, nil
}

func (s *caseTypeService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus jenis kasus")
	}
	return nil
}

func (s *caseTypeService) ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error) {
	rows, err := utils.ReadSheet(file, 0)
	if err != nil {
		return nil, errors.New("gagal membaca file Excel: " + err.Error())
	}

	result := &dto.ImportResult{Errors: []dto.ImportRowError{}}
	if len(rows) < 2 {
		return result, nil
	}

	header := utils.NormalizeHeaders(rows[0])
	colCode := utils.IndexOfHeader(header, "code")
	colLabel := utils.IndexOfHeader(header, "label")

	for i, row := range rows[1:] {
		rowNumber := i + 2
		if utils.IsEmptyRow(row) {
			continue
		}

		code := utils.CellValue(row, colCode)
		label := utils.CellValue(row, colLabel)

		if code == "" {
			utils.AppendRowError(result, rowNumber, "code", "kode wajib diisi")
			continue
		}
		if label == "" {
			utils.AppendRowError(result, rowNumber, "label", "label wajib diisi")
			continue
		}

		_, err := s.repo.FindByCode(code)
		if err == nil {
			utils.AppendRowError(result, rowNumber, "code", "kode sudah ada")
			continue
		}

		ct := &entity.CaseType{Code: code, Label: label, IsActive: true}
		if err := s.repo.Create(ct); err != nil {
			utils.AppendRowError(result, rowNumber, "code", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *caseTypeService) GenerateImportTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"code", "label"}
	examples := [][]string{
		{"CT-001", "Jenis Kasus Umum"},
		{"CT-002", "Jenis Kasus Khusus"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}
