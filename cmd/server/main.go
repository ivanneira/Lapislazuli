package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/lapislazuli/config"
	"github.com/yourusername/lapislazuli/internal/coordinator"
	"github.com/yourusername/lapislazuli/internal/models"
	"github.com/yourusername/lapislazuli/internal/processor"
)

func main() {
	// Configurar logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("Error al cargar configuración")
	}

	// Inicializar clientes y agentes
	lmStudioClient := models.NewLMStudioClient(&cfg.LMStudio)
	processorAgent := processor.NewAgent(lmStudioClient, logger)
	coordinatorAgent := coordinator.NewAgent(processorAgent, logger)

	// Configurar router HTTP
	router := gin.Default()

	// Endpoint de clasificación
	router.POST("/classify", func(c *gin.Context) {
		var request coordinator.Request
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := coordinatorAgent.ProcessRequest(c.Request.Context(), &request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	})

	// Endpoint de salud
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Iniciar servidor
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.WithField("addr", addr).Info("Iniciando servidor HTTP")
	if err := router.Run(addr); err != nil {
		logger.WithError(err).Fatal("Error al iniciar servidor")
	}
}
