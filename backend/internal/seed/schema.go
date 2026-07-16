package seed

import (
	"errors"
	"fmt"
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
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

func PrepareAgreementDocumentUserID(db *gorm.DB) error {
	if err := db.Exec(`
		ALTER TABLE agreement_documents 
		ADD COLUMN IF NOT EXISTS user_id uuid
	`).Error; err != nil {
		return err
	}

	// Kolom requester_id tidak ada di DB ini (seed lama belum pernah berjalan),
	// sehingga tidak ada data untuk di-backfill. Cukup pastikan user_id ada dan
	// bersihkan sisa kolom requester_id bila suatu saat ada.
	if err := db.Exec(`
		ALTER TABLE agreement_documents 
		DROP COLUMN IF EXISTS requester_id
	`).Error; err != nil {
		return err
	}

	return nil
}

func EnforceAgreementDocumentUserIDNotNull(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE agreement_documents 
		ALTER COLUMN user_id SET NOT NULL
	`).Error
}

func EnforceAgreementDocumentTicketNumberNotNull(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE agreement_documents 
		ALTER COLUMN ticket_number SET NOT NULL
	`).Error
}

func PrepareAgreementDocumentTicketNumber(db *gorm.DB) error {
	if err := db.Exec(`
		ALTER TABLE agreement_documents 
		ADD COLUMN IF NOT EXISTS ticket_number varchar(80)
	`).Error; err != nil {
		return err
	}

	return db.Exec(`
		UPDATE agreement_documents 
		SET ticket_number = 'PK-LEGACY-' || SUBSTRING(CAST(id AS text), 1, 8)
		WHERE ticket_number IS NULL
	`).Error
}

func PrepareAgreementDocumentDocumentTypeCode(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE agreement_documents 
		ADD COLUMN IF NOT EXISTS document_type_code varchar(50)
	`).Error
}

func PrepareAgreementDocumentFormData(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE agreement_documents 
		ADD COLUMN IF NOT EXISTS form_data jsonb DEFAULT '{}'::jsonb
	`).Error
}

func BackfillAgreementDocumentDocumentTypeCode(db *gorm.DB) error {
	return db.Exec(`
		UPDATE agreement_documents 
		SET document_type_code = 'PKS'
		WHERE document_type_code IS NULL
			OR BTRIM(document_type_code) = ''
			OR UPPER(BTRIM(document_type_code)) IN (
				'LAIN-LAIN',
				'LAIN LAIN',
				'OTHER',
				'PERJANJIAN KERJA SAMA',
				'PERJANJIAN_KERJA_SAMA'
			)
	`).Error
}

func EnforceAgreementDocumentDocumentTypeCodeNotNull(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE agreement_documents 
		ALTER COLUMN document_type_code SET NOT NULL
	`).Error
}

func EnforceAgreementDocumentFormDataNotNull(db *gorm.DB) error {
	return db.Exec(`
		ALTER TABLE agreement_documents 
		ALTER COLUMN form_data SET NOT NULL
	`).Error
}

func PrepareAgreementAttachmentColumns(db *gorm.DB) error {
	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ADD COLUMN IF NOT EXISTS agreement_document_id uuid
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ADD COLUMN IF NOT EXISTS file_name varchar(255)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ADD COLUMN IF NOT EXISTS file_path varchar(500)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ADD COLUMN IF NOT EXISTS uploaded_by uuid
	`).Error; err != nil {
		return err
	}

	return nil
}

func BackfillAgreementAttachmentColumns(db *gorm.DB) error {
	var adminID uuid.UUID
	if err := db.Raw(`
		SELECT id FROM users WHERE role = 'ADMIN' LIMIT 1
	`).Scan(&adminID).Error; err != nil {
		return err
	}

	if adminID != uuid.Nil {
		if err := db.Exec(`
			UPDATE agreement_attachments 
			SET uploaded_by = ?
			WHERE uploaded_by IS NULL
		`, adminID).Error; err != nil {
			return err
		}
	}

	return db.Exec(`
		UPDATE agreement_attachments SET file_name = '' WHERE file_name IS NULL
	`).Error
}

func EnforceAgreementAttachmentColumnsNotNull(db *gorm.DB) error {
	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ALTER COLUMN agreement_document_id SET NOT NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ALTER COLUMN file_name SET NOT NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ALTER COLUMN file_path SET NOT NULL
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		ALTER TABLE agreement_attachments 
		ALTER COLUMN uploaded_by SET NOT NULL
	`).Error; err != nil {
		return err
	}

	return nil
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

	if err := PrepareAgreementDocumentUserID(db); err != nil {
		return err
	}
	if err := PrepareAgreementDocumentTicketNumber(db); err != nil {
		return err
	}
	if err := PrepareAgreementDocumentDocumentTypeCode(db); err != nil {
		return err
	}
	if err := PrepareAgreementDocumentFormData(db); err != nil {
		return err
	}
	if err := PrepareAgreementAttachmentColumns(db); err != nil {
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
		&entity.AgreementCompanyMaster{},
		&entity.AgreementDocument{},
		&entity.AgreementAttachment{},
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
	if err := BackfillAgreementDocumentDocumentTypeCode(db); err != nil {
		return err
	}
	if err := EnforceAgreementDocumentDocumentTypeCodeNotNull(db); err != nil {
		return err
	}
	if err := EnforceAgreementDocumentFormDataNotNull(db); err != nil {
		return err
	}
	if err := EnforceAgreementDocumentTicketNumberNotNull(db); err != nil {
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
	if err := SeedAgreementCompanyMaster(db); err != nil {
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
