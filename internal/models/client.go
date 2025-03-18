package models

import (
	"context"
	"errors"
	"strings"

	"github.com/ivanneira/Lapislazuli/config"
	"github.com/sashabaranov/go-openai"
)

// LMStudioClient cliente para interactuar con la API de LMStudio
type LMStudioClient struct {
	client *openai.Client
	config *config.LMStudioConfig
}

// Classification resultado de la clasificación
type Classification struct {
	Intent       string                 `json:"intent"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata"`
	RequiresData bool                   `json:"requires_data"`
}

// NewLMStudioClient crea un nuevo cliente de LMStudio
func NewLMStudioClient(config *config.LMStudioConfig) *LMStudioClient {
	clientConfig := openai.DefaultConfig("")
	clientConfig.BaseURL = config.Endpoint

	client := openai.NewClientWithConfig(clientConfig)

	return &LMStudioClient{
		client: client,
		config: config,
	}
}

// Classify clasifica una solicitud usando el modelo clasificador
func (c *LMStudioClient) Classify(ctx context.Context, input string) (*Classification, error) {
	// Simulación de clasificación (reemplazar con lógica real)
	if input == "" {
		return nil, errors.New("input vacío")
	}

	// Ejemplo de clasificación simulada
	return &Classification{
		Intent:       "enviar_correo",
		Confidence:   0.95,
		RequiresData: false,
	}, nil
}

// cleanJSONString limpia el texto para extraer solo el JSON
func cleanJSONString(input string) string {
	// Encontrar el primer { y el último }
	start := strings.Index(input, "{")
	end := strings.LastIndex(input, "}")

	if start >= 0 && end >= 0 && end > start {
		return input[start : end+1]
	}

	return input
}
