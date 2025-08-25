package auth

import (
	"sync"
	"time"
)

var (
	mu         sync.Mutex
	challenges = make(map[string]time.Time)
	// challenge TTL
	ttl = 5 * time.Minute
)

func SaveChallenge(ch string) {
	mu.Lock()
	defer mu.Unlock()
	challenges[ch] = time.Now()
}

func ValidateAndDeleteChallenge(ch string) bool {
	mu.Lock()
	defer mu.Unlock()
	t, ok := challenges[ch]
	if !ok {
		return false
	}
	if time.Since(t) > ttl {
		delete(challenges, ch)
		return false
	}
	delete(challenges, ch)
	return true
}
