package router

// Kerangka tes API: menyalakan aplikasi sungguhan ke database dan bucket khusus
// tes, lalu memanggil endpoint lewat router tanpa membuka port.
//
// Butuh Postgres dan MinIO hidup. Tanpa keduanya tes ini gagal; untuk sengaja
// melewatinya, set TES_API_LEWATI=1. TES_TANPA_MINIO=1 menjalankan tes yang
// tidak menyentuh file dengan Postgres saja.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/entity"
	"legal-riu-portal/internal/seed"
	"legal-riu-portal/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Nama database dan bucket dikunci di sini, tidak dibaca dari environment,
// supaya tes tidak mungkin berjalan di data dev atau produksi.
const (
	namaDBTes     = "legal_riu_test"
	namaBucketTes = "legal-riu-test"
	sandiTes      = "sandi-tes-123"
	surelUserLain = "tes-user-lain@example.com"
)

var semuaPeran = []entity.UserRole{
	entity.RoleAdmin, entity.RoleLegal, entity.RoleLegalAU, entity.RoleUser, entity.RoleExternal,
}

// tanpaMinIO dipakai di GitHub Actions, karena image MinIO tidak bisa diunduh
// lagi di sana. Aplikasi dirakit tanpa penyimpanan file, dan tes yang butuh
// file harus melewati dirinya sendiri. Di lokal jangan di-set.
var tanpaMinIO = os.Getenv("TES_TANPA_MINIO") != ""

var (
	app       *gin.Engine
	dbTes     *gorm.DB
	storeTes  *storage.MinIOClient  // nil kalau TES_TANPA_MINIO
	tokenUser = map[string]string{} // surel -> access token
	tokenCSRF string
)

func TestMain(m *testing.M) {
	if os.Getenv("TES_API_LEWATI") != "" {
		fmt.Fprintln(os.Stderr, "tes API dilewati karena TES_API_LEWATI di-set")
		os.Exit(0)
	}
	if err := siapkan(); err != nil {
		fmt.Fprintln(os.Stderr, "tes API gagal disiapkan:", err)
		fmt.Fprintln(os.Stderr, "nyalakan Postgres dan MinIO, atau set TES_API_LEWATI=1 untuk melewati")
		os.Exit(1)
	}
	// Log aplikasi dibungkam supaya keluaran tes yang gagal tetap terbaca.
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func siapkan() error {
	// Sambungan lokal dibaca dari backend/.env, sama seperti server yang
	// dijalankan dari folder backend. Di CI file itu tidak ada dan nilai bawaan
	// yang dipakai.
	_ = godotenv.Load("../../.env")
	cfg := config.Load()
	cfg.App.Env = "test"
	cfg.Database.Name = namaDBTes
	cfg.MinIO.Bucket = namaBucketTes

	// InitDatabase dan InitMinIO memanggil log.Fatal kalau gagal tersambung,
	// jadi dicek dulu di sini supaya pesannya jelas.
	alamatDB := net.JoinHostPort(cfg.Database.Host, cfg.Database.Port)
	if err := cekHidup(alamatDB); err != nil {
		return fmt.Errorf("Postgres tidak hidup di %s: %w", alamatDB, err)
	}
	if !tanpaMinIO {
		if err := cekHidup(cfg.MinIO.Endpoint); err != nil {
			return fmt.Errorf("MinIO tidak hidup di %s: %w", cfg.MinIO.Endpoint, err)
		}
	}

	if err := buatUlangDB(cfg); err != nil {
		return fmt.Errorf("membuat database %s: %w", namaDBTes, err)
	}
	dbTes = config.InitDatabase(cfg)
	var store *storage.MinIOClient
	if !tanpaMinIO {
		store = storage.InitMinIO(cfg)
	}
	storeTes = store

	// Dua kali: migrasi harus jalan di database kosong dan aman diulang.
	for i := 1; i <= 2; i++ {
		if err := seed.RunAllMigrationsAndSeeds(dbTes, store); err != nil {
			return fmt.Errorf("migrasi dan seed putaran %d: %w", i, err)
		}
	}

	if err := buatUserTes(); err != nil {
		return fmt.Errorf("membuat user tes: %w", err)
	}

	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	var err error
	app, err = New(cfg, dbTes, store)
	return err
}

func cekHidup(alamat string) error {
	conn, err := net.DialTimeout("tcp", alamat, 2*time.Second)
	if err != nil {
		return err
	}
	return conn.Close()
}

// buatUlangDB menghapus lalu membuat lagi database tes, supaya tiap run mulai
// dari database kosong.
func buatUlangDB(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode)
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return err
	}
	sqlDB, err := admin.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := admin.Exec("DROP DATABASE IF EXISTS " + namaDBTes + " WITH (FORCE)").Error; err != nil {
		return err
	}
	return admin.Exec("CREATE DATABASE " + namaDBTes).Error
}

func surelPeran(peran entity.UserRole) string {
	return "tes-" + strings.ToLower(strings.ReplaceAll(string(peran), "_", "-")) + "@example.com"
}

// buatUserTes membuat satu user per peran, plus USER kedua yang dipakai tes
// override permission dan sebagai "orang lain" di tes kepemilikan.
func buatUserTes() error {
	companyID, err := seed.FindFirstCompanyID(dbTes)
	if err != nil {
		return err
	}

	daftar := map[string]entity.UserRole{surelUserLain: entity.RoleUser}
	for _, peran := range semuaPeran {
		daftar[surelPeran(peran)] = peran
	}
	for surel, peran := range daftar {
		if err := tambahUser(surel, peran, companyID); err != nil {
			return err
		}
	}
	return nil
}

// tambahUser membuat user aktif dengan sandi tes, langsung di database.
func tambahUser(surel string, peran entity.UserRole, companyID uuid.UUID) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(sandiTes), bcrypt.MinCost)
	if err != nil {
		return err
	}
	return dbTes.Create(&entity.User{
		FullName:     "Tes " + string(peran),
		Email:        surel,
		PasswordHash: string(hash),
		Position:     "Staff",
		Division:     "Divisi Tes",
		Role:         peran,
		Status:       entity.UserActive,
		CompanyID:    &companyID,
	}).Error
}

// masuk login lewat endpoint asli dan mengembalikan access token. Hasilnya
// disimpan, karena login dibatasi beberapa kali per surel.
func masuk(t testing.TB, surel string) string {
	t.Helper()
	if tkn, ada := tokenUser[surel]; ada {
		return tkn
	}
	w := panggil(t, "", http.MethodPost, "/api/v1/auth/login", map[string]string{"email": surel, "password": sandiTes})
	if w.Code != http.StatusOK {
		t.Fatalf("login %s: status %d, isi %s", surel, w.Code, w.Body)
	}
	var hasil struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &hasil); err != nil || hasil.Data.AccessToken == "" {
		t.Fatalf("login %s: access token tidak ada di jawaban: %s", surel, w.Body)
	}
	tokenUser[surel] = hasil.Data.AccessToken
	return hasil.Data.AccessToken
}

func masukPeran(t testing.TB, peran entity.UserRole) string {
	t.Helper()
	return masuk(t, surelPeran(peran))
}

// panggil mengirim satu permintaan ke router. token kosong berarti tanpa login.
// body nil berarti tanpa isi; selain itu dikirim sebagai JSON.
func panggil(t testing.TB, token, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var isi io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("menyusun isi permintaan: %v", err)
		}
		isi = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, isi)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return kirim(t, token, req)
}

// berkas adalah satu file yang dikirim lewat form.
type berkas struct {
	kolom, nama string
	isi         []byte
}

// panggilForm mengirim multipart/form-data, seperti form unggah di frontend.
func panggilForm(t testing.TB, token, method, path string, kolom map[string]string, files ...berkas) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for nama, nilai := range kolom {
		if err := mw.WriteField(nama, nilai); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range files {
		bagian, err := mw.CreateFormFile(f.kolom, f.nama)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := bagian.Write(f.isi); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return kirim(t, token, req)
}

func kirim(t testing.TB, token string, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			req.Header.Set("X-CSRF-Token", ambilCSRF(t, token))
		}
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w
}

// harapKode menggagalkan tes kalau kode status tidak sesuai, dan menampilkan
// isi jawaban supaya penyebabnya langsung terlihat.
func harapKode(t testing.TB, w *httptest.ResponseRecorder, kode int, konteks string) {
	t.Helper()
	if w.Code != kode {
		t.Fatalf("%s: dapat %d, harap %d\n  isi: %s", konteks, w.Code, kode, w.Body)
	}
}

// isiData membaca kolom "data" dari jawaban API ke dalam tujuan.
func isiData(t testing.TB, w *httptest.ResponseRecorder, tujuan any) {
	t.Helper()
	var amplop struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &amplop); err != nil {
		t.Fatalf("jawaban bukan JSON: %v\n  isi: %s", err, w.Body)
	}
	if err := json.Unmarshal(amplop.Data, tujuan); err != nil {
		t.Fatalf("kolom data tidak terbaca: %v\n  isi: %s", err, w.Body)
	}
}

// bugDiketahui melewati tes yang harapannya benar tetapi belum dipenuhi kode,
// supaya temuan tercatat tanpa membuat seluruh pemeriksaan merah. Setelah
// perbaikannya masuk, hapus panggilannya. TES_BUG=1 menjalankan tesnya apa
// adanya, untuk memastikan sebuah perbaikan benar-benar bekerja.
func bugDiketahui(t testing.TB, ringkas string) {
	t.Helper()
	if os.Getenv("TES_BUG") == "" {
		t.Skip("BUG DIKETAHUI: " + ringkas)
	}
}

// butuhMinIO melewati tes yang mengunggah atau mengunduh file saat aplikasi
// dirakit tanpa penyimpanan file (GitHub Actions).
func butuhMinIO(t testing.TB) {
	t.Helper()
	if tanpaMinIO {
		t.Skip("butuh MinIO; dilewati karena TES_TANPA_MINIO di-set")
	}
}

// ambilCSRF mengambil token CSRF dari jawaban GET, sama seperti yang dilakukan
// frontend. Token berlaku satu jam dan tidak terikat user, jadi cukup sekali.
func ambilCSRF(t testing.TB, token string) string {
	t.Helper()
	if tokenCSRF == "" {
		w := panggil(t, token, http.MethodGet, "/api/v1/auth/me", nil)
		tokenCSRF = w.Header().Get("X-CSRF-Token")
		if tokenCSRF == "" {
			t.Fatalf("token CSRF tidak ada di jawaban GET /auth/me (status %d)", w.Code)
		}
	}
	return tokenCSRF
}
