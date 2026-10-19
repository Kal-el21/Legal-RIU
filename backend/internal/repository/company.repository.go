package repository

import (
	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	FindAll() ([]entity.Company, error)
	FindByID(id uuid.UUID) (*entity.Company, error)
	FindByDomain(domain string) (*entity.Company, error)
	Create(company *entity.Company) error
	Update(company *entity.Company) error
	Delete(id uuid.UUID) error
	FindParent() (*entity.Company, error)
	ParentChangeImpact(targetID uuid.UUID) (legalToMove int64, legalAU []entity.User, err error)
	MakeParent(targetID uuid.UUID) ([]uuid.UUID, error)
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) FindAll() ([]entity.Company, error) {
	var items []entity.Company
	err := r.db.Order("name ASC").Find(&items).Error
	return items, err
}

func (r *companyRepository) FindByID(id uuid.UUID) (*entity.Company, error) {
	var company entity.Company
	err := r.db.Where("id = ?", id).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) FindByDomain(domain string) (*entity.Company, error) {
	var company entity.Company
	err := r.db.Where("LOWER(email_domain) = LOWER(?)", domain).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Create(company *entity.Company) error {
	return r.db.Create(company).Error
}

func (r *companyRepository) Update(company *entity.Company) error {
	return r.db.Save(company).Error
}

func (r *companyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Company{}, "id = ?", id).Error
}

func (r *companyRepository) FindParent() (*entity.Company, error) {
	var company entity.Company
	if err := r.db.Where("is_parent = ?", true).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

// ParentChangeImpact menghitung LEGAL yang akan pindah ke targetID dan
// mengambil Legal AU yang terdaftar di sana.
func (r *companyRepository) ParentChangeImpact(targetID uuid.UUID) (int64, []entity.User, error) {
	var legalToMove int64
	if err := r.db.Model(&entity.User{}).
		Where("role = ? AND company_id IS DISTINCT FROM ?", entity.RoleLegal, targetID).
		Count(&legalToMove).Error; err != nil {
		return 0, nil, err
	}
	var legalAU []entity.User
	err := r.db.Where("role = ? AND company_id = ?", entity.RoleLegalAU, targetID).Order("full_name ASC").Find(&legalAU).Error
	return legalToMove, legalAU, err
}

// MakeParent memindah tanda induk ke targetID dalam satu transaksi, menandainya
// internal, dan memindah semua LEGAL ke sana. Mengembalikan ID LEGAL yang
// dipindah.
func (r *companyRepository) MakeParent(targetID uuid.UUID) ([]uuid.UUID, error) {
	var moved []uuid.UUID
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Induk lama dicabut lebih dulu, karena indeks unik menolak dua induk
		// walau hanya sesaat.
		if err := tx.Model(&entity.Company{}).Where("is_parent").Update("is_parent", false).Error; err != nil {
			return err
		}
		res := tx.Model(&entity.Company{}).Where("id = ?", targetID).Updates(map[string]any{"is_parent": true, "is_internal": true})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Raw(`
			UPDATE users SET company_id = ?, updated_at = now()
			WHERE role = ? AND deleted_at IS NULL AND company_id IS DISTINCT FROM ?
			RETURNING id
		`, targetID, entity.RoleLegal, targetID).Scan(&moved).Error
	})
	return moved, err
}
