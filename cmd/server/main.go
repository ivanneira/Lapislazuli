package main

import (
	"net/http"

	"github.com/ivanneira/Lapislazuli/config"
	"github.com/ivanneira/Lapislazuli/internal/coordinator"

	"github.com/gin-gonic/gin"
)

// RequestPayload representa el JSON de entrada.
type RequestPayload struct {
	Text string `json:"text"`
}

func main() {
	// Cargar configuración desde .env
	config.LoadConfig()

	router := gin.Default()
	// Ruta configurada como /index
	router.POST("/index", func(c *gin.Context) {
		var payload RequestPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Llama al coordinador para procesar el prompt
		if err := coordinator.HandlePrompt(payload.Text); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "procesado"})
	})

	router.Run(":8080")
}
