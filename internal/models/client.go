package models

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ivanneira/lapislazuli/config"
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
	// Prompt para la clasificación
	systemPrompt := `Eres un clasificador preciso que analiza solicitudes y las clasifica según su intención. 
	Tu respuesta debe ser ÚNICAMENTE un objeto JSON con el siguiente formato:
	{
		"intent": "nombre_de_la_intención",
		"confidence": 0.95,
		"metadata": {"key1": "value1", "key2": "value2"},
		"requires_data": true|false
	}
	
	Posibles intenciones: 
	- "pregunta_general": Pregunta que no requiere datos específicos
	- "consulta_datos": Requiere buscar información específica
	- "ejecutar_accion": Requiere realizar alguna acción
	- "no_clasificable": No se puede clasificar claramente
	
	requires_data debe ser true si la solicitud necesita buscar información externa.`

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.config.ClassifierModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: input,
				},
			},
			Temperature: 0.1, // Baja temperatura para decisiones predecibles
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error al clasificar: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no se recibió respuesta del clasificador")
	}

	// Extraer y parsear JSON
	content := resp.Choices[0].Message.Content
	// Limpiar el texto para asegurar que solo tenemos JSON
	content = cleanJSONString(content)

	var classification Classification
	if err := json.Unmarshal([]byte(content), &classification); err != nil {
		return nil, fmt.Errorf("error al parsear clasificación: %w", err)
	}

	return &classification, nil
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
