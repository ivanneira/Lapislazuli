package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ivanneira/Lapislazuli/config"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LMResponseFormat struct {
	Type       string                 `json:"type"`
	JSONSchema map[string]interface{} `json:"json_schema"`
}

type LMChatRequest struct {
	Model             string           `json:"model"`
	Messages          []ChatMessage    `json:"messages"`
	ResponseFormat    LMResponseFormat `json:"response_format"`
	Temperature       *float32         `json:"temperature,omitempty"`
	MaxTokens         *int             `json:"max_tokens,omitempty"`
	Stream            bool             `json:"stream"`
	TopK              *int             `json:"top_k,omitempty"`
	TopP              *float32         `json:"top_p,omitempty"`
	MinP              *float32         `json:"min_p,omitempty"`
	RepetitionPenalty *float32         `json:"repetition_penalty,omitempty"`
}

type LMResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func Process(prompt string) (string, error) {
	// Construir el system prompt con las acciones disponibles.
	systemContent := fmt.Sprintf(
		"Acciones disponibles: %s. Clasifica el siguiente prompt devolviendo un JSON con el campo 'action'.",
		strings.Join(config.Config.Actions, ", "),
	)

	fmt.Println(systemContent)

	// Crear mensajes: uno system y uno user.
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: systemContent,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Definir el response_format para obtener la respuesta en JSON.
	responseFormat := LMResponseFormat{
		Type: "json_schema",
		JSONSchema: map[string]interface{}{
			"name":   "classification_response",
			"strict": "true",
			"schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"action"},
			},
		},
	}

	// Asignar los parámetros opcionales solo si son distintos de -1.
	var temp *float32 = nil
	if config.Config.ClassificatorTemperature != -1 {
		temp = &config.Config.ClassificatorTemperature
	}
	var maxTokens *int = nil
	if config.Config.ClassificatorMaxTokens != -1 {
		maxTokens = &config.Config.ClassificatorMaxTokens
	}
	var topK *int = nil
	if config.Config.ClassificatorTopK != -1 {
		topK = &config.Config.ClassificatorTopK
	}
	var topP *float32 = nil
	if config.Config.ClassificatorTopP != -1 {
		topP = &config.Config.ClassificatorTopP
	}
	var minP *float32 = nil
	if config.Config.ClassificatorMinP != -1 {
		minP = &config.Config.ClassificatorMinP
	}
	var repPenalty *float32 = nil
	if config.Config.ClassificatorRepetitionPenalty != -1 {
		repPenalty = &config.Config.ClassificatorRepetitionPenalty
	}

	requestBody := LMChatRequest{
		Model:             config.Config.ClassificatorModelName,
		Messages:          messages,
		ResponseFormat:    responseFormat,
		Temperature:       temp,
		MaxTokens:         maxTokens,
		Stream:            false,
		TopK:              topK,
		TopP:              topP,
		MinP:              minP,
		RepetitionPenalty: repPenalty,
	}

	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", config.Config.ClassificatorLMAPIURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.Config.ClassificatorAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Config.ClassificatorAPIKey))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error en la llamada a LM Studio: %s", string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var lmResp LMResponse
	if err := json.Unmarshal(respBody, &lmResp); err != nil {
		return "", err
	}

	if len(lmResp.Choices) == 0 {
		return "", fmt.Errorf("no se recibieron respuestas del modelo")
	}

	// Se espera que el mensaje devuelto contenga el JSON con el campo "action".
	return lmResp.Choices[0].Message.Content, nil
}
