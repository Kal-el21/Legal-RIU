package router

import (
	"flag"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"legal-riu-portal/internal/entity"
)

var perbarui = flag.Bool("update", false, "tulis ulang testdata/akses.golden dari perilaku sekarang")

const fileFoto = "testdata/akses.golden"

// Satu-satunya route yang boleh diakses tanpa login. Menambah isi daftar ini
// berarti membuka endpoint ke publik.
var routePublik = map[string]bool{
	"GET /health":                  true,
	"POST /api/v1/auth/ldap-login": true,
	"POST /api/v1/auth/login":      true,
	"POST /api/v1/auth/refresh":    true,
	"POST /api/v1/auth/logout":     true,
}

var polaParam = regexp.MustCompile(`:[A-Za-z]+`)

// routeBerpengaman mengembalikan semua route terdaftar selain yang publik,
// dalam bentuk "METODE /alamat", terurut.
func routeBerpengaman() []string {
	var daftar []string
	for _, r := range app.Routes() {
		kunci := r.Method + " " + r.Path
		if !routePublik[kunci] {
			daftar = append(daftar, kunci)
		}
	}
	sort.Strings(daftar)
	return daftar
}

// panggilRoute memanggil route dengan parameter alamat diisi ID yang tidak ada.
// Yang diuji di file ini hanya pengaman di depan handler, bukan isi jawabannya.
func panggilRoute(t *testing.T, token, route string) int {
	t.Helper()
	metode, alamat, _ := strings.Cut(route, " ")
	alamat = polaParam.ReplaceAllString(alamat, "00000000-0000-0000-0000-000000000000")
	return panggil(t, token, metode, alamat, nil).Code
}

// Setiap endpoint selain login dan /health harus menolak permintaan tanpa
// token. Daftarnya diambil dari router, jadi endpoint baru ikut teruji.
func TestTanpaTokenDitolak(t *testing.T) {
	for _, route := range routeBerpengaman() {
		if kode := panggilRoute(t, "", route); kode != http.StatusUnauthorized {
			t.Errorf("%s tanpa token: dapat %d, harap 401", route, kode)
		}
	}
}

// Foto hak akses: untuk tiap endpoint, peran mana yang lolos pengaman peran dan
// permission. Tes gagal kalau hasilnya berbeda dari testdata/akses.golden.
//
// Kalau perubahan akses memang disengaja, perbarui fotonya lalu commit:
//
//	go test ./internal/router -run TestFotoHakAkses -update
func TestFotoHakAkses(t *testing.T) {
	sekarang := map[string]string{}
	var baris []string
	for _, route := range routeBerpengaman() {
		var lolos []string
		for _, peran := range semuaPeran {
			kode := panggilRoute(t, masukPeran(t, peran), route)
			if kode != http.StatusUnauthorized && kode != http.StatusForbidden {
				lolos = append(lolos, string(peran))
			}
		}
		isi := "-"
		if len(lolos) > 0 {
			isi = strings.Join(lolos, " ")
		}
		sekarang[route] = isi
		baris = append(baris, route+": "+isi)
	}

	if *perbarui {
		kepala := "# Peran yang lolos pengaman tiap endpoint. Tanda - berarti tidak ada.\n" +
			"# Dibuat oleh: go test ./internal/router -run TestFotoHakAkses -update\n"
		if err := os.WriteFile(fileFoto, []byte(kepala+strings.Join(baris, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	data, err := os.ReadFile(fileFoto)
	if err != nil {
		t.Fatalf("membaca foto hak akses: %v (buat dengan -update)", err)
	}
	foto := map[string]string{}
	for b := range strings.SplitSeq(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if b == "" || strings.HasPrefix(b, "#") {
			continue
		}
		route, isi, _ := strings.Cut(b, ": ")
		foto[route] = isi
	}

	for route, isi := range sekarang {
		lama, ada := foto[route]
		switch {
		case !ada:
			t.Errorf("endpoint baru belum ada di foto: %s (lolos: %s)", route, isi)
		case lama != isi:
			t.Errorf("akses berubah: %s\n  foto    : %s\n  sekarang: %s", route, lama, isi)
		}
	}
	for route := range foto {
		if _, ada := sekarang[route]; !ada {
			t.Errorf("endpoint di foto sudah tidak terdaftar: %s", route)
		}
	}
	if t.Failed() {
		t.Log("kalau perubahan ini disengaja: go test ./internal/router -run TestFotoHakAkses -update")
	}
}

// Override per user: ALLOW menambah akses di luar peran, DENY mencabut akses
// bawaan peran, dan menghapus override mengembalikan keadaan semula.
func TestOverridePermission(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	user := masuk(t, surelUserLain)

	var target entity.User
	if err := dbTes.Where("email = ?", surelUserLain).First(&target).Error; err != nil {
		t.Fatal(err)
	}
	atur := func(overrides ...map[string]string) {
		t.Helper()
		if overrides == nil {
			overrides = []map[string]string{}
		}
		w := panggil(t, admin, http.MethodPut, "/api/v1/admin/users/"+target.ID.String()+"/permissions",
			map[string]any{"overrides": overrides})
		if w.Code != http.StatusOK {
			t.Fatalf("mengatur override: status %d, isi %s", w.Code, w.Body)
		}
	}
	harap := func(alamat string, kode int) {
		t.Helper()
		if dapat := panggil(t, user, http.MethodGet, alamat, nil).Code; dapat != kode {
			t.Errorf("GET %s: dapat %d, harap %d", alamat, dapat, kode)
		}
	}
	t.Cleanup(func() { atur() })

	// Bawaan peran USER: boleh melihat repositori dokumen, tidak boleh melihat perusahaan.
	const repositori, perusahaan = "/api/v1/document-repositories", "/api/v1/companies"
	harap(repositori, http.StatusOK)
	harap(perusahaan, http.StatusForbidden)

	atur(
		map[string]string{"code": "master_data.view", "effect": "ALLOW"},
		map[string]string{"code": "document_repository.view", "effect": "DENY"},
	)
	harap(repositori, http.StatusForbidden)
	harap(perusahaan, http.StatusOK)

	atur()
	harap(repositori, http.StatusOK)
	harap(perusahaan, http.StatusForbidden)
}
