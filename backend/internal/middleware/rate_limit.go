package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoginRateLimiter provides per-account rate limiting for login attempts
type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
	limit    int
	window   time.Duration
}

type loginAttempt struct {
	count   int
	resetAt time.Time
}

// NewLoginRateLimiter creates a rate limiter that tracks attempts per account
func NewLoginRateLimiter(limit, windowMinutes int) *LoginRateLimiter {
	if limit <= 0 {
		limit = 5
	}

	window := time.Duration(windowMinutes) * time.Minute
	if window <= 0 {
		window = 15 * time.Minute
	}

	limiter := &LoginRateLimiter{
		attempts: make(map[string]*loginAttempt),
		limit:    limit,
		window:   window,
	}

	go limiter.cleanup()
	return limiter
}

// Middleware membatasi percobaan login per akun. field adalah nama isian di
// body yang dipakai handler sebagai identitas ("email" atau "username"); kalau
// kosong atau tidak ada di body, kuncinya alamat IP.
func (l *LoginRateLimiter) Middleware(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		key := rateLimitKey(field, bodyBytes)
		if key == "" {
			key = "ip:" + c.ClientIP()
		}

		now := time.Now()
		l.mu.Lock()
		attempt, ok := l.attempts[key]
		if !ok || now.After(attempt.resetAt) {
			attempt = &loginAttempt{resetAt: now.Add(l.window)}
			l.attempts[key] = attempt
		}

		if attempt.count >= l.limit {
			resetAfter := int(time.Until(attempt.resetAt).Seconds())
			if resetAfter < 1 {
				resetAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(resetAfter))
			l.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Terlalu banyak percobaan login. Silakan coba lagi nanti.",
			})
			return
		}

		attempt.count++
		l.mu.Unlock()
		c.Next()

		// Login yang berhasil menghapus hitungan, supaya hanya kegagalan yang dibatasi.
		if c.Writer.Status() < http.StatusBadRequest {
			l.mu.Lock()
			delete(l.attempts, key)
			l.mu.Unlock()
		}
	}
}

// rateLimitKey membaca body dengan parser JSON yang sama seperti handler, jadi
// isian lain yang disisipkan di body tidak bisa mengganti kunci.
func rateLimitKey(field string, body []byte) string {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}
	_ = json.Unmarshal(body, &req)

	value := ""
	switch field {
	case "email":
		value = req.Email
	case "username":
		value = req.Username
	}
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	return field + ":" + value
}

func (l *LoginRateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		l.mu.Lock()
		for key, attempt := range l.attempts {
			if now.After(attempt.resetAt) {
				delete(l.attempts, key)
			}
		}
		l.mu.Unlock()
	}
}
