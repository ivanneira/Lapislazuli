package coordinator

import (
	"context"
	"strings"

	"github.com/ivanneira/Lapislazuli/internal/models"
	"github.com/ivanneira/Lapislazuli/internal/processor"
	"github.com/sirupsen/logrus"
)

// Request estructura de solicitud al coordinador
type Request struct {
	Text string `json:"text"`
}

// Response estructura de respuesta del coordinador
type Response struct {
	Text           string                 `json:"text"`
	Classification *models.Classification `json:"classification"`
}

// Agent implementación del Agente Coordinador
type Agent struct {
	processorAgent *processor.Agent
	logger         *logrus.Logger
}

// NewAgent crea una nueva instancia del Agente Coordinador
func NewAgent(processorAgent *processor.Agent, logger *logrus.Logger) *Agent {
	return &Agent{
		processorAgent: processorAgent,
		logger:         logger,
	}
}

// ProcessRequest procesa una solicitud completa
func (a *Agent) ProcessRequest(ctx context.Context, request *Request) (map[string]interface{}, error) {
	a.logger.WithFields(logrus.Fields{
		"request": request.Text,
	}).Info("Procesando solicitud")

	// Paso 1: Clasificar la solicitud
	classification, err := a.processorAgent.ClassifyRequest(ctx, request.Text)
	if err != nil {
		a.logger.WithError(err).Error("Error al clasificar solicitud")
		return nil, err
	}

	// Paso 2: Extraer palabras clave del texto
	keywords := extractKeywords(request.Text)

	// Paso 3: Estructurar la respuesta en el formato solicitado
	response := map[string]interface{}{
		"accion":   classification.Intent,
		"keywords": keywords,
	}

	return response, nil
}

// Función auxiliar para extraer palabras clave
func extractKeywords(text string) []string {
	// Lógica simple para extraer palabras clave (puede mejorarse con NLP)
	words := strings.Fields(text)
	keywords := []string{}
	for _, word := range words {
		if len(word) > 3 && !strings.Contains("quiero enviar leer llamar crear cancelar", word) {
			keywords = append(keywords, word)
		}
	}
	return keywords
}
