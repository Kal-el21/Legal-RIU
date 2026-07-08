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

type PurposeTypeService interface {
	GetAll() ([]entity.PurposeType, error)
	GetByID(id string) (*entity.PurposeType, error)
	GetByName(name string) (*entity.PurposeType, error)
	Create(name, description string) (*entity.PurposeType, error)
	Update(id, name, description string, isActive bool) (*entity.PurposeType, error)
	Delete(id string) error
	ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateImportTemplate() (*bytes.Buffer, error)
}

type purposeTypeService struct {
	repo repository.PurposeTypeRepository
}

func NewPurposeTypeService(repo repository.PurposeTypeRepository) PurposeTypeService {
	return &purposeTypeService{repo: repo}
}

func (s *purposeTypeService) GetAll() ([]entity.PurposeType, error) {
	return s.repo.FindAll()
}

func (s *purposeTypeService) GetByID(id string) (*entity.PurposeType, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	return s.repo.FindByID(uid)
}

func (s *purposeTypeService) GetByName(name string) (*entity.PurposeType, error) {
	return s.repo.FindByName(name)
}

func (s *purposeTypeService) Create(name, description string) (*entity.PurposeType, error) {
	if name == "" {
		return nil, errors.New("nama tujuan pembuatan wajib diisi")
	}
	pt := &entity.PurposeType{
		Name:        name,
		Description: description,
		IsActive:    true,
	}
	if err := s.repo.Create(pt); err != nil {
		return nil, errors.New("gagal membuat tujuan pembuatan")
	}
	return pt, nil
}

func (s *purposeTypeService) Update(id, name, description string, isActive bool) (*entity.PurposeType, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	pt, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("tujuan pembuatan tidak ditemukan")
	}
	pt.Name = name
	pt.Description = description
	pt.IsActive = isActive
	if err := s.repo.Update(pt); err != nil {
		return nil, errors.New("gagal mengupdate tujuan pembuatan")
	}
	return pt, nil
}

func (s *purposeTypeService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus tujuan pembuatan")
	}
	return nil
}

func (s *purposeTypeService) ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error) {
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

		_, err := s.repo.FindByName(name)
		if err == nil {
			utils.AppendRowError(result, rowNumber, "name", "nama sudah ada")
			continue
		}

		pt := &entity.PurposeType{Name: name, Description: description, IsActive: true}
		if err := s.repo.Create(pt); err != nil {
			utils.AppendRowError(result, rowNumber, "name", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *purposeTypeService) GenerateImportTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"name", "description"}
	examples := [][]string{
		{"Asuransi Jiwa", "Untuk produk asuransi jiwa"},
		{"Asuransi Kesehatan", "Untuk produk asuransi kesehatan"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}
