package router

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Setiap alamat yang dipanggil frontend harus ada di backend dan menerima peran
// yang memanggilnya. Tes ini membaca pemanggilan api.get/post/put/patch/delete
// di frontend/src, lalu mencocokkannya dengan route sungguhan (app.Routes) dan
// dengan peran di foto hak akses (testdata/akses.golden).
//
// Alamat yang dipilih menurut peran dijabarkan untuk tiap peran memakai
// salinan fungsi pemilihnya di bawah (awalanPerPeran). Kalau fungsi di
// frontend diubah, salinannya di sini ikut diubah.
func TestAlamatFrontendAda(t *testing.T) {
	const folder = "../../../frontend/src"
	// cek-docker.ps1 hanya membawa program tes ke container, tanpa kode frontend.
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		t.Skip("kode frontend tidak ada di sini (mis. di container); tes ini jalan di lokal dan GitHub Actions")
	}
	panggilan, err := bacaPanggilanAPI(folder)
	if err != nil {
		t.Fatalf("membaca frontend: %v", err)
	}
	if len(panggilan) < 100 {
		t.Fatalf("hanya %d pemanggilan API terbaca; pola pencariannya kemungkinan rusak", len(panggilan))
	}

	route := daftarRoute(t)
	terpakai := map[int]bool{}

	for _, p := range panggilan {
		t.Run(p.letak+" "+p.metode+" "+p.ekspresi, func(t *testing.T) {
			for _, v := range jabarkan(p) {
				masalah := cekAlamat(route, p.metode, v.alamat, v.peran)
				if masalah == "" {
					continue
				}
				i := cariPengecualian(p.metode, v.alamat, v.peran)
				if i < 0 {
					t.Errorf("%s %s%s: %s (%s)", p.metode, v.alamat, labelPeran(v.peran), masalah, p.letak)
					continue
				}
				terpakai[i] = true
				if k := pengecualianAlamat[i]; k.bug {
					bugDiketahui(t, k.alasan)
					t.Errorf("%s %s%s: %s (%s)", p.metode, v.alamat, labelPeran(v.peran), masalah, p.letak)
				}
			}
		})
	}

	// Pengecualian yang tidak lagi dibutuhkan harus dihapus, supaya daftarnya
	// tetap jujur. Pemeriksaan ini hanya berarti kalau semua subtes dijalankan.
	if !strings.Contains(flag.Lookup("test.run").Value.String(), "/") {
		for i, k := range pengecualianAlamat {
			if !terpakai[i] {
				t.Errorf("pengecualian tidak terpakai lagi, hapus dari daftar: %s %s [%s]", k.metode, k.alamat, k.peran)
			}
		}
	}
}

// pengecualianAlamat: alamat yang tidak cocok dengan backend tetapi sudah
// diketahui. bug: true berarti rusak dan belum diperbaiki (dilewati lewat
// bugDiketahui, jalankan dengan TES_BUG=1 untuk melihatnya). Selain itu,
// alamatnya memang tidak dipakai dengan izin bawaan.
//
// alamat boleh diakhiri * (awalan). metode dan peran "*" berarti apa saja;
// peran boleh berisi beberapa peran dipisah spasi.
type pengecualian struct {
	metode, alamat, peran, alasan string
	bug                           bool
}

var pengecualianAlamat = []pengecualian{

	// Fitur yang menunya hanya tampil kalau peran itu diberi izin tambahan
	// (override). Ditunda bersama fitur anak usaha.
	{"*", "/legal/*", "LEGAL_AU", "Legal Opinion, Review Dokumen, Audit Log, dan Laporan Legal AU memanggil alamat LEGAL", false},
	{"*", "/legal-au/cases*", "LEGAL_AU", "Legal AU belum punya alamat ubah dan hapus perkara serta tambah cedant", false},
	{"*", "/external/legal-opinions*", "EXTERNAL", "Legal Opinion External belum punya alamat di backend", false},
	{"*", "/external/review-documents*", "EXTERNAL", "Review Dokumen External belum punya alamat di backend", false},
	{"GET", "/admin/audit-logs*", "USER EXTERNAL", "Audit Log USER dan EXTERNAL belum punya alamat sendiri (daftar dan daftar pelaku)", false},
	{"*", "/admin/materials*", "USER EXTERNAL", "USER dan EXTERNAL belum punya alamat tulis Materi Legal", false},
	{"GET", "/reports/*", "USER EXTERNAL", "Laporan USER dan EXTERNAL belum punya alamat di backend", false},
	{"GET", "/case-types", "USER", "form perkara USER (hanya lewat override) memuat rujukan yang tidak menerima USER", false},
	{"GET", "/case-categories", "USER", "form perkara USER (hanya lewat override) memuat rujukan yang tidak menerima USER", false},
	{"GET", "/companies", "USER", "form perkara USER (hanya lewat override) memuat rujukan yang tidak menerima USER", false},

	// Kode yang tidak pernah dijalankan oleh peran itu.
	{"GET", "/purpose-types", "LEGAL LEGAL_AU USER EXTERNAL", "tujuan pembuatan hanya dimuat halaman admin", false},
	{"GET", "/review-documents/X/pdf", "USER", "unduh PDF Review Dokumen hanya ada di halaman admin dan legal", false},
	{"PATCH", "/agreement-documents/X/*", "USER", "aksi approver; detail perjanjian USER dibuka tanpa approver", false},
	{"GET", "/agreement-company-master", "USER", "master Pihak Pertama hanya dimuat approver (apiBase /admin atau /legal); detail perjanjian USER tidak memanggilnya", false},
	{"GET", "/agreement-signatories", "USER", "Daftar Penandatangan hanya dimuat approver (apiBase /admin atau /legal); detail perjanjian USER tidak memanggilnya", false},
}

type panggilanAPI struct {
	letak, metode, ekspresi string
}

type varian struct {
	alamat, peran string // peran kosong: alamat tetap, cukup dicek keberadaannya
}

var (
	polaPanggil = regexp.MustCompile(`api\.(get|post|put|patch|delete)(?:<[^()]*?>)?\(\s*`)
	polaSisipan = regexp.MustCompile(`\$\{[^}]*\}`)
)

func bacaPanggilanAPI(folder string) ([]panggilanAPI, error) {
	var hasil []panggilanAPI
	err := filepath.WalkDir(folder, func(file string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !(strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".tsx")) {
			return err
		}
		rel, _ := filepath.Rel(folder, file)
		rel = filepath.ToSlash(rel)
		if rel == "services/api.ts" {
			return nil
		}
		isi, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		s := string(isi)
		for _, m := range polaPanggil.FindAllStringSubmatchIndex(s, -1) {
			ekspresi := argumenPertama(s[m[1]:])
			if ekspresi == "" {
				continue
			}
			hasil = append(hasil, panggilanAPI{
				letak:    fmt.Sprintf("%s:%d", rel, strings.Count(s[:m[0]], "\n")+1),
				metode:   strings.ToUpper(s[m[2]:m[3]]),
				ekspresi: ekspresi,
			})
		}
		return nil
	})
	return hasil, err
}

// argumenPertama mengembalikan alamat dari awal argumen: isi string ('...',
// "...", `...`), atau pemanggilan fungsi pemilih seperti getAuditLogEndpoint()
// yang ditulis sebagai ${getAuditLogEndpoint()}.
func argumenPertama(s string) string {
	if s == "" {
		return ""
	}
	if q := s[0]; q == '\'' || q == '"' || q == '`' {
		if akhir := strings.IndexByte(s[1:], q); akhir >= 0 {
			return s[1 : akhir+1]
		}
		return ""
	}
	if akhir := strings.IndexByte(s, ')'); akhir > 0 && strings.HasPrefix(s, "get") {
		return "${" + s[:akhir+1] + "}"
	}
	return ""
}

// Panel tiap peran, sesuai routes/index.tsx. Fungsi pemilih di frontend membaca
// peran dan alamat halaman, jadi keduanya disimulasikan bersama.
var halamanPeran = map[string]string{
	"ADMIN": "/admin/x", "LEGAL": "/legal/x", "LEGAL_AU": "/legal-au/x", "USER": "/dashboard/x", "EXTERNAL": "/external/x",
}

// awalanPerPeran: salinan fungsi pemilih alamat di frontend/src/services.
// Ubah bersama aslinya.
var awalanPerPeran = map[string]func(peran, halaman string) []string{
	// legal-case.service.ts getLegalCaseRouteBase
	"getLegalCaseRouteBase()": func(r, p string) []string {
		switch {
		case r == "USER" || strings.HasPrefix(p, "/dashboard"):
			return []string{"/legal-cases"}
		case r == "LEGAL_AU" || strings.HasPrefix(p, "/legal-au"):
			return []string{"/legal-au/cases"}
		case r == "LEGAL" || strings.HasPrefix(p, "/legal/"):
			return []string{"/legal/legal-cases"}
		case r == "EXTERNAL" || strings.HasPrefix(p, "/external"):
			return []string{"/external/legal-cases"}
		}
		return []string{"/admin/legal-cases"}
	},
	// material.service.ts getMaterialEndpoint
	"getMaterialEndpoint()": func(r, p string) []string {
		switch {
		case r == "LEGAL_AU" || strings.HasPrefix(p, "/legal-au"):
			return []string{"/legal-au/materials"}
		case r == "LEGAL" || strings.HasPrefix(p, "/legal/"):
			return []string{"/legal/materials"}
		}
		return []string{"/admin/materials"}
	},
	// legal-opinion.service.tsx getLegalOpinionEndpoint
	"getLegalOpinionEndpoint()": func(r, p string) []string { return []string{pilihPengajuan(r, p, "legal-opinions")} },
	// document-review.service.ts getDocumentReviewEndpoint
	"getDocumentReviewEndpoint()": func(r, p string) []string { return []string{pilihPengajuan(r, p, "review-documents")} },
	// audit-log.service.ts getAuditLogEndpoint
	"getAuditLogEndpoint()": func(r, p string) []string {
		if r == "LEGAL" || strings.HasPrefix(p, "/legal") {
			return []string{"/legal/audit-logs"}
		}
		return []string{"/admin/audit-logs"}
	},
	// report.service.ts getReportEndpoint
	"getReportEndpoint(feature)": func(r, p string) []string {
		awal := "/reports"
		if r == "ADMIN" || strings.HasPrefix(p, "/admin") {
			awal = "/admin/reports"
		} else if r == "LEGAL" || strings.HasPrefix(p, "/legal") {
			awal = "/legal/reports"
		}
		var hasil []string
		for _, fitur := range []string{"legal-cases", "legal-opinions", "document-reviews"} {
			hasil = append(hasil, awal+"/"+fitur)
		}
		return hasil
	},
}

func pilihPengajuan(r, p, sumber string) string {
	switch {
	case r == "EXTERNAL" || strings.HasPrefix(p, "/external"):
		return "/external/" + sumber
	case r == "ADMIN" || strings.HasPrefix(p, "/admin"):
		return "/admin/" + sumber
	case r == "LEGAL" || strings.HasPrefix(p, "/legal"):
		return "/legal/" + sumber
	}
	return "/" + sumber
}

func init() {
	// legal-case.service.ts getReferenceRouteBase
	for _, sumber := range []string{"case-types", "case-categories", "companies", "purpose-types"} {
		awalanPerPeran["getReferenceRouteBase('"+sumber+"')"] = func(r, p string) []string {
			if r == "ADMIN" || strings.HasPrefix(p, "/admin") {
				return []string{"/admin/" + sumber}
			}
			return []string{"/" + sumber}
		}
	}
}

// Nilai ${base} komponen perjanjian per peran, dari properti apiBase di routes/index.tsx.
var basePerjanjian = map[string]string{"USER": "", "ADMIN": "/admin", "LEGAL": "/legal"}

func jabarkan(p panggilanAPI) []varian {
	var hasil []varian
	for kunci, pilih := range awalanPerPeran {
		if !strings.Contains(p.ekspresi, "${"+kunci+"}") {
			continue
		}
		for peran, halaman := range halamanPeran {
			for _, awal := range pilih(peran, halaman) {
				hasil = append(hasil, varian{rapikan(strings.Replace(p.ekspresi, "${"+kunci+"}", awal, 1)), peran})
			}
		}
		return hasil
	}
	if strings.Contains(p.ekspresi, "${base}") {
		for peran, base := range basePerjanjian {
			hasil = append(hasil, varian{rapikan(strings.Replace(p.ekspresi, "${base}", base, 1)), peran})
		}
		return hasil
	}
	return []varian{{rapikan(p.ekspresi), ""}}
}

// rapikan mengganti sisipan ${...} yang tersisa (ID dan sejenisnya) dengan X
// dan membuang query string.
func rapikan(alamat string) string {
	alamat, _, _ = strings.Cut(polaSisipan.ReplaceAllString(alamat, "X"), "?")
	return alamat
}

type routeAPI struct {
	metode string
	pola   *regexp.Regexp
	peran  map[string]bool // nil: route publik, semua peran boleh
}

func daftarRoute(t *testing.T) []routeAPI {
	t.Helper()
	isi, err := os.ReadFile(fileFoto)
	if err != nil {
		t.Fatalf("membaca foto hak akses: %v", err)
	}
	peranRoute := map[string]map[string]bool{}
	for baris := range strings.SplitSeq(strings.ReplaceAll(string(isi), "\r\n", "\n"), "\n") {
		kunci, daftar, ok := strings.Cut(baris, ": ")
		if !ok || strings.HasPrefix(baris, "#") {
			continue
		}
		peran := map[string]bool{}
		for p := range strings.FieldsSeq(daftar) {
			peran[p] = p != "-"
		}
		peranRoute[kunci] = peran
	}

	var hasil []routeAPI
	for _, r := range app.Routes() {
		alamat, ok := strings.CutPrefix(r.Path, "/api/v1")
		if !ok {
			continue
		}
		hasil = append(hasil, routeAPI{
			metode: r.Method,
			pola:   regexp.MustCompile("^" + polaParam.ReplaceAllString(alamat, "[^/]+") + "$"),
			peran:  peranRoute[r.Method+" "+r.Path],
		})
	}
	return hasil
}

// cekAlamat mengembalikan alasan alamat itu bermasalah, atau kosong kalau baik.
// Route yang tidak menerima peran mana pun dengan izin bawaan (tanda - di
// foto) dianggap baik: route itu memang hanya dibuka lewat override.
func cekAlamat(route []routeAPI, metode, alamat, peran string) string {
	var cocok []routeAPI
	for _, r := range route {
		if r.metode == metode && r.pola.MatchString(alamat) {
			cocok = append(cocok, r)
		}
	}
	if len(cocok) == 0 {
		return "tidak ada di backend"
	}
	if peran == "" {
		return ""
	}
	for _, r := range cocok {
		if r.peran == nil || r.peran[peran] || !adaYangBoleh(r.peran) {
			return ""
		}
	}
	return "ditolak untuk peran ini"
}

func adaYangBoleh(peran map[string]bool) bool {
	for _, boleh := range peran {
		if boleh {
			return true
		}
	}
	return false
}

// cariPengecualian mengembalikan indeks pengecualian yang cocok, atau -1.
func cariPengecualian(metode, alamat, peran string) int {
	for i, k := range pengecualianAlamat {
		cocokAlamat := k.alamat == alamat ||
			strings.HasSuffix(k.alamat, "*") && strings.HasPrefix(alamat, strings.TrimSuffix(k.alamat, "*"))
		cocokPeran := k.peran == "*" || peran != "" && strings.Contains(" "+k.peran+" ", " "+peran+" ")
		if (k.metode == "*" || k.metode == metode) && cocokAlamat && cocokPeran {
			return i
		}
	}
	return -1
}

func labelPeran(peran string) string {
	if peran == "" {
		return ""
	}
	return " [" + peran + "]"
}
