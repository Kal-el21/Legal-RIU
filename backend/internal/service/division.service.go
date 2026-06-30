package service

import (
	"errors"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type DivisionService interface {
	GetAll(search string, limit int) ([]entity.Division, error)
	GetByID(id string) (*entity.Division, error)
	Create(name, description string) (*entity.Division, error)
	Update(id, name, description string) (*entity.Division, error)
	Delete(id string) error
	SyncFromList(divisions []entity.Division) error
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
