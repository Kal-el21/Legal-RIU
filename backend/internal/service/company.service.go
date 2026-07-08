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

type CompanyService interface {
	GetAll() ([]entity.Company, error)
	GetByID(id string) (*entity.Company, error)
	GetByDomain(domain string) (*entity.Company, error)
	Create(name, emailDomain string, isInternal bool) (*entity.Company, error)
	Update(id, name, emailDomain string, isInternal bool) (*entity.Company, error)
	Delete(id string) error
	ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateImportTemplate() (*bytes.Buffer, error)
}

type companyService struct {
	repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) CompanyService {
	return &companyService{repo: repo}
}

func (s *companyService) GetAll() ([]entity.Company, error) {
	return s.repo.FindAll()
}

func (s *companyService) GetByID(id string) (*entity.Company, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	return s.repo.FindByID(uid)
}

func (s *companyService) GetByDomain(domain string) (*entity.Company, error) {
	return s.repo.FindByDomain(domain)
}

func (s *companyService) Create(name, emailDomain string, isInternal bool) (*entity.Company, error) {
	if name == "" || emailDomain == "" {
		return nil, errors.New("nama dan domain email wajib diisi")
	}
	company := &entity.Company{
		Name:        name,
		EmailDomain: emailDomain,
		IsInternal:  isInternal,
	}
	if err := s.repo.Create(company); err != nil {
		return nil, errors.New("gagal membuat perusahaan")
	}
	return company, nil
}

func (s *companyService) Update(id, name, emailDomain string, isInternal bool) (*entity.Company, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	company, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("perusahaan tidak ditemukan")
	}
	company.Name = name
	company.EmailDomain = emailDomain
	company.IsInternal = isInternal
	if err := s.repo.Update(company); err != nil {
		return nil, errors.New("gagal mengupdate perusahaan")
	}
	return company, nil
}

func (s *companyService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus perusahaan")
	}
	return nil
}

func (s *companyService) ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error) {
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
	colDomain := utils.IndexOfHeader(header, "email_domain")

	for i, row := range rows[1:] {
		rowNumber := i + 2
		if utils.IsEmptyRow(row) {
			continue
		}

		name := utils.CellValue(row, colName)
		emailDomain := utils.CellValue(row, colDomain)

		if name == "" {
			utils.AppendRowError(result, rowNumber, "name", "nama wajib diisi")
			continue
		}
		if emailDomain == "" {
			utils.AppendRowError(result, rowNumber, "email_domain", "domain email wajib diisi")
			continue
		}

		_, err := s.repo.FindByDomain(emailDomain)
		if err == nil {
			utils.AppendRowError(result, rowNumber, "email_domain", "domain email sudah ada")
			continue
		}

		company := &entity.Company{Name: name, EmailDomain: emailDomain, IsInternal: false}
		if err := s.repo.Create(company); err != nil {
			utils.AppendRowError(result, rowNumber, "name", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *companyService) GenerateImportTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"name", "email_domain"}
	examples := [][]string{
		{"PT ABC", "abc.co.id"},
		{"PT XYZ", "xyz.co.id"},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
}
