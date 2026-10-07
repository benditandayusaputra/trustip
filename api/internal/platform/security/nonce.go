package security

import (
	"sync"
	"time"
)

type nonceCache struct {
	mu        sync.Mutex
	lifetime  time.Duration
	seen      map[string]time.Time
	lastSweep time.Time
}

func newNonceCache(lifetime time.Duration) *nonceCache {
	return &nonceCache{lifetime: lifetime, seen: map[string]time.Time{}}
}

func (c *nonceCache) claim(nonce string, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sweepExpired(now)
	if usedAt, ok := c.seen[nonce]; ok && now.Sub(usedAt) < c.lifetime {
		return false
	}
	c.seen[nonce] = now
	return true
}

func (c *nonceCache) sweepExpired(now time.Time) {
	if now.Sub(c.lastSweep) < c.lifetime {
		return
	}
	for nonce, usedAt := range c.seen {
		if now.Sub(usedAt) >= c.lifetime {
			delete(c.seen, nonce)
		}
	}
	c.lastSweep = now
}
