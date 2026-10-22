package router

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
)

// barisAudit: satu baris audit_logs untuk sebuah pengajuan perjanjian.
type barisAudit struct {
	Action      string
	OldValue    *string
	NewValue    *string
	Description *string
}

func auditPerjanjian(t *testing.T, id string) []barisAudit {
	t.Helper()
	var rows []barisAudit
	if err := dbTes.Raw("SELECT action, old_value, new_value, description FROM audit_logs WHERE entity_type = 'agreement_document' AND entity_id = ? ORDER BY created_at, id", id).Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}

// harapAuditBaru memeriksa bahwa sejak `sebelum` baris, tepat satu baris audit baru tercatat
// dengan aksi itu dan keterangan yang memuat `memuat`; baris itu dikembalikan.
func harapAuditBaru(t *testing.T, id string, sebelum int, aksi entity.AuditAction, memuat, konteks string) barisAudit {
	t.Helper()
	rows := auditPerjanjian(t, id)
	if len(rows) != sebelum+1 {
		t.Fatalf("%s: %d baris audit baru, harap 1 (%+v)", konteks, len(rows)-sebelum, rows)
	}
	b := rows[len(rows)-1]
	if b.Action != string(aksi) || b.Description == nil || !strings.Contains(*b.Description, memuat) {
		ket := ""
		if b.Description != nil {
			ket = *b.Description
		}
		t.Errorf("%s: aksi %s, keterangan %q; harap %s yang memuat %q", konteks, b.Action, ket, aksi, memuat)
	}
	return b
}

// Setiap aksi pada pengajuan perjanjian tercatat di audit log seperti Legal Opinion dan
// Review Dokumen; aksi yang ditolak tidak meninggalkan baris (PR A tahap A.1).
func TestAuditPerjanjian(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)

	p := buatPerjanjian(t, pemilik, formPerjanjian())
	harapAuditBaru(t, p.ID, 0, entity.ActionUserUpdate, p.TicketNumber+" dibuat", "membuat perjanjian")

	harapKode(t, panggil(t, pemilik, http.MethodPut, urlPerjanjian("", "/", p.ID), map[string]any{"form_data": formPerjanjian()}), http.StatusOK, "mengubah isian")
	harapAuditBaru(t, p.ID, 1, entity.ActionUserUpdate, "diubah", "mengubah isian")

	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	b := harapAuditBaru(t, p.ID, 2, entity.ActionStatusChange, "dari SUBMITTED ke UNDER_REVIEW", "mulai memeriksa")
	if b.OldValue == nil || *b.OldValue != "SUBMITTED" || b.NewValue == nil || *b.NewValue != "UNDER_REVIEW" {
		t.Errorf("nilai lama dan baru status: %v -> %v", b.OldValue, b.NewValue)
	}

	// Ditolak: tidak ada baris baru.
	statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "", http.StatusBadRequest) // catatan wajib
	if n := len(auditPerjanjian(t, p.ID)); n != 3 {
		t.Errorf("perubahan status yang ditolak meninggalkan baris audit: %d baris, harap 3", n)
	}

	harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_place": "Bandung"}), http.StatusOK, "data persetujuan")
	harapAuditBaru(t, p.ID, 3, entity.ActionUserUpdate, "Data persetujuan", "data persetujuan")

	statusPerjanjian(t, legal, p.ID, "NEED_REVISION", "lengkapi lampiran", http.StatusOK)
	harapAuditBaru(t, p.ID, 4, entity.ActionStatusChange, "Catatan: lengkapi lampiran", "kembalikan untuk revisi")

	harapKode(t, panggilForm(t, pemilik, http.MethodPost, urlPerjanjian("", "/", p.ID, "/resubmit"), nil, berkas{"attachments", "revisi.pdf", []byte("revisi")}), http.StatusOK, "ajukan ulang")
	harapAuditBaru(t, p.ID, 5, entity.ActionUserUpdate, "diajukan ulang dengan 1 lampiran", "ajukan ulang")

	statusPerjanjian(t, legal, p.ID, "UNDER_REVIEW", "", http.StatusOK)
	harapAuditBaru(t, p.ID, 6, entity.ActionStatusChange, "dari RESUBMITTED ke UNDER_REVIEW", "periksa ulang")
	statusPerjanjian(t, legal, p.ID, "REJECTED", "tidak sesuai", http.StatusOK)
	harapAuditBaru(t, p.ID, 7, entity.ActionStatusChange, "dari UNDER_REVIEW ke REJECTED", "tolak")

	t.Run("Hapus", func(t *testing.T) {
		h := buatPerjanjian(t, pemilik, formPerjanjian())
		harapKode(t, panggil(t, pemilik, http.MethodDelete, urlPerjanjian("", "/", h.ID), nil), http.StatusOK, "hapus perjanjian")
		harapAuditBaru(t, h.ID, 1, entity.ActionDelete, "dihapus", "hapus perjanjian")
	})

	t.Run("Draf", func(t *testing.T) {
		data, _ := json.Marshal(isiPengajuan(t, pemilik, "PKS", formPerjanjian()))
		w := panggilForm(t, pemilik, http.MethodPost, urlPerjanjian("", "/drafts"), map[string]string{"data": string(data)})
		harapKode(t, w, http.StatusCreated, "simpan draf")
		var d perjanjian
		isiData(t, w, &d)
		if n := len(auditPerjanjian(t, d.ID)); n != 0 {
			t.Errorf("menyimpan draf tercatat di audit log: %d baris", n)
		}
		w = panggil(t, pemilik, http.MethodPost, urlPerjanjian("", "/", d.ID, "/submit"), nil)
		harapKode(t, w, http.StatusOK, "kirim draf")
		var terkirim perjanjian
		isiData(t, w, &terkirim)
		harapAuditBaru(t, d.ID, 0, entity.ActionUserUpdate, terkirim.TicketNumber+" dikirim dari draf", "kirim draf")
	})
}
