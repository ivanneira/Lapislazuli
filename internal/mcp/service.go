package mcp

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
)

type AIService interface {
	Process(ctx ModelContext, input string) (string, error)
	LoadContext(sessionID string) (ModelContext, error)
	SaveContext(ctx ModelContext) error
}

type AIRequest struct {
	Input      string
	SessionID  string
	Properties map[string]interface{}
}

type AIResponse struct {
	Output  string
	Context ModelContext
	Error   error
}

// SessionManager maneja las sesiones de contexto
type SessionManager struct {
	sync.RWMutex
	contexts map[string]ModelContext
	// Agregar LRU cache para sesiones frecuentes
	cache *lru.Cache
}

func NewSessionManager() *SessionManager {
	cache, _ := lru.New(1000) // Cachear hasta 1000 sesiones
	return &SessionManager{
		contexts: make(map[string]ModelContext),
		cache:    cache,
	}
}

func (sm *SessionManager) GetContext(sessionID string) ModelContext {
	sm.RLock()
	// Intentar obtener del cache primero
	if ctx, ok := sm.cache.Get(sessionID); ok {
		sm.RUnlock()
		return ctx.(ModelContext)
	}

	if ctx, exists := sm.contexts[sessionID]; exists {
		sm.cache.Add(sessionID, ctx)
		sm.RUnlock()
		return ctx
	}
	sm.RUnlock()

	// Si no existe, crear nuevo con write lock
	sm.Lock()
	defer sm.Unlock()
	ctx := NewContext(sessionID)
	sm.contexts[sessionID] = ctx
	sm.cache.Add(sessionID, ctx)
	return ctx
}

func (sm *SessionManager) SaveContext(ctx ModelContext) {
	sm.Lock()
	defer sm.Unlock()
	sm.contexts[ctx.GetMetadata().SessionID] = ctx
	sm.cache.Add(ctx.GetMetadata().SessionID, ctx)
}
