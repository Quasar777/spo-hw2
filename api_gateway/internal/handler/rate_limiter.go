package handler

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type tokenBucket struct {
	tokens    float64   
	lastCheck time.Time 
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*tokenBucket
	rate    float64       // скорость пополнения токенов (tokens per second)
	burst   float64       // максимальное количество токенов (ёмкость ведра)
	ttl     time.Duration // сколько держать записи о клиенте без активности
}

func NewRateLimiter(rate float64, burst int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*tokenBucket),
		rate:    rate,           // rate = сколько запросов в секунду,
		burst:   float64(burst), // burst = "запас" для кратковременных всплесков.
		ttl:     ttl,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rl.clientKey(r)

		if !rl.allow(key) {
			http.Error(w, `{"error": "rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow проверяет и обновляет состояние ведёрка для клиента.
// Возвращает true, если можно пропустить запрос.
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, ok := rl.clients[key]
	if !ok {
		rl.clients[key] = &tokenBucket{
			tokens:    rl.burst - 1,
			lastCheck: now,
		}
		return true
	}

	elapsed := now.Sub(bucket.lastCheck).Seconds()
	bucket.tokens += elapsed * rl.rate
	if bucket.tokens > rl.burst {
		bucket.tokens = rl.burst
	}
	bucket.lastCheck = now

	if bucket.tokens < 1 {
		return false
	}

	bucket.tokens -= 1
	return true
}

// clientKey пытается определить ключ клиента (по IP).
// В проде можно смотреть X-Forwarded-For, X-Real-IP и т.п.
// Для простоты берём RemoteAddr.
func (rl *RateLimiter) clientKey(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
