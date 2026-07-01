package seed

import (
	"errors"

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
