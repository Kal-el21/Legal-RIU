package service

import (
	"errors"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type CompanyService interface {
	GetAll() ([]entity.Company, error)
	GetByID(id string) (*entity.Company, error)
	GetByDomain(domain string) (*entity.Company, error)
	Create(name, emailDomain string, isInternal bool) (*entity.Company, error)
	Update(id, name, emailDomain string, isInternal bool) (*entity.Company, error)
	Delete(id string) error
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
