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

type DocumentTypeService interface {
	GetAll() ([]entity.DocumentType, error)
	GetByID(id string) (*entity.DocumentType, error)
	GetByName(name string) (*entity.DocumentType, error)
	Create(name, label string) (*entity.DocumentType, error)
	Update(id, name, label string, isActive bool) (*entity.DocumentType, error)
	Delete(id string) error
	ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateImportTemplate() (*bytes.Buffer, error)
}

type documentTypeService struct {
	repo repository.DocumentTypeRepository
}

func NewDocumentTypeService(repo repository.DocumentTypeRepository) DocumentTypeService {
	return &documentTypeService{repo: repo}
}

func (s *documentTypeService) GetAll() ([]entity.DocumentType, error) {
	return s.repo.FindAll()
}

func (s *documentTypeService) GetByID(id string) (*entity.DocumentType, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	return s.repo.FindByID(uid)
}

func (s *documentTypeService) GetByName(name string) (*entity.DocumentType, error) {
	return s.repo.FindByName(name)
}

func (s *documentTypeService) Create(name, label string) (*entity.DocumentType, error) {
	if name == "" {
		return nil, errors.New("nama wajib diisi")
	}
	if label == "" {
		return nil, errors.New("label wajib diisi")
	}
	dt := &entity.DocumentType{
		Name:     name,
		Label:    label,
		IsActive: true,
	}
	if err := s.repo.Create(dt); err != nil {
		return nil, errors.New("gagal membuat jenis dokumen")
	}
	return dt, nil
}

func (s *documentTypeService) Update(id, name, label string, isActive bool) (*entity.DocumentType, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	dt, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("jenis dokumen tidak ditemukan")
	}
	dt.Name = name
	dt.Label = label
	dt.IsActive = isActive
	if err := s.repo.Update(dt); err != nil {
		return nil, errors.New("gagal mengupdate jenis dokumen")
	}
	return dt, nil
}

func (s *documentTypeService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus jenis dokumen")
	}
	return nil
}

func (s *documentTypeService) ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error) {
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
	colLabel := utils.IndexOfHeader(header, "label")

	for i, row := range rows[1:] {
		rowNumber := i + 2
		if utils.IsEmptyRow(row) {
			continue
		}

		name := utils.CellValue(row, colName)
		label := utils.CellValue(row, colLabel)

		if name == "" {
			utils.AppendRowError(result, rowNumber, "name", "nama wajib diisi")
			continue
		}
		if label == "" {
			utils.AppendRowError(result, rowNumber, "label", "label wajib diisi")
			continue
		}

		_, err := s.repo.FindByName(name)
		if err == nil {
			utils.AppendRowError(result, rowNumber, "name", "nama sudah ada")
			continue
		}

		dt := &entity.DocumentType{Name: name, Label: label, IsActive: true}
		if err := s.repo.Create(dt); err != nil {
			utils.AppendRowError(result, rowNumber, "name", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *documentTypeService) GenerateImportTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"name", "label"}
	examples := [][]string{
		{"surat_perintah_kerja", "Surat Perintah Kerja"},
		{"kontrak_treaty", "Kontrak Treaty"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}