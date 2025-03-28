package processor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ivanneira/Lapislazuli/config"
	"github.com/ivanneira/Lapislazuli/internal/logger"
	"github.com/ivanneira/Lapislazuli/internal/mcp"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
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

// RequestPayload representa el JSON enviado al modelo LLM.
type RequestPayload struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	TopK        int     `json:"top_k"`
	TopP        float32 `json:"top_p"`
	MinP        float32 `json:"min_p"`
}

// Nueva función auxiliar para manejar solicitudes HTTP
func sendLMRequest(requestBody LMChatRequest) (*LMResponse, error) {
	logger.Info("Iniciando petición LLM")
	logger.JSON("Request body", requestBody)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := json.NewEncoder(buf).Encode(requestBody); err != nil {
		logger.Error("Error codificando JSON: %v", err)
		return nil, err
	}

	logger.Debug("URL destino: %s", config.Config.ClassificatorLMAPIURL)
	req, err := http.NewRequest("POST", config.Config.ClassificatorLMAPIURL, buf)
	if err != nil {
		logger.Error("Error creando request: %v", err)
		return nil, err
	}

	logger.Debug("Configurando headers")
	req.Header.Set("Content-Type", "application/json")
	if config.Config.ClassificatorAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Config.ClassificatorAPIKey))
	}

	logger.Info("Enviando petición HTTP")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("HTTP request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	logger.Info("Respuesta recibida, estado: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Error("Error en la llamada a LM Studio: %s", string(bodyBytes))
		return nil, fmt.Errorf("error en la llamada a LM Studio: %s", string(bodyBytes))
	}

	var lmResp LMResponse
	if err := json.NewDecoder(resp.Body).Decode(&lmResp); err != nil {
		logger.Error("Error decodificando respuesta: %v", err)
		return nil, err
	}

	if len(lmResp.Choices) == 0 {
		return nil, fmt.Errorf("no se recibieron respuestas del modelo")
	}

	logger.JSON("Respuesta completa", lmResp)
	return &lmResp, nil
}

// Nueva función auxiliar para crear el cuerpo de la solicitud
func createLMRequestBody(messages []ChatMessage) LMChatRequest {
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

	return LMChatRequest{
		Model:    config.Config.ClassificatorModelName,
		Messages: messages,
		ResponseFormat: LMResponseFormat{
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
		},
		Temperature:       temp,
		MaxTokens:         maxTokens,
		Stream:            false,
		TopK:              topK,
		TopP:              topP,
		MinP:              minP,
		RepetitionPenalty: repPenalty,
	}
}

// Process realiza la clasificación usando el modelo clasificador.
func Process(prompt string) (string, error) {
	messages := []ChatMessage{
		{
			Role: "system",
			Content: fmt.Sprintf("Acciones disponibles: %s. Clasifica el siguiente prompt devolviendo un JSON con el campo 'action'.",
				strings.Join(config.Config.Actions, ", ")),
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	requestBody := createLMRequestBody(messages)
	resp, err := sendLMRequest(requestBody)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

// ProcessWithContext realiza la clasificación usando el contexto del modelo
func ProcessWithContext(ctx mcp.ModelContext, prompt string) (string, error) {
	logger.Info("=== Iniciando procesamiento con contexto ===")
	logger.Debug("Prompt recibido: %s", prompt)
	logger.JSON("Contexto actual", ctx.GetMessages())

	systemContent := fmt.Sprintf(
		"Acciones disponibles: %s. Clasifica el siguiente prompt devolviendo un JSON con el campo 'action'.",
		strings.Join(config.Config.Actions, ", "),
	)
	logger.Debug("System prompt: %s", systemContent)
	ctx.AddMessage("system", systemContent)
	ctx.AddMessage("user", prompt)

	messages := make([]ChatMessage, 0)
	for _, msg := range ctx.GetMessages() {
		messages = append(messages, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	requestBody := createLMRequestBody(messages)
	resp, err := sendLMRequest(requestBody)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

// Respond realiza una respuesta usando el modelo de respuestas.
func Respond(prompt string) (string, error) {
	apiURL := os.Getenv("RESPONSE_LM_API_URL")
	model := os.Getenv("RESPONSE_MODEL_NAME")
	temperature := os.Getenv("RESPONSE_LM_TEMPERATURE")
	maxTokens := os.Getenv("RESPONSE_LM_MAX_TOKENS")
	topK := os.Getenv("RESPONSE_LM_TOP_K")
	topP := os.Getenv("RESPONSE_LM_TOP_P")
	minP := os.Getenv("RESPONSE_LM_MIN_P")

	payload := RequestPayload{
		Model:       model,
		Prompt:      prompt,
		Temperature: parseFloat(temperature),
		MaxTokens:   parseInt(maxTokens),
		TopK:        parseInt(topK),
		TopP:        parseFloat(topP),
		MinP:        parseFloat(minP),
	}

	return makeRequest(apiURL, payload)
}

// RespondWithContext realiza una respuesta usando el contexto del modelo
func RespondWithContext(ctx mcp.ModelContext, prompt string) (string, error) {
	ctx.AddMessage("user", prompt)

	messages := make([]ChatMessage, 0)
	for _, msg := range ctx.GetMessages() {
		messages = append(messages, ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	chatRequest := LMChatRequest{
		Model:       os.Getenv("RESPONSE_MODEL_NAME"),
		Messages:    messages,
		Temperature: nil,
		MaxTokens:   nil,
		TopK:        nil,
		TopP:        nil,
		MinP:        nil,
	}

	body, err := json.Marshal(chatRequest)
	if err != nil {
		return "", err
	}

	fmt.Printf("Payload enviado: %s\n", string(body))

	resp, err := http.Post(os.Getenv("RESPONSE_LM_API_URL"), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error en la respuesta del servidor: %s\n", string(responseBody))
		return "", errors.New("error en la llamada al modelo LLM")
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseBody), nil
}

// makeRequest realiza la llamada HTTP al modelo LLM.
func makeRequest(apiURL string, payload RequestPayload) (string, error) {
	messages := []ChatMessage{
		{
			Role:    "user",
			Content: payload.Prompt,
		},
	}

	chatRequest := LMChatRequest{
		Model:       payload.Model,
		Messages:    messages,
		Temperature: &payload.Temperature,
		MaxTokens:   &payload.MaxTokens,
		TopK:        &payload.TopK,
		TopP:        &payload.TopP,
		MinP:        &payload.MinP,
	}

	body, err := json.Marshal(chatRequest)
	if err != nil {
		return "", err
	}

	fmt.Printf("Payload enviado: %s\n", string(body))

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error en la respuesta del servidor: %s\n", string(responseBody))
		return "", errors.New("error en la llamada al modelo LLM")
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseBody), nil
}

// parseFloat convierte una cadena a float32.
func parseFloat(value string) float32 {
	if f, err := strconv.ParseFloat(value, 32); err == nil {
		return float32(f)
	}
	return 0.0
}

// parseInt convierte una cadena a int.
func parseInt(value string) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return 0
}
