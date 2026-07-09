package seed

import (
	"errors"
	"fmt"
	"time"

	"legal-riu-portal/internal/entity"

	"gorm.io/gorm"
)

func PrepareLegalCasePICMigration(db *gorm.DB) error {
	var nonUUIDCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'pic'
			AND data_type <> 'uuid'
	`).Scan(&nonUUIDCount).Error; err != nil {
		return err
	}
	if nonUUIDCount == 0 {
		return nil
	}

	var legacyCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'pic_legacy'
	`).Scan(&legacyCount).Error; err != nil {
		return err
	}
	if legacyCount > 0 {
		return errors.New("legal_cases.pic masih bukan uuid dan pic_legacy sudah ada; migrasi manual diperlukan")
	}

	return db.Exec(`ALTER TABLE legal_cases RENAME COLUMN pic TO pic_legacy`).Error
}

func BackfillLegalCaseTypeCategory(db *gorm.DB) error {
	var caseTypeColCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'case_type'
	`).Scan(&caseTypeColCount).Error; err != nil {
		return err
	}
	if caseTypeColCount == 0 {
		return nil
	}

	var categoryColCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'category'
	`).Scan(&categoryColCount).Error; err != nil {
		return err
	}
	if categoryColCount == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			UPDATE legal_cases lc
			SET case_type_id = ct.id
			FROM case_types ct
			WHERE lc.case_type = ct.code
		`).Error; err != nil {
			return err
		}

		if err := tx.Exec(`
			UPDATE legal_cases lc
			SET category_id = cc.id
			FROM case_categories cc
			WHERE lc.category = cc.code
		`).Error; err != nil {
			return err
		}

		return nil
	})
}

func DropOldLegalCaseColumns(db *gorm.DB) error {
	var caseTypeColCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'case_type'
	`).Scan(&caseTypeColCount).Error; err != nil {
		return err
	}
	if caseTypeColCount == 0 {
		return nil
	}

	var categoryColCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'category'
	`).Scan(&categoryColCount).Error; err != nil {
		return err
	}
	if categoryColCount == 0 {
		return nil
	}

	if err := db.Exec(`ALTER TABLE legal_cases DROP COLUMN case_type`).Error; err != nil {
		return err
	}
	return db.Exec(`ALTER TABLE legal_cases DROP COLUMN category`).Error
}

func BackfillLegalCaseDefaults(db *gorm.DB) error {
	if err := db.Exec(`
		UPDATE legal_cases
		SET category_id = (SELECT id FROM case_categories LIMIT 1)
		WHERE category_id IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		UPDATE legal_cases
		SET case_type_id = (SELECT id FROM case_types LIMIT 1)
		WHERE case_type_id IS NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		UPDATE legal_cases
		SET company_id = (SELECT id FROM companies LIMIT 1)
		WHERE company_id IS NULL
	`).Error; err != nil {
		return err
	}

	return nil
}

func EnforceLegalCaseNotNull(db *gorm.DB) error {
	if err := db.Exec(`ALTER TABLE legal_cases ALTER COLUMN category_id SET NOT NULL`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE legal_cases ALTER COLUMN case_type_id SET NOT NULL`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE legal_cases ALTER COLUMN company_id SET NOT NULL`).Error; err != nil {
		return err
	}
	return nil
}

func BackfillLegalCaseTicketNumbers(db *gorm.DB) error {
	var count int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'ticket_number'
	`).Scan(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	if err := db.Exec(`UPDATE legal_cases SET ticket_number = '' WHERE ticket_number IS NULL`).Error; err != nil {
		return err
	}

	var existingCount int64
	if err := db.Model(&entity.LegalCase{}).Where("ticket_number <> ''").Count(&existingCount).Error; err != nil {
		return err
	}

	now := time.Now()
	prefix := "LC"
	month := now.Format("200601")

	var cases []entity.LegalCase
	if err := db.Where("ticket_number = ''").Order("created_at ASC, id ASC").Find(&cases).Error; err != nil {
		return err
	}
	if len(cases) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for i, lc := range cases {
			sequence := int(existingCount) + i + 1
			ticket := fmt.Sprintf("%s-%s-%04d", prefix, month, sequence)
			if err := tx.Model(&lc).Update("ticket_number", ticket).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func AddStatusUpdatedAtColumns(db *gorm.DB) error {
	// Add status_updated_at to legal_opinions if not exists
	if !db.Migrator().HasColumn(&entity.LegalOpinion{}, "status_updated_at") {
		if err := db.Migrator().AddColumn(&entity.LegalOpinion{}, "status_updated_at"); err != nil {
			return err
		}
	}

	// Add status_updated_at to document_reviews if not exists
	if !db.Migrator().HasColumn(&entity.DocumentReview{}, "status_updated_at") {
		if err := db.Migrator().AddColumn(&entity.DocumentReview{}, "status_updated_at"); err != nil {
			return err
		}
	}

	// Add status_updated_at to legal_cases if not exists
	if !db.Migrator().HasColumn(&entity.LegalCase{}, "status_updated_at") {
		if err := db.Migrator().AddColumn(&entity.LegalCase{}, "status_updated_at"); err != nil {
			return err
		}
	}

	// Add ticket_number to legal_cases if not exists
	if err := db.Exec(`
		ALTER TABLE legal_cases ADD COLUMN IF NOT EXISTS ticket_number text DEFAULT ''
	`).Error; err != nil {
		return err
	}

	return nil
}

func EnforceLegalCaseTicketNumberNotNull(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&entity.LegalCase{}, "ticket_number") {
		return nil
	}
	if err := db.Exec(`UPDATE legal_cases SET ticket_number = '' WHERE ticket_number IS NULL`).Error; err != nil {
		return err
	}
	return db.Exec(`ALTER TABLE legal_cases ALTER COLUMN ticket_number SET NOT NULL`).Error
}

func MigrateTechnicalReserveToDecimal(db *gorm.DB) error {
	var colCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = current_schema()
			AND table_name = 'legal_cases'
			AND column_name = 'technical_reserve'
			AND data_type = 'character varying'
	`).Scan(&colCount).Error; err != nil {
		return err
	}
	if colCount == 0 {
		return nil
	}

	return db.Exec(`
		ALTER TABLE legal_cases
		ALTER COLUMN technical_reserve TYPE decimal(18,2)
		USING CASE
			WHEN technical_reserve ~ '^[0-9]+(\.[0-9]+)?$' THEN technical_reserve::decimal(18,2)
			ELSE NULL
		END
	`).Error
}

func RunAllMigrationsAndSeeds(db *gorm.DB) error {
	if err := PrepareLegalCasePICMigration(db); err != nil {
		return err
	}
	if err := MigrateTechnicalReserveToDecimal(db); err != nil {
		return err
	}

	if err := AddStatusUpdatedAtColumns(db); err != nil {
		return err
	}
	if err := BackfillLegalCaseTicketNumbers(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&entity.Division{},
		&entity.User{},
		&entity.RefreshToken{},
		&entity.Permission{},
		&entity.RolePermission{},
		&entity.UserPermissionOverride{},
		&entity.LegalOpinion{},
		&entity.LegalOpinionAttachment{},
		&entity.LegalOpinionResult{},
		&entity.DocumentReview{},
		&entity.DocumentReviewAttachment{},
		&entity.DocumentReviewResult{},
		&entity.Regency{},
		&entity.Cedant{},
		&entity.LegalCase{},
		&entity.CaseChronology{},
		&entity.AuditLog{},
		&entity.NotificationSetting{},
		&entity.UserSettings{},
		&entity.Company{},
		&entity.PurposeType{},
		&entity.CaseType{},
		&entity.CaseCategory{},
		&entity.DocumentType{},
		&entity.LegalMaterial{},
	); err != nil {
		return err
	}

	if err := BackfillLegalCaseTypeCategory(db); err != nil {
		return err
	}
	if err := DropOldLegalCaseColumns(db); err != nil {
		return err
	}

	if err := SeedRegencies(db); err != nil {
		return err
	}
	if err := SeedDivisions(db); err != nil {
		return err
	}
	if err := BackfillUserDivisionIDs(db); err != nil {
		return err
	}
	if err := SeedCompanies(db); err != nil {
		return err
	}
	if err := SeedPurposeTypes(db); err != nil {
		return err
	}
	if err := SeedDocumentTypes(db); err != nil {
		return err
	}
	if err := SeedCaseTypes(db); err != nil {
		return err
	}
	if err := SeedCaseCategories(db); err != nil {
		return err
	}
	if err := SeedPermissions(db); err != nil {
		return err
	}
	if err := BackfillLegalCaseDefaults(db); err != nil {
		return err
	}
	if err := EnforceLegalCaseNotNull(db); err != nil {
		return err
	}
	if err := EnforceLegalCaseTicketNumberNotNull(db); err != nil {
		return err
	}
	if err := SeedNotificationSettings(db); err != nil {
		return err
	}

	return nil
}
