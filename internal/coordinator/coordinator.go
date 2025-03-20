package coordinator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ivanneira/Lapislazuli/internal/processor"
)

// ClassificationResult define la estructura de la respuesta JSON.
type ClassificationResult struct {
	Action string `json:"action"`
}

// ExecutableResponse define la estructura de la respuesta JSON del ejecutable.
type ExecutableResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// HandlePrompt recibe el prompt, llama al processor y ejecuta la acción clasificada.
func HandlePrompt(prompt string) error {
	// Llamar al modelo clasificador
	resultJSON, err := processor.Process(prompt)
	if err != nil {
		return err
	}

	var result ClassificationResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return err
	}

	action := result.Action
	fmt.Printf("Acción clasificada: %s\n", action)

	// Verificar si la acción está definida en el archivo .env
	actions := os.Getenv("ACTIONS")
	if !isValidAction(action, actions) {
		return fmt.Errorf("Acción no definida: %s", action)
	}

	// Ejecutar el archivo correspondiente en la carpeta actions
	actionPath := fmt.Sprintf("actions/%s.exe", action)
	if _, err := os.Stat(actionPath); os.IsNotExist(err) {
		return fmt.Errorf("Acción no definida: %s", action)
	}

	// Capturar la salida del ejecutable
	cmd := exec.Command(actionPath)
	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error al ejecutar la acción: %s", err)
	}

	// Leer y deserializar la salida del ejecutable
	var execResponse ExecutableResponse
	if err := json.Unmarshal(outBuffer.Bytes(), &execResponse); err != nil {
		return fmt.Errorf("Error al deserializar la respuesta del ejecutable: %s", err)
	}

	// Imprimir la respuesta del ejecutable en consola
	fmt.Printf("Respuesta del ejecutable: %s (Estado: %s)\n", execResponse.Message, execResponse.Status)

	return nil
}

// isValidAction verifica si la acción está en la lista de acciones permitidas.
func isValidAction(action, actions string) bool {
	for _, a := range strings.Split(actions, ",") {
		if a == action {
			return true
		}
	}
	return false
}
