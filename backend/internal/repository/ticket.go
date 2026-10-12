package repository

import (
	"time"

	"gorm.io/gorm"
)

// nextTicketSequence mengembalikan nomor urut tiket berikutnya untuk sebuah
// awalan di bulan berjalan, mis. 3 kalau LO-202610-0002 adalah yang terbesar.
//
// Nomor diambil dari yang terbesar yang pernah dipakai, termasuk baris yang
// sudah dihapus (soft delete), karena indeks unik ticket_number juga mencakup
// baris itu. Menghitung jumlah baris yang masih ada akan menghasilkan nomor
// yang sudah terpakai begitu ada satu baris dihapus.
func nextTicketSequence(db *gorm.DB, model any, prefix string) (int, error) {
	var last int
	err := db.Unscoped().Model(model).
		Where("ticket_number LIKE ?", prefix+"-"+time.Now().Format("200601")+"-%").
		Select("COALESCE(MAX(CAST(SUBSTRING(ticket_number FROM '[0-9]+$') AS INTEGER)), 0)").
		Scan(&last).Error
	return last + 1, err
}
