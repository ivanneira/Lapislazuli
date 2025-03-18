package config

import (
	"github.com/spf13/viper"
)

// Config estructura de configuración global
type Config struct {
	LMStudio LMStudioConfig
	Server   ServerConfig
}

// LMStudioConfig configuración de conexión a LMStudio
type LMStudioConfig struct {
	Endpoint        string
	Model           string
	ClassifierModel string
}

// ServerConfig configuración del servidor
type ServerConfig struct {
	Port int
}

// LoadConfig carga la configuración desde archivo o variables de entorno
func LoadConfig() (*Config, error) {
	viper.SetDefault("lmstudio.endpoint", "http://192.168.1.40:1234")
	viper.SetDefault("lmstudio.model", "gemma-3-4b-it")
	viper.SetDefault("lmstudio.classifier_model", "gemma-3-1b-it@q8_0")
	viper.SetDefault("server.port", 8080)

	viper.AutomaticEnv()

	return &Config{
		LMStudio: LMStudioConfig{
			Endpoint:        viper.GetString("lmstudio.endpoint"),
			Model:           viper.GetString("lmstudio.model"),
			ClassifierModel: viper.GetString("lmstudio.classifier_model"),
		},
		Server: ServerConfig{
			Port: viper.GetInt("server.port"),
		},
	}, nil
}
