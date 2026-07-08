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

type DivisionService interface {
	GetAll(search string, limit int) ([]entity.Division, error)
	GetByID(id string) (*entity.Division, error)
	Create(name, description string) (*entity.Division, error)
	Update(id, name, description string) (*entity.Division, error)
	Delete(id string) error
	SyncFromList(divisions []entity.Division) error
	ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateImportTemplate() (*bytes.Buffer, error)
}

type divisionService struct {
	repo repository.DivisionRepository
}

func NewDivisionService(repo repository.DivisionRepository) DivisionService {
	return &divisionService{repo: repo}
}

func (s *divisionService) GetAll(search string, limit int) ([]entity.Division, error) {
	return s.repo.FindAll(search, limit)
}

func (s *divisionService) GetByID(id string) (*entity.Division, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	return s.repo.FindByID(uid)
}

func (s *divisionService) Create(name, description string) (*entity.Division, error) {
	if name == "" {
		return nil, errors.New("nama divisi wajib diisi")
	}
	division := &entity.Division{
		Name:        name,
		Description: description,
	}
	if err := s.repo.Create(division); err != nil {
		return nil, errors.New("gagal membuat divisi")
	}
	return division, nil
}

func (s *divisionService) Update(id, name, description string) (*entity.Division, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	if name == "" {
		return nil, errors.New("nama divisi wajib diisi")
	}
	division, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("divisi tidak ditemukan")
	}
	division.Name = name
	division.Description = description
	if err := s.repo.Update(division); err != nil {
		return nil, errors.New("gagal mengupdate divisi")
	}
	return division, nil
}

func (s *divisionService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus divisi")
	}
	return nil
}

func (s *divisionService) SyncFromList(divisions []entity.Division) error {
	for _, d := range divisions {
		_, err := s.repo.FindByName(d.Name)
		if err != nil {
			if err := s.repo.Create(&d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *divisionService) ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error) {
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

		division := &entity.Division{Name: name, Description: description}
		if err := s.repo.Create(division); err != nil {
			utils.AppendRowError(result, rowNumber, "name", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *divisionService) GenerateImportTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"name", "description"}
	examples := [][]string{
		{"Divisi Hukum", "Divisi yang menangani hukum"},
		{"Divisi Teknik", "Divisi yang menangani teknik"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}
