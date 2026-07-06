package service

import (
	"errors"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type DocumentTypeService interface {
	GetAll() ([]entity.DocumentType, error)
	GetByID(id string) (*entity.DocumentType, error)
	GetByName(name string) (*entity.DocumentType, error)
	Create(name, label string) (*entity.DocumentType, error)
	Update(id, name, label string, isActive bool) (*entity.DocumentType, error)
	Delete(id string) error
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