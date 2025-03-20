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
	ServerURL                      string
	ClassificatorAPIKey            string
	ClassificatorModelName         string
	ClassificatorLMAPIURL          string
	Actions                        []string
	ClassificatorTemperature       float32
	ClassificatorMaxTokens         int
	ClassificatorTopK              int
	ClassificatorTopP              float32
	ClassificatorMinP              float32
	ClassificatorRepetitionPenalty float32
}

var Config ConfigStruct

// LoadConfig carga la configuración desde el archivo .env.
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró el archivo .env")
	}

	Config.ServerURL = os.Getenv("SERVER_URL")
	Config.ClassificatorAPIKey = os.Getenv("CLASSIFICATOR_API_KEY")
	Config.ClassificatorModelName = os.Getenv("CLASSIFICATOR_MODEL_NAME")
	Config.ClassificatorLMAPIURL = os.Getenv("CLASSIFICATOR_LM_API_URL")

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

	Config.ClassificatorTemperature = parseFloatOrDefault("CLASSIFICATOR_LM_TEMPERATURE")
	Config.ClassificatorMaxTokens = parseIntOrDefault("CLASSIFICATOR_LM_MAX_TOKENS")
	Config.ClassificatorTopK = parseIntOrDefault("CLASSIFICATOR_LM_TOP_K")
	Config.ClassificatorTopP = parseFloatOrDefault("CLASSIFICATOR_LM_TOP_P")
	Config.ClassificatorMinP = parseFloatOrDefault("CLASSIFICATOR_LM_MIN_P")
	Config.ClassificatorRepetitionPenalty = parseFloatOrDefault("CLASSIFICATOR_LM_REPETITION_PENALTY")
}
