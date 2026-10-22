package router

// Alur Legal Opinion dan Review Document. Kedua modul berbentuk sama (pengajuan
// oleh USER, diproses oleh LEGAL atau ADMIN), jadi tiap tes dijalankan untuk
// keduanya.

import (
	"bytes"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"

	"legal-riu-portal/internal/entity"
)

type modulPengajuan struct {
	nama       string
	tabel      string
	alamat     string            // tanpa awalan peran, mis. "/legal-opinions"
	kolomJudul string            // kolom yang diubah di tes ubah
	kolom      map[string]string // isi form yang sah
	pdfPemilik bool              // USER punya route PDF sendiri (hanya Legal Opinion)
	namaPDF    string            // bagian nama file PDF ringkasan sesudah nomor tiket
}

var semuaModulPengajuan = []modulPengajuan{
	{
		nama: "LegalOpinion", tabel: "legal_opinions", alamat: "/legal-opinions", kolomJudul: "title", pdfPemilik: true, namaPDF: "Legal-Opinion",
		kolom: map[string]string{
			"requestor_name": "Budi", "requestor_position": "Staff", "requestor_division": "Divisi Tes",
			"requestor_email": "budi@example.com", "requestor_phone": "08123456789",
			"legal_type": "Kontrak", "title": "Judul awal",
			"chronology": "Kronologi singkat", "question": "Apa pendapat hukumnya?",
		},
	},
	{
		nama: "ReviewDocument", tabel: "document_reviews", alamat: "/review-documents", kolomJudul: "document_name", namaPDF: "Review-Document",
		kolom: map[string]string{
			"requestor_name": "Budi", "requestor_position": "Staff", "requestor_division": "Divisi Tes",
			"requestor_email": "budi@example.com", "requestor_phone": "08123456789",
			"document_name": "Judul awal", "second_party": "PT Pihak Kedua", "document_type": "Perjanjian",
		},
	},
}

// pengajuan memuat kolom yang diperiksa tes dari kedua modul.
type pengajuan struct {
	ID           string `json:"id"`
	TicketNumber string `json:"ticket_number"`
	Status       string `json:"status"`
	AdminNote    string `json:"admin_note"`
	Title        string `json:"title"`
	DocumentName string `json:"document_name"`
	Attachments  []struct {
		ID          string `json:"id"`
		FileName    string `json:"file_name"`
		FilePath    string `json:"file_path"`
		UploadRound int    `json:"upload_round"`
	} `json:"attachments"`
	Results []struct {
		FilePath string `json:"file_path"`
		Notes    string `json:"notes"`
	} `json:"results"`
}

func (p pengajuan) judul() string { return p.Title + p.DocumentName }

// Alamat modul untuk tiap peran: USER tanpa awalan, lainnya "/legal" atau "/admin".
func (m modulPengajuan) url(awalan string, sisa ...string) string {
	return "/api/v1" + awalan + m.alamat + strings.Join(sisa, "")
}

func (m modulPengajuan) buat(t *testing.T, token string, files ...berkas) pengajuan {
	t.Helper()
	w := panggilForm(t, token, http.MethodPost, m.url(""), m.kolom, files...)
	harapKode(t, w, http.StatusCreated, "membuat pengajuan")
	var p pengajuan
	isiData(t, w, &p)
	return p
}

func (m modulPengajuan) ambil(t *testing.T, token, awalan, id string) pengajuan {
	t.Helper()
	w := panggil(t, token, http.MethodGet, m.url(awalan, "/", id), nil)
	harapKode(t, w, http.StatusOK, "mengambil pengajuan")
	var p pengajuan
	isiData(t, w, &p)
	return p
}

func (m modulPengajuan) ubahStatus(t *testing.T, token, awalan, id, status, catatan string, harap int) {
	t.Helper()
	w := panggil(t, token, http.MethodPatch, m.url(awalan, "/", id, "/status"),
		map[string]string{"status": status, "admin_note": catatan})
	harapKode(t, w, harap, "mengubah status ke "+status)
}

func (m modulPengajuan) adaDiDaftar(t *testing.T, token, awalan, id string) bool {
	t.Helper()
	w := panggil(t, token, http.MethodGet, m.url(awalan, "?limit=100"), nil)
	harapKode(t, w, http.StatusOK, "mengambil daftar")
	var daftar struct {
		Items []pengajuan `json:"items"`
	}
	isiData(t, w, &daftar)
	for _, p := range daftar.Items {
		if p.ID == id {
			return true
		}
	}
	return false
}

// unduh mengambil file dan memeriksa nama unduhannya.
func (m modulPengajuan) unduh(t *testing.T, token, path, nama string) []byte {
	t.Helper()
	w := panggil(t, token, http.MethodGet, m.url("", "/download?path=", url.QueryEscape(path)), nil)
	harapKode(t, w, http.StatusOK, "mengunduh "+path)
	harapNamaUnduhan(t, w, nama, "mengunduh "+path)
	return w.Body.Bytes()
}

// Alur tanpa file: buat, kepemilikan, ubah, mesin status, ajukan ulang, PDF, hapus.
func TestAlurPengajuan(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)

	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			p := m.buat(t, pemilik)
			if p.Status != "SUBMITTED" || p.TicketNumber == "" {
				t.Fatalf("pengajuan baru: status %q, tiket %q", p.Status, p.TicketNumber)
			}

			// Kepemilikan: USER lain tidak bisa melihat, mengubah, atau menghapus.
			if !m.adaDiDaftar(t, pemilik, "", p.ID) {
				t.Error("pengajuan tidak muncul di daftar pemiliknya")
			}
			if m.adaDiDaftar(t, lain, "", p.ID) {
				t.Error("pengajuan muncul di daftar USER lain")
			}
			harapKode(t, panggil(t, lain, http.MethodGet, m.url("", "/", p.ID), nil), http.StatusNotFound, "USER lain melihat")
			harapKode(t, panggil(t, lain, http.MethodPut, m.url("", "/", p.ID), m.kolom), http.StatusBadRequest, "USER lain mengubah")
			harapKode(t, panggil(t, lain, http.MethodDelete, m.url("", "/", p.ID), nil), http.StatusBadRequest, "USER lain menghapus")

			// Pemilik mengubah selagi SUBMITTED.
			baru := map[string]string{}
			for k, v := range m.kolom {
				baru[k] = v
			}
			baru[m.kolomJudul] = "Judul diubah"
			harapKode(t, panggil(t, pemilik, http.MethodPut, m.url("", "/", p.ID), baru), http.StatusOK, "pemilik mengubah")
			if j := m.ambil(t, pemilik, "", p.ID).judul(); j != "Judul diubah" {
				t.Errorf("judul setelah diubah: %q", j)
			}
			salah := map[string]string{}
			for k, v := range m.kolom {
				salah[k] = v
			}
			salah["requestor_email"] = "bukan-surel"
			harapKode(t, panggil(t, pemilik, http.MethodPut, m.url("", "/", p.ID), salah), http.StatusBadRequest, "surel tidak valid")

			// LEGAL dan ADMIN melihat semua pengajuan.
			if !m.adaDiDaftar(t, legal, "/legal", p.ID) {
				t.Error("pengajuan tidak muncul di daftar LEGAL")
			}
			m.ambil(t, legal, "/legal", p.ID)
			m.ambil(t, admin, "/admin", p.ID)

			// Mesin status: lompatan yang tidak sah ditolak dan status tidak berubah.
			m.ubahStatus(t, legal, "/legal", p.ID, "COMPLETED", "", http.StatusBadRequest)
			if s := m.ambil(t, pemilik, "", p.ID).Status; s != "SUBMITTED" {
				t.Fatalf("status setelah lompatan ditolak: %s", s)
			}

			// Setelah diproses, pemilik tidak bisa lagi mengubah, menghapus, atau mengajukan ulang.
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			harapKode(t, panggil(t, pemilik, http.MethodPut, m.url("", "/", p.ID), baru), http.StatusBadRequest, "ubah saat UNDER_REVIEW")
			harapKode(t, panggil(t, pemilik, http.MethodDelete, m.url("", "/", p.ID), nil), http.StatusBadRequest, "hapus saat UNDER_REVIEW")
			harapKode(t, panggil(t, pemilik, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil), http.StatusBadRequest, "ajukan ulang saat UNDER_REVIEW")

			// Minta revisi: catatan sampai ke pemilik, lalu pemilik mengajukan ulang.
			m.ubahStatus(t, legal, "/legal", p.ID, "NEED_REVISION", "lengkapi data", http.StatusOK)
			if got := m.ambil(t, pemilik, "", p.ID); got.Status != "NEED_REVISION" || got.AdminNote != "lengkapi data" {
				t.Errorf("setelah minta revisi: status %q, catatan %q", got.Status, got.AdminNote)
			}
			harapKode(t, panggil(t, lain, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil), http.StatusBadRequest, "USER lain mengajukan ulang")
			harapKode(t, panggil(t, pemilik, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil), http.StatusOK, "ajukan ulang")
			if s := m.ambil(t, pemilik, "", p.ID).Status; s != "RESUBMITTED" {
				t.Errorf("status setelah ajukan ulang: %s", s)
			}

			// ADMIN menolak, pemilik mengajukan ulang lagi, lalu selesai.
			m.ubahStatus(t, admin, "/admin", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			m.ubahStatus(t, admin, "/admin", p.ID, "REJECTED", "tidak sesuai", http.StatusOK)
			harapKode(t, panggil(t, pemilik, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil), http.StatusOK, "ajukan ulang setelah ditolak")
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			m.ubahStatus(t, legal, "/legal", p.ID, "COMPLETED", "", http.StatusOK)
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusBadRequest) // COMPLETED adalah akhir

			// PDF untuk LEGAL, dan untuk pemilik di modul yang menyediakannya.
			pemintaPDF := []struct{ token, awalan string }{{legal, "/legal"}}
			if m.pdfPemilik {
				pemintaPDF = append(pemintaPDF, struct{ token, awalan string }{pemilik, ""})
			}
			for _, k := range pemintaPDF {
				w := panggil(t, k.token, http.MethodGet, m.url(k.awalan, "/", p.ID, "/pdf"), nil)
				harapKode(t, w, http.StatusOK, "membuat PDF")
				if !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
					t.Errorf("jawaban PDF%s bukan PDF", k.awalan)
				}
				harapNamaUnduhan(t, w, p.TicketNumber+"_"+m.namaPDF+".pdf", "PDF"+k.awalan)
			}

			// Kolom wajib yang kosong ditolak.
			kurang := map[string]string{}
			for k, v := range m.kolom {
				kurang[k] = v
			}
			delete(kurang, m.kolomJudul)
			harapKode(t, panggilForm(t, pemilik, http.MethodPost, m.url(""), kurang), http.StatusBadRequest, "kolom wajib kosong")
		})
	}
}

// Pemilik bisa menghapus pengajuannya selagi SUBMITTED, dan penghapusan itu
// tidak mengganggu pengajuan berikutnya: nomor tiket yang sudah dipakai tidak
// boleh dipakai lagi.
func TestHapusPengajuan(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			p := m.buat(t, pemilik)

			harapKode(t, panggil(t, pemilik, http.MethodDelete, m.url("", "/", p.ID), nil), http.StatusOK, "pemilik menghapus")
			harapKode(t, panggil(t, pemilik, http.MethodGet, m.url("", "/", p.ID), nil), http.StatusNotFound, "melihat yang sudah dihapus")
			if m.adaDiDaftar(t, pemilik, "", p.ID) {
				t.Error("pengajuan yang dihapus masih muncul di daftar")
			}

			if lagi := m.buat(t, pemilik); lagi.TicketNumber == p.TicketNumber {
				t.Errorf("nomor tiket %s dipakai ulang setelah pengajuannya dihapus", lagi.TicketNumber)
			}
		})
	}
}

// Alur dengan file: lampiran, putaran unggah, hasil kajian, dan unduhan.
func TestBerkasPengajuan(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)

	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			isiA, isiB, isiHasil := []byte("lampiran pertama"), []byte("lampiran kedua"), []byte("hasil kajian")

			p := m.buat(t, pemilik, berkas{"attachments", "Surat Penawaran Résumé.pdf", isiA})
			if len(p.Attachments) != 1 || p.Attachments[0].UploadRound != 1 || p.Attachments[0].FileName != "Surat Penawaran Résumé.pdf" {
				t.Fatalf("lampiran setelah dibuat: %+v", p.Attachments)
			}
			if got := m.unduh(t, pemilik, p.Attachments[0].FilePath, p.TicketNumber+"_Lampiran_Surat Penawaran Résumé.pdf"); !bytes.Equal(got, isiA) {
				t.Errorf("isi lampiran yang diunduh berbeda: %q", got)
			}

			// Ajukan ulang dengan lampiran baru: lampiran lama tetap, yang baru putaran 2.
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			m.ubahStatus(t, legal, "/legal", p.ID, "NEED_REVISION", "tambah lampiran", http.StatusOK)
			w := panggilForm(t, pemilik, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil, berkas{"attachments", "revisi.pdf", isiB})
			harapKode(t, w, http.StatusOK, "ajukan ulang dengan lampiran")
			p = m.ambil(t, pemilik, "", p.ID)
			putaran := map[int]int{}
			for _, a := range p.Attachments {
				putaran[a.UploadRound]++
			}
			if len(p.Attachments) != 2 || putaran[1] != 1 || putaran[2] != 1 {
				t.Errorf("lampiran setelah ajukan ulang: %+v", p.Attachments)
			}

			// Unggah hasil saat UNDER_REVIEW menyelesaikan pengajuan.
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			harapKode(t, panggilForm(t, legal, http.MethodPost, m.url("/legal", "/", p.ID, "/result"), nil), http.StatusBadRequest, "unggah hasil tanpa file")
			w = panggilForm(t, legal, http.MethodPost, m.url("/legal", "/", p.ID, "/result"),
				map[string]string{"notes": "sudah dikaji"}, berkas{"result", "hasil.pdf", isiHasil})
			harapKode(t, w, http.StatusOK, "unggah hasil")
			p = m.ambil(t, pemilik, "", p.ID)
			if p.Status != "COMPLETED" || len(p.Results) != 1 || p.Results[0].Notes != "sudah dikaji" {
				t.Fatalf("setelah unggah hasil: status %q, hasil %+v", p.Status, p.Results)
			}
			if got := m.unduh(t, pemilik, p.Results[0].FilePath, p.TicketNumber+"_Hasil_hasil.pdf"); !bytes.Equal(got, isiHasil) {
				t.Errorf("isi hasil yang diunduh berbeda: %q", got)
			}
		})
	}
}

// Unduh lewat path hanya melayani file yang terdaftar sebagai lampiran atau
// hasil kajian pengajuan modul itu, dan hanya untuk yang boleh melihat
// pengajuannya. Selain itu dijawab 404, sama seperti file yang tidak ada.
func TestUnduhBerkasPengajuan(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)
	admin := masukPeran(t, entity.RoleAdmin)
	isi := []byte("lampiran untuk diunduh")

	// Satu lampiran per modul, semuanya milik USER yang sama.
	lampiran := map[string]string{}
	for _, m := range semuaModulPengajuan {
		lampiran[m.nama] = m.buat(t, pemilik, berkas{"attachments", "unduh.pdf", isi}).Attachments[0].FilePath
	}
	// File perkara: ada di penyimpanan, tetapi bukan milik modul pengajuan.
	perk := buatPerkara(t, legal)
	harapKode(t, panggilForm(t, legal, http.MethodPost, perkaraLegal+"/"+perk.ID+"/upload-document", nil, berkas{"document", "putusan.pdf", isi}), http.StatusOK, "unggah dokumen perkara")
	filePerkara := ambilPerkara(t, legal, perkaraLegal, perk.ID).DocumentLink

	for i, m := range semuaModulPengajuan {
		modulLain := semuaModulPengajuan[1-i]
		unduh := func(token, awalan, path string) (int, []byte) {
			t.Helper()
			w := panggil(t, token, http.MethodGet, m.url(awalan, "/download?path=", url.QueryEscape(path)), nil)
			return w.Code, w.Body.Bytes()
		}
		t.Run(m.nama, func(t *testing.T) {
			for _, k := range []struct{ nama, token, awalan string }{{"pemilik", pemilik, ""}, {"LEGAL", legal, "/legal"}, {"ADMIN", admin, "/admin"}} {
				kode, isiUnduhan := unduh(k.token, k.awalan, lampiran[m.nama])
				if kode != http.StatusOK || !bytes.Equal(isiUnduhan, isi) {
					t.Errorf("%s mengunduh lampiran: dapat %d %q", k.nama, kode, isiUnduhan)
				}
			}
			for _, k := range []struct{ nama, token, path string }{
				{"USER lain, lampiran milik orang", lain, lampiran[m.nama]},
				{"pemilik, lampiran milik modul " + modulLain.nama, pemilik, lampiran[modulLain.nama]},
				{"pemilik, file perkara", pemilik, filePerkara},
				{"pemilik, path yang tidak terdaftar", pemilik, "legal-opinions/attachments/tidak-ada.pdf"},
			} {
				if kode, _ := unduh(k.token, "", k.path); kode != http.StatusNotFound {
					t.Errorf("%s: dapat %d, harap 404", k.nama, kode)
				}
			}
		})
	}
}

// Lampiran diperiksa sebelum ada yang disimpan. Satu file yang ditolak dijawab
// 400 dengan alasannya, tanpa meninggalkan pengajuan baru maupun lampiran
// separuh jadi, saat membuat maupun saat mengajukan ulang.
func TestLampiranDitolak(t *testing.T) {
	butuhMinIO(t)
	pemilik := masukPeran(t, entity.RoleUser)
	legal := masukPeran(t, entity.RoleLegal)
	// File sah dikirim lebih dulu, supaya terlihat kalau ada yang telanjur disimpan.
	sah, ditolak := berkas{"attachments", "sah.pdf", []byte("isi")}, berkas{"attachments", "program.exe", []byte("x")}
	harapAlasan := func(t *testing.T, w *httptest.ResponseRecorder, konteks string) {
		t.Helper()
		harapKode(t, w, http.StatusBadRequest, konteks)
		if !strings.Contains(w.Body.String(), "tipe file tidak diizinkan") {
			t.Errorf("%s: alasan penolakan tidak disebut: %s", konteks, w.Body)
		}
	}

	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			var sebelum, sesudah int64
			dbTes.Table(m.tabel).Count(&sebelum)
			harapAlasan(t, panggilForm(t, pemilik, http.MethodPost, m.url(""), m.kolom, sah, ditolak), "membuat pengajuan dengan lampiran .exe")
			dbTes.Table(m.tabel).Count(&sesudah)
			if sesudah != sebelum {
				t.Errorf("pengajuan tetap tersimpan walau lampirannya ditolak (%d -> %d baris)", sebelum, sesudah)
			}

			p := m.buat(t, pemilik, sah)
			m.ubahStatus(t, legal, "/legal", p.ID, "UNDER_REVIEW", "", http.StatusOK)
			m.ubahStatus(t, legal, "/legal", p.ID, "NEED_REVISION", "perbaiki", http.StatusOK)
			harapAlasan(t, panggilForm(t, pemilik, http.MethodPost, m.url("", "/", p.ID, "/resubmit"), nil, sah, ditolak), "ajukan ulang dengan lampiran .exe")
			if q := m.ambil(t, pemilik, "", p.ID); q.Status != "NEED_REVISION" || len(q.Attachments) != 1 {
				t.Errorf("setelah ajukan ulang ditolak: status %q, %d lampiran; harap NEED_REVISION dan 1 lampiran", q.Status, len(q.Attachments))
			}
		})
	}
}

// Harapan yang benar tetapi belum dipenuhi kode. Tiap subtes dilewati lewat
// bugDiketahui sampai perbaikannya masuk. Untuk melihat semuanya gagal apa
// adanya: TES_BUG=1 go test ./internal/router -run TestPengajuanBelumBenar -v
func TestPengajuanBelumBenar(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	lain := masuk(t, surelUserLain)
	legal := masukPeran(t, entity.RoleLegal)

	for _, m := range semuaModulPengajuan {
		t.Run(m.nama, func(t *testing.T) {
			// Jalur JSON memakai pemeriksaan kolom wajib yang sama dengan jalur form.
			t.Run("BuatLewatJSONKosongDitolak", func(t *testing.T) {
				w := panggil(t, pemilik, http.MethodPost, m.url(""), map[string]string{})
				harapKode(t, w, http.StatusBadRequest, "membuat pengajuan dengan JSON kosong")
			})

			// page dan limit yang tidak masuk akal diganti nilai bawaan, bukan
			// dipakai membagi.
			t.Run("LimitNolTidakMembuatError", func(t *testing.T) {
				for _, q := range []string{"?limit=0", "?limit=-5", "?page=0&limit=0"} {
					harapKode(t, panggil(t, pemilik, http.MethodGet, m.url("", q), nil), http.StatusOK, "daftar dengan "+q)
				}
			})

			t.Run("PDFMilikOrangLainDitolak", func(t *testing.T) {
				if !m.pdfPemilik {
					t.Skip("modul ini tidak punya route PDF untuk USER")
				}
				p := m.buat(t, pemilik)
				w := panggil(t, lain, http.MethodGet, m.url("", "/", p.ID, "/pdf"), nil)
				harapKode(t, w, http.StatusNotFound, "USER lain meminta PDF")
				harapKode(t, panggil(t, pemilik, http.MethodGet, m.url("", "/", p.ID, "/pdf"), nil), http.StatusOK, "pemilik meminta PDF")
			})

			t.Run("PDFUntukIDTidakAdaDitolak", func(t *testing.T) {
				const tidakAda = "00000000-0000-0000-0000-000000000000"
				harapKode(t, panggil(t, legal, http.MethodGet, m.url("/legal", "/", tidakAda, "/pdf"), nil), http.StatusNotFound, "LEGAL meminta PDF yang tidak ada")
			})


		})
	}
}

// Pengajuan yang dibuat bersamaan semuanya berhasil, dengan nomor tiket unik dan berurutan.
// Nomor diambil di luar transaksi, jadi yang kalah bentrok mengulang dengan nomor berikutnya
// (PR A tahap A.3). Tanpa pengulangan, sebagian permintaan gagal "gagal membuat pengajuan".
func TestNomorTiketBersamaan(t *testing.T) {
	pemilik := masukPeran(t, entity.RoleUser)
	m := semuaModulPengajuan[0]
	// Satu permintaan paling banyak kalah sebanyak permintaan lain yang menang lebih dulu, jadi
	// tiga pengajuan serentak pasti selesai dalam 3 percobaan (batas keputusan A2). Lebih dari itu
	// bisa gagal walau logikanya benar; logika pengulangannya diuji pasti di TestDenganNomorTiket.
	const jumlah = 3

	var wg sync.WaitGroup
	kode := make([]int, jumlah)
	tiket := make([]string, jumlah)
	for i := range jumlah {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w := panggilForm(t, pemilik, http.MethodPost, m.url(""), m.kolom)
			kode[i] = w.Code
			if w.Code == http.StatusCreated {
				var p pengajuan
				if err := json.Unmarshal(w.Body.Bytes(), &struct{ Data *pengajuan }{&p}); err == nil {
					tiket[i] = p.TicketNumber
				}
			}
		}(i)
	}
	wg.Wait()

	unik := map[string]bool{}
	for i := range jumlah {
		if kode[i] != http.StatusCreated || tiket[i] == "" {
			t.Errorf("pengajuan %d: kode %d, tiket %q", i, kode[i], tiket[i])
			continue
		}
		unik[tiket[i]] = true
	}
	if len(unik) != jumlah {
		t.Errorf("%d nomor tiket unik dari %d pengajuan: %v", len(unik), jumlah, tiket)
	}
	// Berurutan tanpa celah.
	urut := slices.Sorted(maps.Keys(unik))
	for i := 1; i < len(urut); i++ {
		if nomorBerikut(t, urut[i-1]) != urut[i] {
			t.Errorf("ada celah nomor di antara %s dan %s", urut[i-1], urut[i])
		}
	}
}
