package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// jumlahAudit: banyaknya baris audit untuk satu jenis entitas, aksi, dan ID entitas. Baris
// impor Excel memakai ID kosong (uuid.Nil).
func jumlahAudit(t *testing.T, jenis string, aksi entity.AuditAction, id string) int64 {
	t.Helper()
	return hitungBaris(t, "SELECT COUNT(*) FROM audit_logs WHERE entity_type = ? AND action = ? AND entity_id = ?", jenis, aksi, id)
}

// harapSatuAuditBaru menjalankan aksi lalu memeriksa bahwa tepat satu baris audit baru tercatat
// untuk jenis entitas, aksi, dan ID itu (PR A tahap A.1b).
func harapSatuAuditBaru(t *testing.T, jenis string, aksi entity.AuditAction, id, konteks string, jalankan func()) {
	t.Helper()
	sebelum := jumlahAudit(t, jenis, aksi, id)
	jalankan()
	if n := jumlahAudit(t, jenis, aksi, id) - sebelum; n != 1 {
		t.Errorf("%s: %d baris audit %s %s baru, harap 1", konteks, n, jenis, aksi)
	}
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

// Login lewat proxy (nginx) dengan X-Forwarded-For berisi beberapa alamat tetap tercatat, dengan
// alamat pengguna saja. Dulu seluruh isi header disimpan, lebih panjang dari kolom ip_address
// (45 karakter), sehingga baris login gagal tersimpan tanpa galat (PR A tahap A.1b).
func TestLoginLewatProxyTercatat(t *testing.T) {
	const surel = "tes-login-proxy@example.com"
	userKhusus(t, surel)
	var idUser string
	dbTes.Raw("SELECT id::text FROM users WHERE email = ?", surel).Scan(&idUser)

	isi, _ := json.Marshal(map[string]string{"email": surel, "password": sandiTes})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(isi))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 10.10.10.11, 10.10.10.12, 10.10.10.13, 10.10.10.14")
	harapKode(t, kirim(t, "", req), http.StatusOK, "login lewat proxy")

	var ip []string
	dbTes.Raw("SELECT ip_address FROM audit_logs WHERE action = ? AND entity_id = ?", entity.ActionLogin, idUser).Scan(&ip)
	if len(ip) != 1 || ip[0] != "203.0.113.7" {
		t.Errorf("baris login lewat proxy: %v, harap satu baris dengan IP 203.0.113.7", ip)
	}
}

// auditTerakhir: baris audit terbaru untuk satu data.
func auditTerakhir(t *testing.T, jenis, id string) barisAudit {
	t.Helper()
	var b barisAudit
	if err := dbTes.Raw("SELECT action, old_value, new_value, description FROM audit_logs WHERE entity_type = ? AND entity_id = ? ORDER BY created_at DESC, id DESC LIMIT 1", jenis, id).Scan(&b).Error; err != nil {
		t.Fatal(err)
	}
	return b
}

// harapAuditTerakhir: baris audit terbaru data itu beraksi `aksi`, dengan nilai lama dan baru, dan
// keterangannya memuat setiap teks di `memuat`.
func harapAuditTerakhir(t *testing.T, jenis, id string, aksi entity.AuditAction, lama, baru string, memuat ...string) {
	t.Helper()
	b := auditTerakhir(t, jenis, id)
	nilai := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	ket := nilai(b.Description)
	if b.Action != string(aksi) || nilai(b.OldValue) != lama || nilai(b.NewValue) != baru {
		t.Errorf("audit %s: %s %q -> %q, harap %s %q -> %q (keterangan %q)", jenis, b.Action, nilai(b.OldValue), nilai(b.NewValue), aksi, lama, baru, ket)
	}
	for _, m := range memuat {
		if !strings.Contains(ket, m) {
			t.Errorf("keterangan audit %s %q tidak memuat %q", jenis, ket, m)
		}
	}
}

// Perubahan status Legal Opinion, Review Dokumen, dan perkara mencatat status lama, status baru, dan
// catatan approver, termasuk status yang berubah lewat unggah hasil dan lewat form ubah perkara;
// keterangannya berbahasa Indonesia (PR Audit Log tahap L.3).
func TestAuditStatusLamaDanCatatan(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	label := map[string]string{"LegalOpinion": "Legal Opinion", "ReviewDocument": "Review Dokumen"}
	jenisDari := map[string]string{"legal_opinions": "legal_opinion", "document_reviews": "document_review"}

	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			jenis, nama := jenisDari[m.tabel], label[m.nama]
			p := m.buat(t, pemilik)
			harapAuditTerakhir(t, jenis, p.ID, entity.ActionUserUpdate, "", "", "Pengajuan "+nama+" "+p.TicketNumber+" dibuat")

			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			harapAuditTerakhir(t, jenis, p.ID, entity.ActionStatusChange, "SUBMITTED", "UNDER_REVIEW", "Status "+nama+" "+p.TicketNumber+" diubah dari SUBMITTED ke UNDER_REVIEW")
			if ket := auditTerakhir(t, jenis, p.ID).Description; ket != nil && strings.Contains(*ket, "Catatan") {
				t.Errorf("perubahan status tanpa catatan menulis catatan: %q", *ket)
			}

			m.ubahStatus(t, legal, "/legal", p.ID, "NEED_REVISION", "lengkapi data", http.StatusOK)
			harapAuditTerakhir(t, jenis, p.ID, entity.ActionStatusChange, "UNDER_REVIEW", "NEED_REVISION", ". Catatan: lengkapi data")

			harapKode(t, panggil(t, pemilik, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil), http.StatusOK, "ajukan ulang")
			harapAuditTerakhir(t, jenis, p.ID, entity.ActionUserUpdate, "", "", "Pengajuan "+nama+" "+p.TicketNumber+" diajukan ulang")

			if !tanpaMinIO {
				m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
				harapKode(t, panggilForm(t, legal, http.MethodPost, m.url("/legal", "/", p.ID, "/result"),
					map[string]string{"notes": "sudah dikaji"}, berkas{"result", "hasil.pdf", []byte("hasil")}), http.StatusOK, "unggah hasil")
				harapAuditTerakhir(t, jenis, p.ID, entity.ActionFileUpload, "UNDER_REVIEW", "COMPLETED",
					"diunggah: hasil.pdf", "status diubah dari UNDER_REVIEW ke COMPLETED", ". Catatan: sudah dikaji")
			}

			h := m.buat(t, pemilik)
			harapKode(t, panggil(t, pemilik, http.MethodDelete, m.url("", "/", h.ID), nil), http.StatusOK, "hapus pengajuan")
			harapAuditTerakhir(t, jenis, h.ID, entity.ActionDelete, "", "", "Pengajuan "+nama+" dihapus")
		})
	}

	t.Run("Perkara", func(t *testing.T) {
		p := buatPerkara(t, legal)
		harapAuditTerakhir(t, "legal_case", p.ID, entity.ActionUserUpdate, "", "", "Perkara "+p.TicketNumber+" dibuat")

		// LEGAL dan ADMIN mengubah status lewat form ubah perkara.
		form := formPerkara(t)
		form["current_status"] = "Banding"
		harapKode(t, panggil(t, legal, http.MethodPut, perkaraLegal+"/"+p.ID, form), http.StatusOK, "mengubah status lewat form")
		harapAuditTerakhir(t, "legal_case", p.ID, entity.ActionStatusChange, "Berjalan", "Banding", "Data perkara "+p.TicketNumber+" diubah; status dari Berjalan ke Banding")

		form["case_name"] = "Perkara tes diubah"
		harapKode(t, panggil(t, legal, http.MethodPut, perkaraLegal+"/"+p.ID, form), http.StatusOK, "mengubah tanpa ganti status")
		harapAuditTerakhir(t, "legal_case", p.ID, entity.ActionUserUpdate, "", "", "Data perkara "+p.TicketNumber+" diubah")

		harapKode(t, panggil(t, masukPeran(t, entity.RoleLegalAU), http.MethodPatch, perkaraLegalAU+"/"+p.ID+"/status",
			map[string]string{"current_status": "Kasasi"}), http.StatusOK, "Legal AU mengubah status")
		harapAuditTerakhir(t, "legal_case", p.ID, entity.ActionStatusChange, "Banding", "Kasasi", "Status perkara "+p.TicketNumber+" diubah dari Banding ke Kasasi")
	})
}
