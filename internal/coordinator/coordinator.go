package coordinator

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/lapislazuli/internal/models"
	"github.com/yourusername/lapislazuli/internal/processor"
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
func (a *Agent) ProcessRequest(ctx context.Context, request *Request) (*Response, error) {
	a.logger.WithFields(logrus.Fields{
		"request": request.Text,
	}).Info("Procesando solicitud")

	// Paso 1: Clasificar la solicitud
	classification, err := a.processorAgent.ClassifyRequest(ctx, request.Text)
	if err != nil {
		a.logger.WithError(err).Error("Error al clasificar solicitud")
		return &Response{
			Text: "Error al procesar la solicitud",
		}, err
	}

	// Por ahora, solo devolvemos la clasificación
	// En versiones futuras, aquí iría la lógica completa del flujo
	return &Response{
		Text:           "Solicitud clasificada como: " + classification.Intent,
		Classification: classification,
	}, nil
}
