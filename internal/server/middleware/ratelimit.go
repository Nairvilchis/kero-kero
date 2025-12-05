package middleware

import (
	"net/http"
	"sync"
	"time"

	"kero-kero/pkg/errors"

	"golang.org/x/time/rate"
)

// RateLimiter implementa rate limiting por IP
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit
	b        int
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	r := rate.Limit(float64(requestsPerWindow) / window.Seconds())
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		r:        r,
		b:        requestsPerWindow,
	}
}

// getLimiter obtiene o crea un limiter para una IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// Middleware retorna el middleware de rate limiting
func (rl *RateLimiter) Middleware() func(next http.Handler) http.Handler {
	// Limpiar visitantes antiguos cada 5 minutos
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			rl.mu.Lock()
			rl.visitors = make(map[string]*rate.Limiter)
			rl.mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener IP del cliente
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				errors.WriteJSON(w, errors.New(http.StatusTooManyRequests, "Demasiadas solicitudes, intente más tarde"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
