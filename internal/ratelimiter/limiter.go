package ratelimiter

import (
	"math"
	"sync"
	"time"
)

type Config struct {
	IPRatePerSec     float64
	IPBurst          int
	TokenRatePerSec  float64
	TokenBurst       int
	CleanupInterval  time.Duration
	ClientExpiration time.Duration
}

type bucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
}

type Limiter struct {
	mu      sync.Mutex
	clients map[string]*bucket
	cfg     Config
	stopCh  chan struct{}
}

func NewLimiter(cfg Config) *Limiter {
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = time.Minute
	}
	if cfg.ClientExpiration <= 0 {
		cfg.ClientExpiration = 3 * time.Minute
	}

	l := &Limiter{
		clients: make(map[string]*bucket),
		cfg:     cfg,
		stopCh:  make(chan struct{}),
	}

	go l.cleanupLoop()
	return l
}

func (l *Limiter) Allow(key string, isAPIToken bool) (allowed bool, limit int, remaining int, retryAfterSec int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	rate := l.cfg.IPRatePerSec
	burst := l.cfg.IPBurst
	if isAPIToken {
		rate = l.cfg.TokenRatePerSec
		burst = l.cfg.TokenBurst
	}

	b, exists := l.clients[key]
	if !exists {
		b = &bucket{
			tokens:     float64(burst),
			maxTokens:  float64(burst),
			refillRate: rate,
			lastRefill: now,
		}
		l.clients[key] = b
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.maxTokens, b.tokens+(elapsed*b.refillRate))
		b.lastRefill = now
	}

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, burst, int(math.Floor(b.tokens)), 0
	}

	needed := 1.0 - b.tokens
	waitSec := int(math.Ceil(needed / b.refillRate))
	if waitSec < 1 {
		waitSec = 1
	}
	return false, burst, 0, waitSec
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.cfg.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for k, b := range l.clients {
				if now.Sub(b.lastRefill) > l.cfg.ClientExpiration {
					delete(l.clients, k)
				}
			}
			l.mu.Unlock()
		case <-l.stopCh:
			return
		}
	}
}

func (l *Limiter) Stop() {
	close(l.stopCh)
}
