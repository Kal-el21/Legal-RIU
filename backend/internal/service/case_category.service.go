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

type CaseCategoryService interface {
	GetAll() ([]entity.CaseCategory, error)
	GetByID(id string) (*entity.CaseCategory, error)
	GetByCode(code string) (*entity.CaseCategory, error)
	Create(code, label string) (*entity.CaseCategory, error)
	Update(id, code, label string, isActive bool) (*entity.CaseCategory, error)
	Delete(id string) error
	ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateImportTemplate() (*bytes.Buffer, error)
}

type caseCategoryService struct {
	repo repository.CaseCategoryRepository
}

func NewCaseCategoryService(repo repository.CaseCategoryRepository) CaseCategoryService {
	return &caseCategoryService{repo: repo}
}

func (s *caseCategoryService) GetAll() ([]entity.CaseCategory, error) {
	return s.repo.FindAll()
}

func (s *caseCategoryService) GetByID(id string) (*entity.CaseCategory, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	return s.repo.FindByID(uid)
}

func (s *caseCategoryService) GetByCode(code string) (*entity.CaseCategory, error) {
	return s.repo.FindByCode(code)
}

func (s *caseCategoryService) Create(code, label string) (*entity.CaseCategory, error) {
	if code == "" || label == "" {
		return nil, errors.New("kode dan label kategori wajib diisi")
	}
	cc := &entity.CaseCategory{
		Code:     code,
		Label:    label,
		IsActive: true,
	}
	if err := s.repo.Create(cc); err != nil {
		return nil, errors.New("gagal membuat kategori")
	}
	return cc, nil
}

func (s *caseCategoryService) Update(id, code, label string, isActive bool) (*entity.CaseCategory, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	cc, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("kategori tidak ditemukan")
	}
	cc.Code = code
	cc.Label = label
	cc.IsActive = isActive
	if err := s.repo.Update(cc); err != nil {
		return nil, errors.New("gagal mengupdate kategori")
	}
	return cc, nil
}

func (s *caseCategoryService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus kategori")
	}
	return nil
}

func (s *caseCategoryService) ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error) {
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

		cc := &entity.CaseCategory{Code: code, Label: label, IsActive: true}
		if err := s.repo.Create(cc); err != nil {
			utils.AppendRowError(result, rowNumber, "code", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *caseCategoryService) GenerateImportTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"code", "label"}
	examples := [][]string{
		{"CC-001", "Kategori Umum"},
		{"CC-002", "Kategori Khusus"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}
