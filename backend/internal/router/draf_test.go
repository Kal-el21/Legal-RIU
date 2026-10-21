package router

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"

	"github.com/google/uuid"
)

// modulDraf: satu fitur pengajuan yang punya draf (PR Draf).
type modulDraf struct {
	nama, tabel string
	url         func(awalan string, sisa ...string) string
	buat        func(t *testing.T, token string) string // membuat pengajuan biasa, mengembalikan ID
	butuhMinIO  bool
	statusLegal map[string]string // isi PATCH /status yang sah untuk pengajuan bukan draf
}

func semuaModulDraf() []modulDraf {
	daftar := make([]modulDraf, 0, 3)
	for _, m := range semuaModulPengajuan {
		daftar = append(daftar, modulDraf{
			nama: m.nama, tabel: m.tabel, url: m.url,
			buat:        func(t *testing.T, token string) string { return m.buat(t, token).ID },
			statusLegal: map[string]string{"status": "UNDER_REVIEW"},
		})
	}
	return append(daftar, modulDraf{
		nama: "Perjanjian", tabel: "agreement_documents", url: urlPerjanjian,
		buat:        func(t *testing.T, token string) string { return buatPerjanjian(t, token, formPerjanjian()).ID },
		butuhMinIO:  true,
		statusLegal: map[string]string{"status": "UNDER_REVIEW"},
	})
}

// jadikanDraf mengubah pengajuan yang baru dibuat menjadi draf langsung di database, sama
// dengan hasil Simpan draf: status DRAFT dan nomor sementara DRAF-<uuid> (keputusan K1).
func jadikanDraf(t *testing.T, tabel, id string) {
	t.Helper()
	if err := dbTes.Exec("UPDATE "+tabel+" SET status = ?, ticket_number = ? WHERE id = ?",
		entity.StatusDraft, "DRAF-"+uuid.NewString(), id).Error; err != nil {
		t.Fatalf("menjadikan draf: %v", err)
	}
}

// memuat: jawaban berhasil dan isinya menyebut ID itu.
func memuat(t *testing.T, token, alamat, id string) bool {
	t.Helper()
	w := panggil(t, token, http.MethodGet, alamat, nil)
	harapKode(t, w, http.StatusOK, "mengambil "+alamat)
	return strings.Contains(w.Body.String(), id)
}

// Draf hanya terlihat pemiliknya: tidak di daftar, detail, dashboard, laporan, maupun
// lonceng milik LEGAL, ADMIN, atau USER lain, dan alamat approver menjawabnya tidak ditemukan.
func TestDrafTersembunyi(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)

	for _, m := range semuaModulDraf() {
		t.Run(m.nama, func(t *testing.T) {
			if m.butuhMinIO {
				butuhMinIO(t)
			}
			id := m.buat(t, pemilik)
			jadikanDraf(t, m.tabel, id)

			// Pemilik: tidak di daftar biasa, ada di daftar draf, detail terbuka.
			if memuat(t, pemilik, m.url("", "?limit=100"), id) {
				t.Error("draf muncul di daftar biasa pemiliknya")
			}
			if !memuat(t, pemilik, m.url("", "?limit=100&status=DRAFT"), id) {
				t.Error("draf tidak muncul di daftar draf pemiliknya")
			}
			harapKode(t, panggil(t, pemilik, http.MethodGet, m.url("", "/", id), nil), http.StatusOK, "pemilik membuka drafnya")

			// Orang lain, termasuk yang boleh melihat semua pengajuan.
			for _, k := range []struct{ nama, token, awalan string }{{"USER lain", lain, ""}, {"LEGAL", legal, "/legal"}, {"ADMIN", admin, "/admin"}} {
				for _, saringan := range []string{"", "&status=DRAFT"} {
					if memuat(t, k.token, m.url(k.awalan, "?limit=100"+saringan), id) {
						t.Errorf("%s melihat draf di daftar%s", k.nama, saringan)
					}
				}
				harapKode(t, panggil(t, k.token, http.MethodGet, m.url(k.awalan, "/", id), nil), http.StatusNotFound, k.nama+" membuka draf")
			}
			w := panggil(t, legal, http.MethodPatch, m.url("/legal", "/", id, "/status"), m.statusLegal)
			if w.Code == http.StatusOK {
				t.Error("LEGAL bisa mengubah status draf")
			}
			if s := statusDi(t, m.tabel, id); s != string(entity.StatusDraft) {
				t.Errorf("status draf berubah menjadi %s", s)
			}
			if p := ambilPengingat(t, legal, "").cari(id); p != nil {
				t.Errorf("draf muncul di lonceng LEGAL: %+v", p)
			}
		})
	}

	t.Run("PDFDanUnggahHasil", func(t *testing.T) {
		m := semuaModulPengajuan[0]
		id := m.buat(t, pemilik).ID
		jadikanDraf(t, m.tabel, id)
		harapKode(t, panggil(t, legal, http.MethodGet, m.url("/legal", "/", id, "/pdf"), nil), http.StatusNotFound, "LEGAL meminta PDF draf")
		w := panggilForm(t, legal, http.MethodPost, m.url("/legal", "/", id, "/result"), map[string]string{"notes": "x"}, berkas{"result", "hasil.pdf", []byte("hasil")})
		if w.Code == http.StatusOK {
			t.Error("LEGAL bisa mengunggah hasil ke draf")
		}
	})

	t.Run("PerjanjianMetaDanVersiTemplate", func(t *testing.T) {
		butuhMinIO(t)
		p := buatPerjanjian(t, pemilik, formPerjanjian())
		berjalan := func() float64 {
			t.Helper()
			w := panggil(t, legal, http.MethodGet, "/api/v1/legal/agreement-templates", nil)
			harapKode(t, w, http.StatusOK, "daftar template")
			var daftar []struct {
				ID               string  `json:"id"`
				RunningDocuments float64 `json:"running_documents"`
			}
			isiData(t, w, &daftar)
			for _, v := range daftar {
				if v.ID == p.TemplateID {
					return v.RunningDocuments
				}
			}
			t.Fatalf("versi %s tidak ada di daftar template", p.TemplateID)
			return 0
		}
		sebelum := berjalan()
		jadikanDraf(t, "agreement_documents", p.ID)
		if got := berjalan(); got != sebelum-1 {
			t.Errorf("pengajuan berjalan di versi template: dapat %v, harap %v (draf tidak dihitung)", got, sebelum-1)
		}
		w := panggil(t, legal, http.MethodPatch, urlPerjanjian("/legal", "/", p.ID, "/meta"), map[string]string{"signing_date": "2026-10-11"})
		if w.Code == http.StatusOK {
			t.Error("LEGAL bisa mengubah data persetujuan draf")
		}
	})
}

// Dashboard dan laporan tidak menghitung draf.
func TestDrafTidakDihitung(t *testing.T) {
	const surel = "tes-draf-dashboard@example.com"
	userKhusus(t, surel)
	pemilik := masuk(t, surel)
	legal := masukPeran(t, entity.RoleLegal)
	statistik := map[string]struct{ token, alamat string }{
		"pemilik": {pemilik, "/api/v1/dashboard/stats"},
		"legal":   {legal, "/api/v1/legal/dashboard/stats"},
	}
	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) { drafTidakDihitung(t, m, pemilik, legal, statistik) })
	}
}

func drafTidakDihitung(t *testing.T, m modulPengajuan, pemilik, legal string, statistik map[string]struct{ token, alamat string }) {
	kolom := map[string]string{"LegalOpinion": "total_legal_opinions", "ReviewDocument": "total_document_reviews"}[m.nama]
	laporan := "/api/v1/legal/reports/" + map[string]string{"LegalOpinion": "legal-opinions", "ReviewDocument": "document-reviews"}[m.nama] + "?group_by=status"
	ambilLaporan := func() float64 {
		t.Helper()
		w := panggil(t, legal, http.MethodGet, laporan, nil)
		harapKode(t, w, http.StatusOK, "laporan")
		var g grafik
		isiData(t, w, &g)
		return g.jumlah()
	}
	sebelum := map[string]float64{}
	for nama, s := range statistik {
		sebelum[nama] = ambilAngka(t, s.token, s.alamat, kolom)
	}
	laporanSebelum := ambilLaporan()

	id := m.buat(t, pemilik).ID
	jadikanDraf(t, m.tabel, id)

	for nama, s := range statistik {
		if got := ambilAngka(t, s.token, s.alamat, kolom); got != sebelum[nama] {
			t.Errorf("%s untuk %s setelah ada draf: dapat %v, harap %v", kolom, nama, got, sebelum[nama])
		}
	}
	if got := ambilLaporan(); got != laporanSebelum {
		t.Errorf("laporan %s setelah ada draf: dapat %v, harap %v", m.nama, got, laporanSebelum)
	}
	for _, s := range []struct{ token, alamat string }{{pemilik, "/api/v1/dashboard/recent"}, {legal, "/api/v1/legal/dashboard/recent"}} {
		if memuat(t, s.token, s.alamat, id) {
			t.Errorf("draf muncul di aktivitas terbaru %s", s.alamat)
		}
	}
}

// Repositori Dokumen hanya memuat file milik data yang belum dihapus dan bukan draf
// (temuan Perencanaan PR Draf nomor 2). Baris lampiran dibuat langsung di tabelnya, jadi
// tes ini tidak butuh MinIO; backfill hanya membaca tabel.
func TestRepositoriTanpaPengajuanTerhapus(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	uploader := idPemilik(t, "legal_opinions", semuaModulPengajuan[0].buat(t, pemilik).ID)

	type kasus struct{ nama, tabel, id string }
	var daftar []kasus
	for _, m := range semuaModulPengajuan {
		for _, keadaan := range []string{"biasa", "draf", "dihapus"} {
			id := m.buat(t, pemilik).ID
			switch keadaan {
			case "draf":
				jadikanDraf(t, m.tabel, id)
			case "dihapus":
				harapKode(t, panggil(t, pemilik, http.MethodDelete, m.url("", "/", id), nil), http.StatusOK, "hapus pengajuan")
			}
			daftar = append(daftar, kasus{m.nama + " " + keadaan, m.tabel, id})
		}
	}
	if !tanpaMinIO { // membuat perjanjian butuh template di MinIO
		for _, keadaan := range []string{"biasa", "draf", "dihapus"} {
			id := buatPerjanjian(t, pemilik, formPerjanjian()).ID
			switch keadaan {
			case "draf":
				jadikanDraf(t, "agreement_documents", id)
			case "dihapus":
				harapKode(t, panggil(t, pemilik, http.MethodDelete, urlPerjanjian("", "/", id), nil), http.StatusOK, "hapus perjanjian")
			}
			daftar = append(daftar, kasus{"Perjanjian " + keadaan, "agreement_documents", id})
		}
	}
	perk := buatPerkara(t, legal)
	daftar = append(daftar, kasus{"perkara dihapus", "legal_cases", perk.ID})

	// Satu lampiran dan satu hasil per pengajuan, dan satu dokumen untuk perkara.
	path := map[string][]string{}
	for _, k := range daftar {
		uid := uuid.MustParse(k.id)
		p := "uji-repo/" + k.id
		var err error
		switch k.tabel {
		case "legal_opinions":
			err = dbTes.Create(&entity.LegalOpinionAttachment{LegalOpinionID: uid, FileName: "a.pdf", FilePath: p + "/a.pdf", UploadRound: 1}).Error
			if err == nil {
				err = dbTes.Create(&entity.LegalOpinionResult{LegalOpinionID: uid, UploadedBy: uploader, FileName: "h.pdf", FilePath: p + "/h.pdf"}).Error
			}
			path[k.nama] = []string{p + "/a.pdf", p + "/h.pdf"}
		case "document_reviews":
			err = dbTes.Create(&entity.DocumentReviewAttachment{DocumentReviewID: uid, FileName: "a.pdf", FilePath: p + "/a.pdf", UploadRound: 1}).Error
			if err == nil {
				err = dbTes.Create(&entity.DocumentReviewResult{DocumentReviewID: uid, UploadedBy: uploader, FileName: "h.pdf", FilePath: p + "/h.pdf"}).Error
			}
			path[k.nama] = []string{p + "/a.pdf", p + "/h.pdf"}
		case "agreement_documents":
			err = dbTes.Create(&entity.AgreementAttachment{AgreementDocumentID: uid, FileName: "a.pdf", FilePath: p + "/a.pdf", UploadRound: 1, UploadedBy: uploader}).Error
			if err == nil { // file akhir: hanya perjanjian Selesai yang punya, tetapi backfill tidak melihat status
				err = dbTes.Exec("UPDATE agreement_documents SET generated_pdf_path = ? WHERE id = ?", p+"/final.pdf", k.id).Error
			}
			path[k.nama] = []string{p + "/a.pdf"}
			if !strings.HasSuffix(k.nama, "draf") {
				path[k.nama] = append(path[k.nama], p+"/final.pdf")
			}
		case "legal_cases":
			err = dbTes.Exec("UPDATE legal_cases SET document_link = ?, photo = ? WHERE id = ?", p+"/d.pdf", p+"/f.png", k.id).Error
			if err == nil {
				err = dbTes.Create(&entity.CaseChronology{CaseID: uid, AgendaDate: time.Now(), Agenda: "uji", Documents: `["` + p + `/k.pdf"]`}).Error
			}
			if err == nil {
				err = dbTes.Exec("UPDATE legal_cases SET deleted_at = now() WHERE id = ?", k.id).Error
			}
			path[k.nama] = []string{p + "/d.pdf", p + "/f.png", p + "/k.pdf"}
		}
		if err != nil {
			t.Fatalf("menyiapkan file %s: %v", k.nama, err)
		}
	}

	if err := seed.BackfillRepositoryDocuments(dbTes); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	for _, k := range daftar {
		harapAktif := strings.HasSuffix(k.nama, "biasa")
		for _, p := range path[k.nama] {
			var aktif int64
			dbTes.Raw("SELECT count(*) FROM repository_documents WHERE file_path = ? AND is_active", p).Scan(&aktif)
			if (aktif > 0) != harapAktif {
				t.Errorf("%s: %s aktif di Repositori Dokumen = %v, harap %v", k.nama, p, aktif > 0, harapAktif)
			}
		}
	}
}

// Repositori Dokumen tidak tersalin ulang setiap server menyala, nama asli perkara tidak
// ditimpa nama objek, dan salinan lama dibersihkan (PR Draf tahap D.1b, keputusan K7).
func TestRepositoriTidakKembar(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	jumlah := func(path string) (n int64) {
		dbTes.Raw("SELECT count(*) FROM repository_documents WHERE file_path = ?", path).Scan(&n)
		return n
	}
	namaDi := func(path string) (nama string) {
		dbTes.Raw("SELECT file_name FROM repository_documents WHERE file_path = ?", path).Scan(&nama)
		return nama
	}
	backfill := func() {
		t.Helper()
		if err := seed.BackfillRepositoryDocuments(dbTes); err != nil {
			t.Fatalf("backfill: %v", err)
		}
	}

	// Lampiran Legal Opinion dan dokumen perkara yang nama aslinya tersimpan saat unggah.
	lo := semuaModulPengajuan[0].buat(t, pemilik).ID
	lampiran := "uji-kembar/" + lo + "/a.pdf"
	if err := dbTes.Create(&entity.LegalOpinionAttachment{LegalOpinionID: uuid.MustParse(lo), FileName: "Surat Asli.pdf", FilePath: lampiran, UploadRound: 1}).Error; err != nil {
		t.Fatal(err)
	}
	perk := buatPerkara(t, legal)
	dokumen := "legal-cases/documents/case-" + perk.ID + "-uji.pdf"
	if err := dbTes.Exec("UPDATE legal_cases SET document_link = ? WHERE id = ?", dokumen, perk.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := dbTes.Create(&entity.RepositoryDocument{Title: "Gugatan Asli.pdf", FeatureCode: "case_management", SourceType: "legal_cases",
		SourceID: perk.ID, FilePath: dokumen, FileName: "Gugatan Asli.pdf", IsActive: true}).Error; err != nil {
		t.Fatal(err)
	}

	backfill()
	backfill()
	for _, p := range []string{lampiran, dokumen} {
		if n := jumlah(p); n != 1 {
			t.Errorf("%s tercatat %d kali sesudah dua kali backfill, harap 1", p, n)
		}
	}
	if got := namaDi(dokumen); got != "Gugatan Asli.pdf" {
		t.Errorf("nama dokumen perkara sesudah backfill: %q, harap nama aslinya", got)
	}

	t.Run("SalinanLamaDibersihkan", func(t *testing.T) {
		var indeks int64
		dbTes.Raw("SELECT count(*) FROM pg_indexes WHERE indexname = 'idx_repository_documents_sumber'").Scan(&indeks)
		if indeks != 1 {
			t.Fatal("indeks unik Repositori Dokumen tidak terpasang saat server menyala")
		}
		// Keadaan sebelum D.1b: tanpa indeks unik, dengan salinan bernama objek di kiri-kanan nama asli.
		if err := dbTes.Exec("DROP INDEX IF EXISTS idx_repository_documents_sumber").Error; err != nil {
			t.Fatal(err)
		}
		salinan := func(nama string) entity.RepositoryDocument {
			return entity.RepositoryDocument{Title: nama, FeatureCode: "case_management", SourceType: "legal_cases",
				SourceID: perk.ID, FilePath: dokumen, FileName: nama}
		}
		for _, nama := range []string{"case-" + perk.ID + "-uji.pdf", "case-" + perk.ID + "-uji.pdf"} {
			if err := dbTes.Create(ptr(salinan(nama))).Error; err != nil {
				t.Fatal(err)
			}
		}
		// Salinan bernama objek dibuat paling dulu, supaya yang disisakan dipilih karena namanya, bukan umurnya.
		if err := dbTes.Exec("UPDATE repository_documents SET created_at = created_at - interval '1 day' WHERE file_path = ? AND file_name <> 'Gugatan Asli.pdf'", dokumen).Error; err != nil {
			t.Fatal(err)
		}
		if err := seed.RapikanRepositoriKembar(dbTes); err != nil {
			t.Fatalf("merapikan: %v", err)
		}
		if n := jumlah(dokumen); n != 1 {
			t.Errorf("sesudah dirapikan %s tercatat %d kali, harap 1", dokumen, n)
		}
		if got := namaDi(dokumen); got != "Gugatan Asli.pdf" {
			t.Errorf("baris yang disisakan bernama %q, harap yang bernama asli", got)
		}
		if err := dbTes.Create(ptr(salinan("lagi.pdf"))).Error; err == nil {
			t.Error("salinan baru diterima; indeks unik tidak terpasang")
		}
	})
}

func ptr[T any](v T) *T { return &v }

func statusDi(t *testing.T, tabel, id string) string {
	t.Helper()
	var s string
	dbTes.Raw("SELECT status FROM "+tabel+" WHERE id = ?", id).Scan(&s)
	return s
}

func idPemilik(t *testing.T, tabel, id string) uuid.UUID {
	t.Helper()
	var s string
	dbTes.Raw("SELECT user_id::text FROM "+tabel+" WHERE id = ?", id).Scan(&s)
	u, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("pemilik %s: %v", id, err)
	}
	return u
}
