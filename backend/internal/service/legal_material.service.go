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

type LegalMaterialService interface {
	GetAll() ([]entity.LegalMaterial, error)
	GetByID(id string) (*entity.LegalMaterial, error)
	Create(title, excerpt, content string, createdBy uuid.UUID) (*entity.LegalMaterial, error)
	Update(id, title, excerpt, content string, updatedBy uuid.UUID) (*entity.LegalMaterial, error)
	Delete(id string) error
	ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error)
	GenerateImportTemplate() (*bytes.Buffer, error)
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

func (s *legalMaterialService) ImportFromExcel(file *multipart.FileHeader) (*dto.ImportResult, error) {
	rows, err := utils.ReadSheet(file, 0)
	if err != nil {
		return nil, errors.New("gagal membaca file Excel: " + err.Error())
	}

	result := &dto.ImportResult{Errors: []dto.ImportRowError{}}
	if len(rows) < 2 {
		return result, nil
	}

	header := utils.NormalizeHeaders(rows[0])
	colTitle := utils.IndexOfHeader(header, "title")
	colExcerpt := utils.IndexOfHeader(header, "excerpt")
	colContent := utils.IndexOfHeader(header, "content")

	createdBy, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")

	for i, row := range rows[1:] {
		rowNumber := i + 2
		if utils.IsEmptyRow(row) {
			continue
		}

		title := utils.CellValue(row, colTitle)
		excerpt := utils.CellValue(row, colExcerpt)
		content := utils.CellValue(row, colContent)

		if title == "" {
			utils.AppendRowError(result, rowNumber, "title", "judul wajib diisi")
			continue
		}
		if content == "" {
			utils.AppendRowError(result, rowNumber, "content", "konten wajib diisi")
			continue
		}

		material := &entity.LegalMaterial{
			Title:     title,
			Excerpt:   excerpt,
			Content:   content,
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		}
		if err := s.repo.Create(material); err != nil {
			utils.AppendRowError(result, rowNumber, "title", "gagal menyimpan")
			continue
		}
		result.Imported++
	}

	return result, nil
}

func (s *legalMaterialService) GenerateImportTemplate() (*bytes.Buffer, error) {
	wb := excelize.NewFile()
	defer wb.Close()

	headers := []string{"title", "excerpt", "content"}
	examples := [][]string{
		{"Contoh Materi", "Ringkasan materi", "Isi lengkap materi..."},
	}
	utils.GenerateTemplate(wb, "Template", headers, examples)

	var buf bytes.Buffer
	if err := wb.Write(&buf); err != nil {
		return nil, errors.New("gagal membuat template: " + err.Error())
	}
	return &buf, nil
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
