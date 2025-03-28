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
	return c.messages
}

func (c *Context) GetMetadata() ContextMetadata {
	return c.metadata
}

func (c *Context) SetProperty(key string, value interface{}) {
	logger.Debug("Setting property: %s=%v", key, value)
	c.metadata.Properties[key] = value
}

func (c *Context) GetProperty(key string) interface{} {
	return c.metadata.Properties[key]
}

func (c *Context) Clear() {
	c.messages = c.messages[:0] // Reutilizar slice subyacente
	c.metadata.LastUpdated = time.Now()
}
