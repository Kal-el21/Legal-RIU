package service

import (
	"errors"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type CaseCategoryService interface {
	GetAll() ([]entity.CaseCategory, error)
	GetByID(id string) (*entity.CaseCategory, error)
	GetByCode(code string) (*entity.CaseCategory, error)
	Create(code, label string) (*entity.CaseCategory, error)
	Update(id, code, label string, isActive bool) (*entity.CaseCategory, error)
	Delete(id string) error
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
