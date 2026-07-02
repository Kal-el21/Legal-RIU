package service

import (
	"errors"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type PurposeTypeService interface {
	GetAll() ([]entity.PurposeType, error)
	GetByID(id string) (*entity.PurposeType, error)
	GetByName(name string) (*entity.PurposeType, error)
	Create(name, description string) (*entity.PurposeType, error)
	Update(id, name, description string, isActive bool) (*entity.PurposeType, error)
	Delete(id string) error
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
