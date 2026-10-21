package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
)

// Pengajuan baru dengan tiga termin diterima. Persen termin terakhir dan nilai
// setiap termin dihitung server dari nilai kontrak, apa pun yang dikirim browser.
// Persen melebihi 100 dan tanpa termin ditolak dengan pesan yang menyebut termin.
func TestPengajuanTigaTermin(t *testing.T) {
	butuhMinIO(t)
	user := masukPeran(t, entity.RoleUser)
	form := formPerjanjian()
	form["termin"] = []any{
		map[string]any{"persen": 30, "nilai": 1, "syarat": "setelah penandatanganan"},
		map[string]any{"persen": "30", "syarat": "setelah laporan"},
		map[string]any{"persen": 99, "nilai": 999, "syarat": "setelah serah terima"},
	}
	p := buatPerjanjian(t, user, form)
	termin, _ := json.Marshal(ambilPerjanjian(t, user, "", p.ID).FormData["termin"])
	if string(termin) != `[{"nilai":300000,"persen":30,"syarat":"setelah penandatanganan"},{"nilai":300000,"persen":30,"syarat":"setelah laporan"},{"nilai":400000,"persen":40,"syarat":"setelah serah terima"}]` {
		t.Errorf("termin tersimpan %s", termin)
	}

	for _, c := range []struct {
		nama   string
		termin any
		pesan  string
	}{
		{"MelebihiSeratus", []any{map[string]any{"persen": 70, "syarat": "a"}, map[string]any{"persen": 40, "syarat": "b"}, map[string]any{"syarat": "c"}}, "Termin Pembayaran: jumlah persen termin melebihi 100%"},
		{"TanpaSisa", []any{map[string]any{"persen": 100, "syarat": "a"}, map[string]any{"syarat": "b"}}, "Termin Pembayaran: persen sudah 100%, tidak ada sisa untuk termin terakhir"},
		{"TanpaTermin", []any{}, "Termin Pembayaran wajib diisi minimal satu baris"},
		{"TerminTidakDikirim", nil, "Termin Pembayaran wajib diisi minimal satu baris"},
	} {
		t.Run(c.nama, func(t *testing.T) {
			form := formPerjanjian()
			form["termin"] = c.termin
			w := panggil(t, user, http.MethodPost, urlPerjanjian(""), isiPengajuan(t, user, "PKS", form))
			harapKode(t, w, http.StatusBadRequest, c.nama)
			if !strings.Contains(w.Body.String(), c.pesan) {
				t.Errorf("pesan harus %q: %s", c.pesan, w.Body)
			}
		})
	}
}

// sesudahDeploy1 menyiapkan keadaan server sesudah Deploy 1: template bawaan PR 10
// belum terpasang, versi bawaan PR 8 aktif dengan field PR 9, field kerja PKS
// berisi kerja, dan satu pengajuan berjalan di versi PR 8. Semuanya dikembalikan
// setelah tes.
func sesudahDeploy1(t *testing.T, kerja string) (versiPR9 string, berjalan perjanjian, user string) {
	t.Helper()
	butuhMinIO(t)
	user = masukPeran(t, entity.RoleUser)
	checksum := checksumTemplateBawaan(t)
	aktif, asliKerja := templateAktif(t), fieldJenis(t, "PKS")
	var bawaan []string
	dbTes.Raw("SELECT id FROM agreement_templates WHERE code = 'PKS' AND checksum = ? AND deleted_at IS NULL", checksum).Scan(&bawaan)
	if len(bawaan) == 0 {
		t.Fatal("template bawaan belum terpasang di database tes")
	}
	t.Cleanup(func() {
		// Versi yang dipasang seed selama tes dihapus; template bawaan asli kembali.
		var baru []string
		dbTes.Raw("SELECT id FROM agreement_templates WHERE code = 'PKS' AND checksum = ?", checksum).Scan(&baru)
		for _, id := range append(baru, versiPR9) {
			hapusVersiUji(id)
		}
		dbTes.Exec("UPDATE agreement_templates SET checksum = ? WHERE id IN ?", checksum, bawaan)
		dbTes.Exec("UPDATE agreement_templates SET status = 'ARCHIVED' WHERE code = 'PKS' AND status = 'ACTIVE'")
		dbTes.Exec("UPDATE agreement_templates SET status = 'ACTIVE' WHERE id = ?", aktif.ID)
		dbTes.Exec("UPDATE agreement_types SET fields = CAST(? AS jsonb) WHERE code = 'PKS'", asliKerja)
	})
	eksekusi(t, "UPDATE agreement_templates SET checksum = 'bawaan-pr10-uji' WHERE id IN ?", bawaan)
	eksekusi(t, "UPDATE agreement_templates SET status = 'ARCHIVED' WHERE code = 'PKS' AND status = 'ACTIVE'")
	versiPR9 = pasangVersiPR9(t, string(entity.TemplateStatusActive), fieldSesudahDeploy1(t))
	eksekusi(t, "UPDATE agreement_types SET fields = CAST(? AS jsonb) WHERE code = 'PKS'", kerja)
	berjalan = buatPerjanjian(t, user, formPerjanjianPR9())
	return versiPR9, berjalan, user
}

func isiPengajuanDB(t *testing.T, id string) string {
	t.Helper()
	var isi string
	if err := dbTes.Raw("SELECT form_data::text FROM agreement_documents WHERE id = ?", id).Scan(&isi).Error; err != nil {
		t.Fatal(err)
	}
	return isi
}

// Deploy 2 (PR 10) di atas server sesudah Deploy 1: seed memasang template bawaan
// termin berulang sekali saja, dengan definisi field-nya sendiri, dan
// mengaktifkannya. Field kerja yang belum diubah Admin ikut diganti. Pengajuan
// yang berjalan di versi PR 8 tetap di versinya dengan isian yang sama, dan
// tetap bisa dipratinjau dan disetujui (keputusan D1).
func TestMigrasiTermin(t *testing.T) {
	versiPR9, berjalan, user := sesudahDeploy1(t, fieldSesudahDeploy1(t))
	isiSebelum := isiPengajuanDB(t, berjalan.ID)

	jalankanSeedTemplate(t)

	bawaan := templateBawaanDB(t)
	if len(bawaan) != 1 || bawaan[0].Status != string(entity.TemplateStatusActive) {
		t.Fatalf("template bawaan setelah seed: %+v, harus terpasang sekali dan aktif", bawaan)
	}
	field := fieldTemplate(t, bawaan[0].ID)
	if !strings.Contains(field, `"type": "group"`) || strings.Contains(field, "termin_1_persen") {
		t.Errorf("template bawaan baru tidak memakai grup termin: %s", field)
	}
	if kerja := fieldJenis(t, "PKS"); kerja != field {
		t.Errorf("field kerja yang belum diubah Admin tidak diganti ke field bawaan baru:\n%s", kerja)
	}
	if v := templateDB(t, versiPR9); v.Status != string(entity.TemplateStatusArchived) {
		t.Errorf("versi PR 8 berstatus %s, harus ARCHIVED", v.Status)
	}
	// Deploy 2 juga memuat PR 11: versi PR 8 hanya mendapat tanda nilai utama (keputusan P1).
	if f := fieldTemplate(t, versiPR9); !strings.Contains(f, "termin_1_persen") || !samaJSON(t, f, fieldPR9(t)) {
		t.Errorf("field versi PR 8 berubah selain tanda nilai utama: %s", f)
	}
	if id, _ := templatePengajuan(t, berjalan.ID); id != versiPR9 {
		t.Errorf("pengajuan berjalan pindah ke %s, harus tetap di versi PR 8", id)
	}
	if isi := isiPengajuanDB(t, berjalan.ID); isi != isiSebelum {
		t.Errorf("isian pengajuan berjalan berubah:\n%s\nsebelumnya:\n%s", isi, isiSebelum)
	}
	if p := buatPerjanjian(t, user, formPerjanjian()); p.TemplateID != bawaan[0].ID {
		t.Errorf("pengajuan baru memakai template %s, harus template bawaan baru", p.TemplateID)
	}

	t.Run("PengajuanLamaDisetujui", func(t *testing.T) {
		butuhLibreOffice(t)
		legal := masukPeran(t, entity.RoleLegal)
		w := panggil(t, user, http.MethodGet, urlPerjanjian("", "/", berjalan.ID, "/preview"), nil)
		harapKode(t, w, http.StatusOK, "pratinjau pengajuan di versi PR 8")
		if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
			t.Error("pratinjau bukan PDF")
		}
		statusPerjanjian(t, legal, berjalan.ID, "UNDER_REVIEW", "", http.StatusOK)
		harapKode(t, panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", berjalan.ID, "/meta"), map[string]string{"signing_date": "2026-10-11"}), http.StatusOK, "mengisi tanggal tanda tangan")
		statusPerjanjian(t, legal, berjalan.ID, "COMPLETED", "", http.StatusOK)
	})
}

// Field kerja yang sudah diubah Admin tidak ditimpa saat template bawaan baru
// dipasang. Template bawaan baru terpasang sebagai Draft dengan definisinya
// sendiri, dan versi yang aktif tetap aktif (keputusan D3).
func TestIsianKerjaSaatBawaanBaru(t *testing.T) {
	kerja := strings.Replace(fieldSesudahDeploy1(t), `"label":"Nama Perusahaan"`, `"label":"Nama Mitra"`, 1)
	if kerja == fieldSesudahDeploy1(t) {
		t.Fatal("label uji tidak ditemukan di field PR 9")
	}
	versiPR9, _, _ := sesudahDeploy1(t, kerja)

	jalankanSeedTemplate(t)

	bawaan := templateBawaanDB(t)
	if len(bawaan) != 1 || bawaan[0].Status != string(entity.TemplateStatusDraft) {
		t.Fatalf("template bawaan setelah seed: %+v, harus terpasang sekali sebagai Draft", bawaan)
	}
	if f := fieldTemplate(t, bawaan[0].ID); !strings.Contains(f, `"type": "group"`) {
		t.Errorf("template bawaan baru tidak membawa definisinya sendiri: %s", f)
	}
	if s := templateDB(t, versiPR9).Status; s != string(entity.TemplateStatusActive) {
		t.Errorf("versi aktif berstatus %s setelah seed, harus tetap ACTIVE", s)
	}
	// Isian kerja Admin tidak ditimpa; hanya tanda nilai utama PR 11 yang ditambahkan.
	harap := strings.Replace(fieldPR9(t), `"label":"Nama Perusahaan"`, `"label":"Nama Mitra"`, 1)
	if k := fieldJenis(t, "PKS"); !samaJSON(t, k, harap) {
		t.Errorf("field kerja yang diubah Admin tertimpa:\n%s", k)
	}
}
