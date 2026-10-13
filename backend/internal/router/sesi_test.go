package router

// Login dan sesi, pengaturan akun sendiri, dan pengelolaan user oleh ADMIN.
//
// Tes di file ini memakai user khusus, tidak user per peran dari kerangka tes,
// karena mengubah sandi, status, dan hitungan percobaan login.

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"legal-riu-portal/internal/entity"

	"github.com/google/uuid"
)

type hasilLogin struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Permissions  []string `json:"permissions"`
	User         akun     `json:"user"`
}

type akun struct {
	ID                 string `json:"id"`
	FullName           string `json:"full_name"`
	Email              string `json:"email"`
	Status             string `json:"status"`
	EmailNotifications bool   `json:"email_notifications"`
	TwoFAEnabled       bool   `json:"two_fa_enabled"`
}

// cobaLogin memanggil endpoint login apa adanya, tanpa menyimpan tokennya.
func cobaLogin(t *testing.T, surel, sandi string) *httptest.ResponseRecorder {
	t.Helper()
	return panggil(t, "", http.MethodPost, "/api/v1/auth/login", map[string]string{"email": surel, "password": sandi})
}

// userKhusus membuat user baru berperan USER di perusahaan pertama.
func userKhusus(t *testing.T, surel string) {
	t.Helper()
	var companyID string
	if err := dbTes.Raw("SELECT company_id FROM users WHERE email = ?", surelPeran(entity.RoleUser)).Scan(&companyID).Error; err != nil {
		t.Fatal(err)
	}
	if err := tambahUser(surel, entity.RoleUser, uuid.MustParse(companyID)); err != nil {
		t.Fatal(err)
	}
}

func akunSaya(t *testing.T, token string) akun {
	t.Helper()
	w := panggil(t, token, http.MethodGet, "/api/v1/auth/me", nil)
	harapKode(t, w, http.StatusOK, "mengambil akun sendiri")
	var a akun
	isiData(t, w, &a)
	return a
}

// Login, perpanjangan sesi, dan logout.
func TestSesiLogin(t *testing.T) {
	const surel = "tes-sesi@example.com"
	userKhusus(t, surel)

	// Surel yang tidak terdaftar dan sandi yang salah dijawab dengan pesan yang
	// sama, supaya tidak bisa dipakai menebak surel mana yang terdaftar.
	salah := cobaLogin(t, surel, "sandi-yang-salah")
	asing := cobaLogin(t, "tidak-terdaftar@example.com", sandiTes)
	harapKode(t, salah, http.StatusBadRequest, "login dengan sandi salah")
	harapKode(t, asing, http.StatusBadRequest, "login dengan surel tidak terdaftar")
	if salah.Body.String() != asing.Body.String() {
		t.Errorf("pesan berbeda untuk sandi salah dan surel tidak terdaftar:\n  %s\n  %s", salah.Body, asing.Body)
	}
	harapKode(t, cobaLogin(t, "bukan-surel", sandiTes), http.StatusBadRequest, "login dengan surel tidak valid")

	w := cobaLogin(t, surel, sandiTes)
	harapKode(t, w, http.StatusOK, "login")
	var sesi hasilLogin
	isiData(t, w, &sesi)
	if sesi.AccessToken == "" || sesi.RefreshToken == "" || sesi.User.Email != surel {
		t.Fatalf("jawaban login tidak lengkap: %s", w.Body)
	}
	if !slices.Contains(sesi.Permissions, "legal_opinion.view.own") {
		t.Errorf("permission USER di jawaban login: %v", sesi.Permissions)
	}
	if a := akunSaya(t, sesi.AccessToken); a.Email != surel {
		t.Errorf("akun dari token: %+v", a)
	}
	harapKode(t, panggil(t, "token-ngawur", http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "token tidak sah")

	// Perpanjangan sesi mengganti refresh token; yang lama tidak bisa dipakai lagi.
	w = panggil(t, "", http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": sesi.RefreshToken})
	harapKode(t, w, http.StatusOK, "memperpanjang sesi")
	var baru hasilLogin
	isiData(t, w, &baru)
	if baru.RefreshToken == "" || baru.RefreshToken == sesi.RefreshToken {
		t.Errorf("refresh token tidak diganti saat sesi diperpanjang")
	}
	akunSaya(t, baru.AccessToken)
	w = panggil(t, "", http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": sesi.RefreshToken})
	harapKode(t, w, http.StatusUnauthorized, "memakai ulang refresh token lama")

	// Logout mematikan refresh token.
	w = panggil(t, "", http.MethodPost, "/api/v1/auth/logout", map[string]string{"refresh_token": baru.RefreshToken})
	harapKode(t, w, http.StatusOK, "logout")
	w = panggil(t, "", http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": baru.RefreshToken})
	harapKode(t, w, http.StatusUnauthorized, "memperpanjang sesi setelah logout")
}

// Percobaan login yang gagal berulang kali dikunci, termasuk untuk sandi yang benar.
func TestBatasLogin(t *testing.T) {
	const surel = "tes-batas@example.com"
	userKhusus(t, surel)

	terkunci := false
	for i := 1; i <= 20 && !terkunci; i++ {
		terkunci = cobaLogin(t, surel, "sandi-yang-salah").Code == http.StatusTooManyRequests
	}
	if !terkunci {
		t.Fatal("20 kali login gagal tidak pernah dikunci")
	}
	w := cobaLogin(t, surel, sandiTes)
	harapKode(t, w, http.StatusTooManyRequests, "login dengan sandi benar saat terkunci")
	if w.Header().Get("Retry-After") == "" {
		t.Error("jawaban terkunci tidak menyebut kapan boleh mencoba lagi (Retry-After)")
	}

	// Kuncinya per surel: user lain tetap bisa login.
	harapKode(t, cobaLogin(t, surelPeran(entity.RoleUser), sandiTes), http.StatusOK, "user lain login saat satu surel terkunci")
}

// Ganti sandi dan pengaturan akun sendiri.
func TestPengaturanAkun(t *testing.T) {
	const surel, sandiBaru = "tes-akun@example.com", "sandi-baru-456"
	userKhusus(t, surel)
	token := masuk(t, surel)

	t.Run("GantiSandi", func(t *testing.T) {
		const alamat = "/api/v1/auth/change-password"
		harapKode(t, panggil(t, token, http.MethodPost, alamat, map[string]string{"current_password": "salah", "new_password": sandiBaru}), http.StatusBadRequest, "sandi lama salah")
		harapKode(t, panggil(t, token, http.MethodPost, alamat, map[string]string{"current_password": sandiTes, "new_password": "pendek"}), http.StatusBadRequest, "sandi baru terlalu pendek")
		harapKode(t, panggil(t, token, http.MethodPost, alamat, map[string]string{"current_password": sandiTes, "new_password": sandiBaru}), http.StatusOK, "ganti sandi")
		harapKode(t, cobaLogin(t, surel, sandiTes), http.StatusBadRequest, "login dengan sandi lama")
		harapKode(t, cobaLogin(t, surel, sandiBaru), http.StatusOK, "login dengan sandi baru")
	})

	t.Run("Profil", func(t *testing.T) {
		var divisi string
		dbTes.Raw("SELECT id FROM divisions WHERE deleted_at IS NULL LIMIT 1").Scan(&divisi)
		w := panggil(t, token, http.MethodPut, "/api/v1/settings/profile", map[string]string{"full_name": "Nama Baru", "position": "Manajer", "division": divisi})
		harapKode(t, w, http.StatusOK, "mengubah profil")
		if a := akunSaya(t, token); a.FullName != "Nama Baru" {
			t.Errorf("nama setelah diubah: %q", a.FullName)
		}
		w = panggil(t, token, http.MethodPut, "/api/v1/settings/profile", map[string]string{"full_name": "Nama Baru", "position": "Manajer", "division": "Divisi Yang Tidak Ada"})
		harapKode(t, w, http.StatusBadRequest, "profil dengan divisi yang tidak ada")
	})

	t.Run("NotifikasiDan2FA", func(t *testing.T) {
		harapKode(t, panggil(t, token, http.MethodPut, "/api/v1/settings/notifications", map[string]bool{"email_notifications": false}), http.StatusOK, "mematikan notifikasi surel")
		if akunSaya(t, token).EmailNotifications {
			t.Error("notifikasi surel masih menyala setelah dimatikan")
		}
		harapKode(t, panggil(t, token, http.MethodPut, "/api/v1/settings/two-fa", map[string]any{"enabled": true, "password": "salah"}), http.StatusBadRequest, "2FA dengan sandi salah")
		harapKode(t, panggil(t, token, http.MethodPut, "/api/v1/settings/two-fa", map[string]any{"enabled": true, "password": sandiBaru}), http.StatusOK, "menyalakan 2FA")
		if !akunSaya(t, token).TwoFAEnabled {
			t.Error("2FA belum menyala setelah dinyalakan")
		}
	})
}

// ADMIN mengelola user: buat, ubah, nonaktifkan, reset sandi.
func TestKelolaUser(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/admin/users"
	const surel = "tes-kelola@example.com"
	var divisi string
	dbTes.Raw("SELECT id FROM divisions WHERE deleted_at IS NULL LIMIT 1").Scan(&divisi)
	baru := func() map[string]string {
		return map[string]string{"full_name": "User Kelola", "email": surel, "password": sandiTes, "position": "Staff", "division": divisi, "role": "USER"}
	}

	w := panggil(t, admin, http.MethodPost, alamat, baru())
	harapKode(t, w, http.StatusCreated, "membuat user")
	var u akun
	isiData(t, w, &u)

	t.Run("Validasi", func(t *testing.T) {
		harapKode(t, panggil(t, admin, http.MethodPost, alamat, baru()), http.StatusBadRequest, "surel kembar")
		kasus := map[string]func(map[string]string){
			"peran tidak dikenal":       func(m map[string]string) { m["email"], m["role"] = "lain-1@example.com", "DEWA" },
			"sandi terlalu pendek":      func(m map[string]string) { m["email"], m["password"] = "lain-2@example.com", "pendek" },
			"LEGAL_AU tanpa perusahaan": func(m map[string]string) { m["email"], m["role"] = "lain-3@example.com", "LEGAL_AU" },
			"divisi tidak ada":          func(m map[string]string) { m["email"], m["division"] = "lain-4@example.com", "Divisi Yang Tidak Ada" },
		}
		for nama, rusak := range kasus {
			data := baru()
			rusak(data)
			harapKode(t, panggil(t, admin, http.MethodPost, alamat, data), http.StatusBadRequest, nama)
		}
	})

	t.Run("UbahDanCari", func(t *testing.T) {
		w := panggil(t, admin, http.MethodPut, alamat+"/"+u.ID, map[string]string{"full_name": "User Kelola Diubah", "position": "Staff", "division": divisi})
		harapKode(t, w, http.StatusOK, "mengubah user")
		w = panggil(t, admin, http.MethodGet, alamat+"?search=Kelola+Diubah&limit=100", nil)
		harapKode(t, w, http.StatusOK, "mencari user")
		var daftar struct {
			Items []akun `json:"items"`
		}
		isiData(t, w, &daftar)
		if len(daftar.Items) != 1 || daftar.Items[0].ID != u.ID {
			t.Errorf("pencarian user: %+v", daftar.Items)
		}
	})

	t.Run("NonaktifkanDanAktifkan", func(t *testing.T) {
		w := cobaLogin(t, surel, sandiTes)
		harapKode(t, w, http.StatusOK, "login sebelum dinonaktifkan")
		var sesi hasilLogin
		isiData(t, w, &sesi)

		harapKode(t, panggil(t, admin, http.MethodPatch, alamat+"/"+u.ID+"/status", map[string]string{"status": "TIDAK-ADA"}), http.StatusBadRequest, "status tidak dikenal")
		harapKode(t, panggil(t, admin, http.MethodPatch, alamat+"/"+u.ID+"/status", map[string]string{"status": "INACTIVE"}), http.StatusOK, "menonaktifkan user")
		harapKode(t, cobaLogin(t, surel, sandiTes), http.StatusBadRequest, "login setelah dinonaktifkan")
		w = panggil(t, "", http.MethodPost, "/api/v1/auth/refresh", map[string]string{"refresh_token": sesi.RefreshToken})
		harapKode(t, w, http.StatusUnauthorized, "memperpanjang sesi setelah dinonaktifkan")

		t.Run("TokenLamaDitolak", func(t *testing.T) {
			bugDiketahui(t, "access token user yang sudah dinonaktifkan tetap diterima sampai kedaluwarsa, karena status user tidak diperiksa saat token dipakai")
			harapKode(t, panggil(t, sesi.AccessToken, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "memakai token setelah dinonaktifkan")
		})

		harapKode(t, panggil(t, admin, http.MethodPatch, alamat+"/"+u.ID+"/status", map[string]string{"status": "ACTIVE"}), http.StatusOK, "mengaktifkan lagi")
		harapKode(t, cobaLogin(t, surel, sandiTes), http.StatusOK, "login setelah diaktifkan lagi")
	})

	t.Run("ResetSandi", func(t *testing.T) {
		const sandiReset = "sandi-reset-789"
		harapKode(t, panggil(t, admin, http.MethodPost, alamat+"/"+u.ID+"/reset-password", map[string]string{"new_password": "pendek"}), http.StatusBadRequest, "reset dengan sandi pendek")
		harapKode(t, panggil(t, admin, http.MethodPost, alamat+"/"+u.ID+"/reset-password", map[string]string{"new_password": sandiReset}), http.StatusOK, "reset sandi")
		harapKode(t, cobaLogin(t, surel, sandiTes), http.StatusBadRequest, "login dengan sandi sebelum reset")
		harapKode(t, cobaLogin(t, surel, sandiReset), http.StatusOK, "login dengan sandi hasil reset")
	})
}

// Perusahaan user mengikuti perannya (docs/case_company_access.md): LEGAL selalu
// terdaftar di induk, LEGAL_AU wajib di anak usaha, USER dan ADMIN mengikuti
// pilihan admin, EXTERNAL tanpa perusahaan. Kalau LEGAL bisa terdaftar di anak
// usaha, ia ikut dibatasi ke perkara anak usaha itu; kalau LEGAL_AU bisa
// terdaftar di induk, ia melihat perkara semua perusahaan.
func TestPerusahaanUserPerPeran(t *testing.T) {
	admin := masukPeran(t, entity.RoleAdmin)
	const alamat = "/api/v1/admin/users"
	var divisi, induk, asei string
	dbTes.Raw("SELECT id FROM divisions WHERE deleted_at IS NULL LIMIT 1").Scan(&divisi)
	dbTes.Raw("SELECT id FROM companies WHERE email_domain = ? AND deleted_at IS NULL", entity.ParentCompanyDomain).Scan(&induk)
	dbTes.Raw("SELECT id FROM companies WHERE email_domain = ? AND deleted_at IS NULL", "asei.co.id").Scan(&asei)
	if divisi == "" || induk == "" || asei == "" {
		t.Fatalf("data awal tidak lengkap: divisi %q, induk %q, ASEI %q", divisi, induk, asei)
	}

	// tersimpan membaca peran dan perusahaan user langsung dari database.
	tersimpan := func(id string) (peran, perusahaan string) {
		t.Helper()
		var baris struct{ Role, Perusahaan string }
		dbTes.Raw("SELECT role, COALESCE(company_id::text, '') AS perusahaan FROM users WHERE id = ?", id).Scan(&baris)
		return baris.Role, baris.Perusahaan
	}
	buat := func(tanda, peran, dipilih string) *httptest.ResponseRecorder {
		t.Helper()
		data := map[string]string{"full_name": "User Peran", "email": "tes-peran-" + tanda + "@example.com", "password": sandiTes, "position": "Staff", "division": divisi, "role": peran}
		if dipilih != "" {
			data["company_id"] = dipilih
		}
		return panggil(t, admin, http.MethodPost, alamat, data)
	}

	kasus := []struct {
		tanda, peran, dipilih string
		kode                  int
		harap                 string // perusahaan yang tersimpan
	}{
		{"legal-kosong", "LEGAL", "", http.StatusCreated, induk},
		{"legal-anak-usaha", "LEGAL", asei, http.StatusCreated, induk},
		{"au-anak-usaha", "LEGAL_AU", asei, http.StatusCreated, asei},
		{"au-induk", "LEGAL_AU", induk, http.StatusBadRequest, ""},
		{"au-kosong", "LEGAL_AU", "", http.StatusBadRequest, ""},
		{"user-anak-usaha", "USER", asei, http.StatusCreated, asei},
		{"user-kosong", "USER", "", http.StatusCreated, induk},
		{"admin-anak-usaha", "ADMIN", asei, http.StatusCreated, asei},
		{"external", "EXTERNAL", asei, http.StatusCreated, ""},
	}
	for _, k := range kasus {
		w := buat(k.tanda, k.peran, k.dipilih)
		if w.Code != k.kode {
			t.Errorf("%s: dapat %d, harap %d; isi %s", k.tanda, w.Code, k.kode, w.Body)
			continue
		}
		if k.kode != http.StatusCreated {
			continue
		}
		var u akun
		isiData(t, w, &u)
		if _, got := tersimpan(u.ID); got != k.harap {
			t.Errorf("%s: terdaftar di perusahaan %q, harap %q", k.tanda, got, k.harap)
		}
	}

	// Perusahaan yang diganti lewat Edit harus benar-benar tersimpan.
	t.Run("PindahPerusahaan", func(t *testing.T) {
		w := buat("pindah", "USER", induk)
		harapKode(t, w, http.StatusCreated, "membuat user di induk")
		var u akun
		isiData(t, w, &u)
		w = panggil(t, admin, http.MethodPut, alamat+"/"+u.ID, map[string]string{"full_name": "User Peran", "position": "Staff", "division": divisi, "role": "USER", "company_id": asei})
		harapKode(t, w, http.StatusOK, "memindahkan user ke anak usaha")
		if _, got := tersimpan(u.ID); got != asei {
			t.Errorf("user dipindah ke anak usaha, tetapi tersimpan di %q", got)
		}
	})

	t.Run("GantiPeran", func(t *testing.T) {
		w := buat("ganti", "LEGAL_AU", asei)
		harapKode(t, w, http.StatusCreated, "membuat Legal AU")
		var u akun
		isiData(t, w, &u)
		ganti := func(peran, dipilih string) int {
			t.Helper()
			return panggil(t, admin, http.MethodPut, alamat+"/"+u.ID, map[string]string{"full_name": "User Peran", "position": "Staff", "division": divisi, "role": peran, "company_id": dipilih}).Code
		}
		harapTersimpan := func(konteks, peran, perusahaan string) {
			t.Helper()
			if p, c := tersimpan(u.ID); p != peran || c != perusahaan {
				t.Errorf("%s: tersimpan peran %s di %q, harap %s di %q", konteks, p, c, peran, perusahaan)
			}
		}

		// Form masih membawa perusahaan lamanya; yang tersimpan tetap induk.
		if kode := ganti("LEGAL", asei); kode != http.StatusOK {
			t.Fatalf("Legal AU menjadi LEGAL: dapat %d, harap 200", kode)
		}
		harapTersimpan("Legal AU menjadi LEGAL", "LEGAL", induk)

		if kode := ganti("LEGAL_AU", induk); kode != http.StatusBadRequest {
			t.Errorf("LEGAL menjadi Legal AU tanpa memilih anak usaha: dapat %d, harap 400", kode)
		}
		harapTersimpan("setelah ditolak", "LEGAL", induk)

		if kode := ganti("LEGAL_AU", asei); kode != http.StatusOK {
			t.Errorf("LEGAL menjadi Legal AU di anak usaha: dapat %d, harap 200", kode)
		}
		harapTersimpan("LEGAL menjadi Legal AU", "LEGAL_AU", asei)
	})
}
