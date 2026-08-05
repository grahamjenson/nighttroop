// A Significant amount of CPU cycles is dedicated to verifying legacy ns tokens.
// So now we are going to cache passwords, only for ns
package security

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/crypto/bcrypt"
)

type passwordCacheEntry struct {
	verifiedAt time.Time
}

type PasswordVerifier struct {
	cache *lru.Cache[string, passwordCacheEntry]
	mu    sync.Mutex
	ttl   time.Duration
}

func NewPasswordVerifier() *PasswordVerifier {
	cache, err := lru.New[string, passwordCacheEntry](10_000)
	if err != nil {
		panic(err)
	}

	return &PasswordVerifier{
		cache: cache,
		ttl:   5 * time.Minute,
	}
}

func (p *PasswordVerifier) VerifyPassword(plainPassword, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
	return err == nil
}

func (p *PasswordVerifier) VerifyPasswordWithCache(plainPassword, storedHash string) bool {
	key := passwordCacheKey(plainPassword, storedHash)

	if entry, ok := p.cache.Get(key); ok {
		if time.Since(entry.verifiedAt) < p.ttl {
			return true
		}

		p.cache.Remove(key)
	}

	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))

	if err == nil {
		p.cache.Add(key, passwordCacheEntry{verifiedAt: time.Now()})
		return true
	}
	return false
}

func passwordCacheKey(password, hash string) string {
	h := sha256.New()
	h.Write([]byte(hash))
	h.Write([]byte{0})
	h.Write([]byte(password))

	return hex.EncodeToString(h.Sum(nil))
}
