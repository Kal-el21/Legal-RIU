package service

import (
	"errors"

	"legal-riu-portal/internal/dto"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/repository"

	"github.com/google/uuid"
)

type LegalMaterialService interface {
	GetAll() ([]entity.LegalMaterial, error)
	GetByID(id string) (*entity.LegalMaterial, error)
	Create(title, excerpt, content string, createdBy uuid.UUID) (*entity.LegalMaterial, error)
	Update(id, title, excerpt, content string, updatedBy uuid.UUID) (*entity.LegalMaterial, error)
	Delete(id string) error
}

type legalMaterialService struct {
	repo repository.LegalMaterialRepository
}

func NewLegalMaterialService(repo repository.LegalMaterialRepository) LegalMaterialService {
	return &legalMaterialService{repo: repo}
}

func (s *legalMaterialService) GetAll() ([]entity.LegalMaterial, error) {
	return s.repo.FindAll()
}

func (s *legalMaterialService) GetByID(id string) (*entity.LegalMaterial, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	return s.repo.FindByID(uid)
}

func (s *legalMaterialService) Create(title, excerpt, content string, createdBy uuid.UUID) (*entity.LegalMaterial, error) {
	if title == "" || content == "" {
		return nil, errors.New("judul dan konten wajib diisi")
	}
	material := &entity.LegalMaterial{
		Title:     title,
		Excerpt:   excerpt,
		Content:   content,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}
	if err := s.repo.Create(material); err != nil {
		return nil, errors.New("gagal membuat materi")
	}
	return material, nil
}

func (s *legalMaterialService) Update(id, title, excerpt, content string, updatedBy uuid.UUID) (*entity.LegalMaterial, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}
	material, err := s.repo.FindByID(uid)
	if err != nil {
		return nil, errors.New("materi tidak ditemukan")
	}
	material.Title = title
	material.Excerpt = excerpt
	material.Content = content
	material.UpdatedBy = updatedBy
	if err := s.repo.Update(material); err != nil {
		return nil, errors.New("gagal mengupdate materi")
	}
	return material, nil
}

func (s *legalMaterialService) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	if _, err := s.repo.FindByID(uid); err != nil {
		return errors.New("materi tidak ditemukan")
	}
	if err := s.repo.Delete(uid); err != nil {
		return errors.New("gagal menghapus materi")
	}
	return nil
}

func toLegalMaterialResponse(material *entity.LegalMaterial) dto.LegalMaterialResponse {
	return dto.LegalMaterialResponse{
		ID:        material.ID.String(),
		Title:     material.Title,
		Excerpt:   material.Excerpt,
		Content:   material.Content,
		CreatedBy: material.CreatedBy.String(),
		UpdatedBy: material.UpdatedBy.String(),
		CreatedAt: material.CreatedAt,
		UpdatedAt: material.UpdatedAt,
	}
}

func toLegalMaterialResponseList(materials []entity.LegalMaterial) []dto.LegalMaterialResponse {
	resp := make([]dto.LegalMaterialResponse, 0, len(materials))
	for i := range materials {
		resp = append(resp, toLegalMaterialResponse(&materials[i]))
	}
	return resp
}
