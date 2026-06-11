package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var csrfStore = &CSRFTokenStore{
	tokens: make(map[string]time.Time),
}

type CSRFTokenStore struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

func (s *CSRFTokenStore) Generate() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[token] = time.Now().Add(1 * time.Hour)
	s.mu.Unlock()
	return token
}

func (s *CSRFTokenStore) Validate(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiry, exists := s.tokens[token]
	if !exists || time.Now().After(expiry) {
		return false
	}
	return true
}

func (s *CSRFTokenStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for token, expiry := range s.tokens {
		if now.After(expiry) {
			delete(s.tokens, token)
		}
	}
}

func init() {
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			csrfStore.Cleanup()
		}
	}()
}

func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			token := csrfStore.Generate()
			c.Header("X-CSRF-Token", token)
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" || !csrfStore.Validate(token) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "CSRF token invalid",
			})
			return
		}
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'self'; base-uri 'self'; form-action 'self'")

		if c.FullPath() != "/health" {
			c.Header("Cache-Control", "no-store")
		}

		c.Next()
	}
}