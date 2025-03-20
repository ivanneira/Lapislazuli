package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// ConfigStruct almacena las variables de entorno.
type ConfigStruct struct {
	ServerURL         string
	APIKey            string
	ModelName         string
	LMAPIURL          string
	Actions           []string
	Temperature       float32
	MaxTokens         int
	TopK              int
	TopP              float32
	MinP              float32
	RepetitionPenalty float32
}

var Config ConfigStruct

// LoadConfig carga la configuración desde el archivo .env.
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró el archivo .env")
	}

	Config.ServerURL = os.Getenv("SERVER_URL")
	Config.APIKey = os.Getenv("API_KEY")
	Config.ModelName = os.Getenv("MODEL_NAME")
	Config.LMAPIURL = os.Getenv("LM_API_URL")

	actions := os.Getenv("ACTIONS")
	if actions != "" {
		Config.Actions = strings.Split(actions, ",")
	} else {
		Config.Actions = []string{"llamada", "mensaje", "correo"}
	}

	// Función auxiliar para parsear valores numéricos; si hay error o es vacío se asigna -1.
	parseFloatOrDefault := func(key string) float32 {
		val := os.Getenv(key)
		if val != "" {
			if f, err := strconv.ParseFloat(val, 32); err == nil {
				return float32(f)
			}
		}
		return -1
	}
	parseIntOrDefault := func(key string) int {
		val := os.Getenv(key)
		if val != "" {
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
		return -1
	}

	Config.Temperature = parseFloatOrDefault("LM_TEMPERATURE")
	Config.MaxTokens = parseIntOrDefault("LM_MAX_TOKENS")
	Config.TopK = parseIntOrDefault("LM_TOP_K")
	Config.TopP = parseFloatOrDefault("LM_TOP_P")
	Config.MinP = parseFloatOrDefault("LM_MIN_P")
	Config.RepetitionPenalty = parseFloatOrDefault("LM_REPETITION_PENALTY")
}
