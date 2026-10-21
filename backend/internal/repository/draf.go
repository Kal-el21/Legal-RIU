package repository

import (
	"legal-riu-portal/internal/entity"

	"gorm.io/gorm"
)

// bukanDraf menyaring draf pengajuan (status DRAFT), yang hanya boleh dilihat pemiliknya
// lewat daftar ?status=DRAFT. Kolom disebut lengkap dengan tabelnya supaya tidak bentrok
// dengan users.status saat tabel users ikut di-join.
func bukanDraf(tabel string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(tabel+".status <> ?", entity.StatusDraft)
	}
}
