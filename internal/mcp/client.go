package mcp

import (
	"github.com/ivanneira/Lapislazuli/internal/logger"
)

// AIClient proporciona una interfaz simple para aplicaciones
type AIClient struct {
	service    AIService
	sessionMgr *SessionManager
}

func NewAIClient(service AIService) *AIClient {
	return &AIClient{
		service:    service,
		sessionMgr: NewSessionManager(),
	}
}

func (c *AIClient) Process(request AIRequest) AIResponse {
	logger.Info("Processing request for session: %s", request.SessionID)
	logger.JSON("Request", request)

	ctx := c.sessionMgr.GetContext(request.SessionID)
	logger.Debug("Retrieved context for session: %s", request.SessionID)

	// Aplicar propiedades al contexto
	for k, v := range request.Properties {
		ctx.SetProperty(k, v)
	}

	output, err := c.service.Process(ctx, request.Input)
	if err != nil {
		logger.Error("Processing error: %v", err)
	} else {
		logger.Debug("Processing successful")
		logger.JSON("Output", output)
	}

	// Guardar contexto actualizado
	c.sessionMgr.SaveContext(ctx)
	logger.Debug("Context saved for session: %s", request.SessionID)

	response := AIResponse{
		Output:  output,
		Context: ctx,
		Error:   err,
	}
	logger.JSON("Response", response)
	return response
}
