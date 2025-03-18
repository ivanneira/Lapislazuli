package processor

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/lapislazuli/internal/models"
)

// Agent implementación del Agente Procesador
type Agent struct {
	llmClient *models.LMStudioClient
	logger    *logrus.Logger
}

// NewAgent crea una nueva instancia del Agente Procesador
func NewAgent(llmClient *models.LMStudioClient, logger *logrus.Logger) *Agent {
	return &Agent{
		llmClient: llmClient,
		logger:    logger,
	}
}

// ClassifyRequest clasifica una solicitud de usuario
func (a *Agent) ClassifyRequest(ctx context.Context, input string) (*models.Classification, error) {
	a.logger.WithFields(logrus.Fields{
		"input": input,
	}).Info("Clasificando solicitud")

	classification, err := a.llmClient.Classify(ctx, input)
	if err != nil {
		a.logger.WithError(err).Error("Error al clasificar solicitud")
		return nil, err
	}

	a.logger.WithFields(logrus.Fields{
		"intent":        classification.Intent,
		"confidence":    classification.Confidence,
		"requires_data": classification.RequiresData,
	}).Info("Solicitud clasificada")

	return classification, nil
}
