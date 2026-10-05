package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimitKey(t *testing.T) {
	cases := []struct{ field, body, want string }{
		{"username", `{"username":" Nurman ","password":"x"}`, "username:nurman"},
		{"email", `{"email":"Budi@Indonesiare.co.id","password":"x"}`, "email:budi@indonesiare.co.id"},
		// Isian "email" sisipan tidak boleh mengganti kunci login LDAP.
		{"username", `{"email":"acak1","username":"nurman","password":"x"}`, "username:nurman"},
		// Kunci ganda: handler memakai yang terakhir, limiter harus sama.
		{"email", `{"email":"acak","email":"budi@x.id"}`, "email:budi@x.id"},
		{"username", `{"password":"x"}`, ""},
		{"", `{"email":"budi@x.id"}`, ""},
		{"email", `bukan json`, ""},
	}
	for _, c := range cases {
		if got := rateLimitKey(c.field, []byte(c.body)); got != c.want {
			t.Errorf("rateLimitKey(%q, %s) = %q, mau %q", c.field, c.body, got, c.want)
		}
	}
}

func TestLoginRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewLoginRateLimiter(2, 15)
	r := gin.New()
	r.POST("/login", limiter.Middleware("username"), func(c *gin.Context) {
		if c.Query("ok") == "1" {
			c.Status(http.StatusOK)
			return
		}
		c.Status(http.StatusBadRequest)
	})
	post := func(url, body string) int {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, url, strings.NewReader(body)))
		return w.Code
	}

	// Gagal 2x, percobaan ke-3 diblokir walau disisipi "email" acak.
	post("/login", `{"username":"a"}`)
	post("/login", `{"username":"a"}`)
	if code := post("/login", `{"email":"acak","username":"A"}`); code != http.StatusTooManyRequests {
		t.Fatalf("percobaan ke-3 = %d, mau 429", code)
	}
	// Username lain dari IP yang sama tidak ikut terblokir.
	if code := post("/login", `{"username":"b"}`); code != http.StatusBadRequest {
		t.Fatalf("username lain = %d, mau 400", code)
	}
	// Login berhasil menghapus hitungan.
	post("/login?ok=1", `{"username":"b"}`)
	post("/login", `{"username":"b"}`)
	if code := post("/login", `{"username":"b"}`); code != http.StatusBadRequest {
		t.Fatalf("setelah login berhasil = %d, mau 400", code)
	}
}
