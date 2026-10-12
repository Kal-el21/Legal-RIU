package service

import (
	"testing"
	"time"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
)

// Aturan isi lonceng: siapa menerima pengajuan berstatus apa, dan kapan
// pengajuan menjadi peringatan. Ambang bawaan: kuning 3 hari, merah 14 hari.
func TestBuildReminderRow(t *testing.T) {
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	pemilik, orangLain := uuid.New(), uuid.New()
	hariLalu := func(n int) *time.Time {
		waktu := now.Add(-time.Duration(n) * 24 * time.Hour)
		return &waktu
	}

	hanyaPemilik := ReminderScope{Owner: true}
	pemroses := ReminderScope{Owner: true, Processor: true}

	kasus := []struct {
		nama        string
		scope       ReminderScope
		viewer      uuid.UUID
		status      entity.SubmissionStatus
		sejak       int // hari sejak masuk status sekarang
		kuning      int
		merah       int
		masuk       bool
		level       WarningLevel
		berubahLalu int // hari sejak baris terakhir berubah
	}{
		{"pemilik diberi kabar saat direview", hanyaPemilik, pemilik, entity.StatusUnderReview, 0, 3, 14, true, WarningLevelNone, 0},
		{"pemilik diberi kabar saat selesai", hanyaPemilik, pemilik, entity.StatusCompleted, 10, 3, 14, true, WarningLevelNone, 10},
		{"kabar selesai tepat 30 hari masih tampil", hanyaPemilik, pemilik, entity.StatusCompleted, 30, 3, 14, true, WarningLevelNone, 30},
		{"kabar selesai lebih dari 30 hari tidak tampil", hanyaPemilik, pemilik, entity.StatusCompleted, 31, 3, 14, false, "", 0},
		{"kabar ditolak lebih dari 30 hari tidak tampil", hanyaPemilik, pemilik, entity.StatusRejected, 31, 3, 14, false, "", 0},
		{"permintaan revisi lama tetap tampil", hanyaPemilik, pemilik, entity.StatusNeedRevision, 40, 3, 14, true, WarningLevelNone, 40},
		{"pemilik tidak diberi kabar atas pengajuannya sendiri", hanyaPemilik, pemilik, entity.StatusSubmitted, 0, 3, 14, false, "", 0},
		{"pemilik tidak diberi kabar atas ajukan ulangnya sendiri", hanyaPemilik, pemilik, entity.StatusResubmitted, 0, 3, 14, false, "", 0},
		{"pemilik tidak mendapat peringatan", hanyaPemilik, pemilik, entity.StatusUnderReview, 30, 3, 14, true, WarningLevelNone, 30},
		{"bukan pemilik tidak melihat pengajuan orang", hanyaPemilik, orangLain, entity.StatusNeedRevision, 0, 3, 14, false, "", 0},

		{"pemroses diberi kabar pengajuan baru", pemroses, orangLain, entity.StatusSubmitted, 0, 3, 14, true, WarningLevelNone, 0},
		{"pemroses diberi kabar pengajuan ulang", pemroses, orangLain, entity.StatusResubmitted, 1, 3, 14, true, WarningLevelNone, 1},
		{"sedang direview belum melewati ambang tidak masuk", pemroses, orangLain, entity.StatusUnderReview, 2, 3, 14, false, "", 0},
		{"sedang direview melewati ambang kuning", pemroses, orangLain, entity.StatusUnderReview, 5, 3, 14, true, WarningLevelYellow, 2},
		{"belum diambil melewati ambang merah", pemroses, orangLain, entity.StatusSubmitted, 20, 3, 14, true, WarningLevelRed, 6},
		{"tepat di ambang merah sudah merah", pemroses, orangLain, entity.StatusSubmitted, 14, 3, 14, true, WarningLevelRed, 0},
		{"menunggu revisi pemilik bukan urusan pemroses", pemroses, orangLain, entity.StatusNeedRevision, 30, 3, 14, false, "", 0},
		{"selesai tidak pernah jadi peringatan", pemroses, orangLain, entity.StatusCompleted, 30, 3, 14, false, "", 0},

		{"kedua ambang nonaktif: tanpa peringatan", pemroses, orangLain, entity.StatusSubmitted, 30, 0, 0, true, WarningLevelNone, 30},
		{"hanya ambang merah aktif", pemroses, orangLain, entity.StatusSubmitted, 5, 0, 14, true, WarningLevelNone, 5},
		{"hanya ambang kuning aktif", pemroses, orangLain, entity.StatusSubmitted, 30, 3, 0, true, WarningLevelYellow, 27},
	}

	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			src := reminderSource{
				submissionType: "legal_opinion", id: uuid.New(), ownerID: pemilik,
				status: k.status, createdAt: *hariLalu(90), statusUpdatedAt: hariLalu(k.sejak),
			}
			row, masuk := buildReminderRow(src, k.scope, k.viewer, k.kuning, k.merah, now)
			if masuk != k.masuk {
				t.Fatalf("masuk lonceng: %v, harap %v", masuk, k.masuk)
			}
			if !masuk {
				return
			}
			if row.item.WarningLevel != string(k.level) {
				t.Errorf("level: %s, harap %s", row.item.WarningLevel, k.level)
			}
			if row.item.DaysWaiting != k.sejak {
				t.Errorf("lama menunggu: %d hari, harap %d", row.item.DaysWaiting, k.sejak)
			}
			if !row.changedAt.Equal(*hariLalu(k.berubahLalu)) {
				t.Errorf("terakhir berubah: %s, harap %d hari lalu", row.changedAt, k.berubahLalu)
			}
		})
	}
}

// Pengajuan lama yang belum pernah berganti status dihitung dari tanggal dibuat,
// dan catatan pemroses hanya ikut selagi statusnya masih revisi atau ditolak.
func TestBuildReminderRowWaktuDanCatatan(t *testing.T) {
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	pemilik := uuid.New()
	dibuat := now.Add(-5 * 24 * time.Hour)

	row, masuk := buildReminderRow(reminderSource{ownerID: uuid.New(), status: entity.StatusSubmitted, createdAt: dibuat},
		ReminderScope{Processor: true}, pemilik, 3, 14, now)
	if !masuk || row.item.DaysWaiting != 5 || row.item.WarningLevel != string(WarningLevelYellow) {
		t.Errorf("tanpa status_updated_at: masuk %v, %d hari, level %s", masuk, row.item.DaysWaiting, row.item.WarningLevel)
	}

	for status, harap := range map[entity.SubmissionStatus]string{
		entity.StatusNeedRevision: "lengkapi data",
		entity.StatusRejected:     "lengkapi data",
		entity.StatusUnderReview:  "",
		entity.StatusCompleted:    "",
	} {
		row, _ := buildReminderRow(reminderSource{ownerID: pemilik, status: status, note: "lengkapi data", createdAt: dibuat},
			ReminderScope{Owner: true}, pemilik, 3, 14, now)
		if row.item.Note != harap {
			t.Errorf("catatan saat %s: %q, harap %q", status, row.item.Note, harap)
		}
	}
}
