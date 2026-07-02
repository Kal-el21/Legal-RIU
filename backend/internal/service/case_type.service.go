package service

import (
	"errors"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type CaseTypeService interface {
	GetAll() ([]entity.CaseType, error)
	GetByID(id string) (*entity.CaseType, error)
	GetByCode(code string) (*entity.CaseType, error)
	Create(code, label string) (*entity.CaseType, error)
	Update(id, code, label string, isActive bool) (*entity.CaseType, error)
	Delete(id string) error
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
