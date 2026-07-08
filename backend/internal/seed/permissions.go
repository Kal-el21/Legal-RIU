package seed

import (
	"errors"

	"legal-riu-portal/internal/entity"

	"gorm.io/gorm"
)

func SeedPermissions(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		permissions := permissionSeedData()
		for _, permission := range permissions {
			if err := upsertSeedPermission(tx, permission); err != nil {
				return err
			}
		}

		var stored []entity.Permission
		if err := tx.Where("is_active = ?", true).Find(&stored).Error; err != nil {
			return err
		}

		permissionByCode := make(map[string]entity.Permission, len(stored))
		for _, permission := range stored {
			permissionByCode[permission.Code] = permission
		}

		rolePermissions := rolePermissionSeedData(permissionByCode)
		for role, codes := range rolePermissions {
			if err := tx.Unscoped().Where("role = ?", role).Delete(&entity.RolePermission{}).Error; err != nil {
				return err
			}

			for _, code := range codes {
				permission, ok := permissionByCode[code]
				if !ok {
					continue
				}
				if err := tx.Create(&entity.RolePermission{
					Role:         role,
					PermissionID: permission.ID,
				}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func upsertSeedPermission(tx *gorm.DB, permission entity.Permission) error {
	var existing entity.Permission
	err := tx.Where("code = ?", permission.Code).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&permission).Error
		}
		return err
	}

	return nil
}

func permissionSeedData() []entity.Permission {
	return []entity.Permission{
		perm("dashboard.user.view", "dashboard", "view", "own", "Dashboard User", "Melihat dashboard user"),
		perm("dashboard.admin.view", "dashboard", "view", "all", "Dashboard Admin", "Melihat dashboard admin"),
		perm("dashboard.legal.view", "dashboard", "view", "all", "Dashboard Legal", "Melihat dashboard legal"),

		perm("legal_opinion.view.own", "legal_opinion", "view", "own", "Legal Opinion - Lihat Milik Sendiri", "Melihat legal opinion milik sendiri"),
		perm("legal_opinion.view.all", "legal_opinion", "view", "all", "Legal Opinion - Lihat Semua", "Melihat semua legal opinion"),
		perm("legal_opinion.create.own", "legal_opinion", "create", "own", "Legal Opinion - Buat", "Membuat pengajuan legal opinion"),
		perm("legal_opinion.update.own", "legal_opinion", "update", "own", "Legal Opinion - Edit", "Mengedit legal opinion milik sendiri"),
		perm("legal_opinion.delete.own", "legal_opinion", "delete", "own", "Legal Opinion - Hapus", "Menghapus legal opinion milik sendiri"),
		perm("legal_opinion.resubmit.own", "legal_opinion", "resubmit", "own", "Legal Opinion - Ajukan Ulang", "Mengajukan ulang legal opinion milik sendiri"),
		perm("legal_opinion.update_status.all", "legal_opinion", "update_status", "all", "Legal Opinion - Ubah Status", "Mengubah status legal opinion"),
		perm("legal_opinion.upload_result.all", "legal_opinion", "upload_result", "all", "Legal Opinion - Upload Hasil", "Mengunggah hasil kajian legal opinion"),
		perm("legal_opinion.download.all", "legal_opinion", "download", "all", "Legal Opinion - Download", "Mengunduh dokumen legal opinion"),

		perm("document_review.view.own", "document_review", "view", "own", "Review Dokumen - Lihat Milik Sendiri", "Melihat review dokumen milik sendiri"),
		perm("document_review.view.all", "document_review", "view", "all", "Review Dokumen - Lihat Semua", "Melihat semua review dokumen"),
		perm("document_review.create.own", "document_review", "create", "own", "Review Dokumen - Buat", "Membuat pengajuan review dokumen"),
		perm("document_review.update.own", "document_review", "update", "own", "Review Dokumen - Edit", "Mengedit review dokumen milik sendiri"),
		perm("document_review.delete.own", "document_review", "delete", "own", "Review Dokumen - Hapus", "Menghapus review dokumen milik sendiri"),
		perm("document_review.resubmit.own", "document_review", "resubmit", "own", "Review Dokumen - Ajukan Ulang", "Mengajukan ulang review dokumen milik sendiri"),
		perm("document_review.update_status.all", "document_review", "update_status", "all", "Review Dokumen - Ubah Status", "Mengubah status review dokumen"),
		perm("document_review.upload_result.all", "document_review", "upload_result", "all", "Review Dokumen - Upload Hasil", "Mengunggah hasil review dokumen"),
		perm("document_review.download.all", "document_review", "download", "all", "Review Dokumen - Download", "Mengunduh dokumen review dokumen"),

		perm("case_management.view", "case_management", "view", "all", "Case Management - Lihat", "Melihat case management"),
		perm("case_management.create", "case_management", "create", "all", "Case Management - Buat", "Membuat case"),
		perm("case_management.update", "case_management", "update", "all", "Case Management - Edit", "Mengedit case"),
		perm("case_management.delete", "case_management", "delete", "all", "Case Management - Hapus", "Menghapus case"),
		perm("case_management.update_status", "case_management", "update_status", "all", "Case Management - Ubah Status", "Mengubah status case"),
		perm("case_management.manage_document", "case_management", "manage_document", "all", "Case Management - Dokumen", "Mengelola dokumen case"),
		perm("case_management.manage_chronology", "case_management", "manage_chronology", "all", "Case Management - Kronologi", "Mengelola kronologi case"),
		perm("case_management.manage_reference", "case_management", "manage_reference", "all", "Case Management - Referensi", "Mengelola referensi cedant dan lokasi case"),

		perm("user_management.view", "user_management", "view", "all", "User Management - Lihat", "Melihat daftar user"),
		perm("user_management.create", "user_management", "create", "all", "User Management - Buat", "Membuat user"),
		perm("user_management.update", "user_management", "update", "all", "User Management - Edit", "Mengedit user"),
		perm("user_management.update_status", "user_management", "update_status", "all", "User Management - Status", "Mengubah status user"),
		perm("user_management.reset_password", "user_management", "reset_password", "all", "User Management - Reset Password", "Mereset password user"),
		perm("user_management.manage_permissions", "user_management", "manage_permissions", "all", "User Management - Permission", "Mengelola permission user"),
		perm("user_management.delete", "user_management", "delete", "all", "User Management - Hapus", "Menghapus user"),

		perm("audit_log.view", "audit_log", "view", "all", "Audit Log - Lihat", "Melihat audit log"),
		perm("master_data.view", "master_data", "view", "all", "Master Data - Lihat", "Melihat master data"),
		perm("master_data.manage", "master_data", "manage", "all", "Master Data - Kelola", "Mengelola master data"),
		perm("notification_setting.manage", "notification_setting", "manage", "all", "Notifikasi - Kelola", "Mengelola pengaturan notifikasi"),
		perm("legal_material.view", "legal_material", "view", "all", "Materi Legal - Lihat", "Melihat materi legal"),
		perm("legal_material.manage", "legal_material", "manage", "all", "Materi Legal - Kelola", "Mengelola materi legal"),

		perm("report.legal_case.view", "report", "legal_case", "all", "Report - Legal Case", "Melihat laporan legal case"),
		perm("report.legal_opinion.view", "report", "legal_opinion", "all", "Report - Legal Opinion", "Melihat laporan legal opinion"),
		perm("report.document_review.view", "report", "document_review", "all", "Report - Review Dokumen", "Melihat laporan review dokumen"),
	}
}

func rolePermissionSeedData(permissionByCode map[string]entity.Permission) map[entity.UserRole][]string {
	allCodes := make([]string, 0, len(permissionByCode))
	for code := range permissionByCode {
		allCodes = append(allCodes, code)
	}

	user := []string{
		"dashboard.user.view",
		"legal_opinion.view.own",
		"legal_opinion.create.own",
		"legal_opinion.update.own",
		"legal_opinion.delete.own",
		"legal_opinion.resubmit.own",
		"legal_opinion.download.all",
		"document_review.view.own",
		"document_review.create.own",
		"document_review.update.own",
		"document_review.delete.own",
		"document_review.resubmit.own",
		"document_review.download.all",
		"legal_material.view",
	}

	legal := []string{
		"dashboard.legal.view",
		"legal_opinion.view.all",
		"legal_opinion.update_status.all",
		"legal_opinion.upload_result.all",
		"legal_opinion.download.all",
		"document_review.view.all",
		"document_review.update_status.all",
		"document_review.upload_result.all",
		"document_review.download.all",
		"case_management.view",
		"case_management.create",
		"case_management.update",
		"case_management.delete",
		"case_management.update_status",
		"case_management.manage_document",
		"case_management.manage_chronology",
		"case_management.manage_reference",
		"audit_log.view",
		"legal_material.view",
		"legal_material.manage",
		"report.legal_case.view",
		"report.legal_opinion.view",
		"report.document_review.view",
	}

	external := []string{
		"case_management.view",
		"case_management.manage_document",
		"case_management.manage_chronology",
	}

	legalAU := []string{
		"case_management.view",
		"case_management.create",
		"case_management.update_status",
		"case_management.manage_chronology",
		"legal_material.view",
		"legal_material.manage",
	}

	return map[entity.UserRole][]string{
		entity.RoleAdmin:    allCodes,
		entity.RoleUser:     user,
		entity.RoleLegal:    legal,
		entity.RoleExternal: external,
		entity.RoleLegalAU:  legalAU,
	}
}

func perm(code, feature, action, scope, label, description string) entity.Permission {
	return entity.Permission{
		Code:        code,
		Feature:     feature,
		Action:      action,
		Scope:       scope,
		Label:       label,
		Description: description,
		IsActive:    true,
	}
}
