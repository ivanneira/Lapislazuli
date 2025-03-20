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

// Process realiza la clasificación usando el modelo clasificador.
func Process(prompt string) (string, error) {
	// Construir el system prompt con las acciones disponibles.
	systemContent := fmt.Sprintf(
		"Acciones disponibles: %s. Clasifica el siguiente prompt devolviendo un JSON con el campo 'action'.",
		strings.Join(config.Config.Actions, ", "),
	)

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

	// Crear el cuerpo de la solicitud.
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

	// Serializar el cuerpo de la solicitud a JSON.
	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Crear la solicitud HTTP.
	req, err := http.NewRequest("POST", config.Config.ClassificatorLMAPIURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.Config.ClassificatorAPIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Config.ClassificatorAPIKey))
	}

	// Enviar la solicitud.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Manejar errores de respuesta.
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error en la llamada a LM Studio: %s", string(bodyBytes))
	}

	// Leer la respuesta del servidor.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Deserializar la respuesta.
	var lmResp LMResponse
	if err := json.Unmarshal(respBody, &lmResp); err != nil {
		return "", err
	}

	// Verificar si se recibieron respuestas.
	if len(lmResp.Choices) == 0 {
		return "", fmt.Errorf("no se recibieron respuestas del modelo")
	}

	// Se espera que el mensaje devuelto contenga el JSON con el campo "action".
	return lmResp.Choices[0].Message.Content, nil
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

// makeRequest realiza la llamada HTTP al modelo LLM.
func makeRequest(apiURL string, payload RequestPayload) (string, error) {
	// Crear el campo "messages" requerido por el servidor
	messages := []ChatMessage{
		{
			Role:    "user", // El rol puede ser "user" para el mensaje del usuario
			Content: payload.Prompt,
		},
	}

	// Crear el nuevo payload con el campo "messages"
	chatRequest := LMChatRequest{
		Model:       payload.Model,
		Messages:    messages,
		Temperature: &payload.Temperature,
		MaxTokens:   &payload.MaxTokens,
		TopK:        &payload.TopK,
		TopP:        &payload.TopP,
		MinP:        &payload.MinP,
	}

	// Serializar el nuevo payload a JSON
	body, err := json.Marshal(chatRequest)
	if err != nil {
		return "", err
	}

	// Log para depurar el payload
	fmt.Printf("Payload enviado: %s\n", string(body))

	// Realizar la solicitud HTTP
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Manejar errores de respuesta
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Error en la respuesta del servidor: %s\n", string(responseBody))
		return "", errors.New("error en la llamada al modelo LLM")
	}

	// Leer la respuesta del servidor
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
