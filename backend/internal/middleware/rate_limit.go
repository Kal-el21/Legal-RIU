package middleware

import (
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
	window   time.Duration
}

type loginAttempt struct {
	count   int
	resetAt time.Time
}

// NewLoginRateLimiter creates a rate limiter that tracks attempts per account
func NewLoginRateLimiter(windowMinutes int) *LoginRateLimiter {
	window := time.Duration(windowMinutes) * time.Minute
	if window <= 0 {
		window = 15 * time.Minute
	}

	limiter := &LoginRateLimiter{
		attempts: make(map[string]*loginAttempt),
		window:   window,
	}

	go limiter.cleanup()
	return limiter
}

// Middleware returns middleware that extracts email from request body for rate limiting
func (l *LoginRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read body to extract email for key
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		// Extract email from JSON body (simple string search for performance)
		email := extractEmailFromJSON(string(bodyBytes))
		if email == "" {
			email = c.ClientIP()
		}

		now := time.Now()
		l.mu.Lock()
		attempt, ok := l.attempts[email]
		if !ok || now.After(attempt.resetAt) {
			l.attempts[email] = &loginAttempt{count: 1, resetAt: now.Add(l.window)}
			l.mu.Unlock()
			c.Next()
			return
		}

		if attempt.count >= 5 {
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
	}
}

func extractEmailFromJSON(body string) string {
	// Simple extraction - find "email":"value" pattern
	start := strings.Index(body, `"email"`)
	if start == -1 {
		return ""
	}
	remaining := body[start:]
	colonIdx := strings.Index(remaining, ":")
	if colonIdx == -1 {
		return ""
	}
	start = colonIdx + 1
	// Skip whitespace and quotes
	for start < len(remaining) && (remaining[start] == ' ' || remaining[start] == '"' || remaining[start] == ':') {
		start++
	}
	end := start
	for end < len(remaining) && remaining[end] != '"' && remaining[end] != ',' && remaining[end] != '}' {
		end++
	}
	if start >= end || start >= len(remaining) {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(remaining[start:end]))
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