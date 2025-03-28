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

// Función genérica para obtener valores del entorno
func getEnvValue[T any](key string, parser func(string) (T, error), defaultValue T) T {
	if val := os.Getenv(key); val != "" {
		if parsed, err := parser(val); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Parsers específicos
func parseFloat32(s string) (float32, error) {
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

// LoadConfig carga la configuración desde el archivo .env.
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró el archivo .env")
	}

	Config.ServerURL = os.Getenv("SERVER_URL")
	Config.ClassificatorAPIKey = os.Getenv("CLASSIFICATOR_API_KEY")
	Config.ClassificatorModelName = os.Getenv("CLASSIFICATOR_MODEL_NAME")
	Config.ClassificatorLMAPIURL = os.Getenv("CLASSIFICATOR_LM_API_URL")

	if actions := os.Getenv("ACTIONS"); actions != "" {
		Config.Actions = strings.Split(actions, ",")
	} else {
		Config.Actions = []string{"llamada", "mensaje", "correo"}
	}

	Config.ClassificatorTemperature = getEnvValue("CLASSIFICATOR_LM_TEMPERATURE", parseFloat32, -1)
	Config.ClassificatorMaxTokens = getEnvValue("CLASSIFICATOR_LM_MAX_TOKENS", strconv.Atoi, -1)
	Config.ClassificatorTopK = getEnvValue("CLASSIFICATOR_LM_TOP_K", strconv.Atoi, -1)
	Config.ClassificatorTopP = getEnvValue("CLASSIFICATOR_LM_TOP_P", parseFloat32, -1)
	Config.ClassificatorMinP = getEnvValue("CLASSIFICATOR_LM_MIN_P", parseFloat32, -1)
	Config.ClassificatorRepetitionPenalty = getEnvValue("CLASSIFICATOR_LM_REPETITION_PENALTY", parseFloat32, -1)
}
