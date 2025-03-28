package mcp

import (
	"sync"
	"time"

	"github.com/ivanneira/Lapislazuli/internal/logger"
)

var messagePool = sync.Pool{
	New: func() interface{} {
		return &Message{}
	},
}

type ContextMetadata struct {
	SessionID   string
	Created     time.Time
	LastUpdated time.Time
	Properties  map[string]interface{}
}

// Optimizar estructura Message para mejor alineación en memoria
type Message struct {
	Timestamp time.Time
	Role      string
	Content   string
}

type ModelContext interface {
	AddMessage(role, content string)
	GetMessages() []Message
	GetMetadata() ContextMetadata
	SetProperty(key string, value interface{})
	GetProperty(key string) interface{}
	Clear()
}

type Context struct {
	mu       sync.RWMutex
	messages []Message
	metadata ContextMetadata
}

func NewContext(sessionID string) *Context {
	now := time.Now()
	return &Context{
		messages: make([]Message, 0),
		metadata: ContextMetadata{
			SessionID:   sessionID,
			Created:     now,
			LastUpdated: now,
			Properties:  make(map[string]interface{}),
		},
	}
}

func (c *Context) AddMessage(role, content string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := messagePool.Get().(*Message)
	msg.Role = role
	msg.Content = content
	msg.Timestamp = time.Now()

	logger.Debug("Adding message: role=%s, content=%s", role, content)
	c.messages = append(c.messages, *msg)
	c.metadata.LastUpdated = time.Now()

	// Limpiar y devolver el mensaje al pool
	msg.Role = ""
	msg.Content = ""
	messagePool.Put(msg)
}

func (c *Context) GetMessages() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Crear una copia para evitar race conditions
	messages := make([]Message, len(c.messages))
	copy(messages, c.messages)
	return messages
}

func (c *Context) GetMetadata() ContextMetadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Crear una copia profunda de metadata
	properties := make(map[string]interface{}, len(c.metadata.Properties))
	for k, v := range c.metadata.Properties {
		properties[k] = v
	}

	return ContextMetadata{
		SessionID:   c.metadata.SessionID,
		Created:     c.metadata.Created,
		LastUpdated: c.metadata.LastUpdated,
		Properties:  properties,
	}
}

func (c *Context) SetProperty(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Debug("Setting property: %s=%v", key, value)
	c.metadata.Properties[key] = value
	c.metadata.LastUpdated = time.Now()
}

func (c *Context) GetProperty(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metadata.Properties[key]
}

func (c *Context) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages = c.messages[:0] // Reutilizar slice subyacente
	c.metadata.LastUpdated = time.Now()

	// Limpiar propiedades manteniendo la capacidad
	for k := range c.metadata.Properties {
		delete(c.metadata.Properties, k)
	}
}
