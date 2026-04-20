package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type TokenManager struct {
	mu    sync.RWMutex
	token string
}

func NewTokenManager() *TokenManager {
	return &TokenManager{
		token: generateRandomToken(),
	}
}

func (tm *TokenManager) GetToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.token
}

func (tm *TokenManager) Rotate() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.token = generateRandomToken()
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
