package coordinator

import (
	"encoding/json"
	"fmt"

	"github.com/ivanneira/Lapislazuli/internal/processor"
)

// ClassificationResult define la estructura de la respuesta JSON.
type ClassificationResult struct {
	Action string `json:"action"`
}

// HandlePrompt recibe el prompt, llama al processor y muestra en consola la acción clasificada.
func HandlePrompt(prompt string) error {
	resultJSON, err := processor.Process(prompt)
	if err != nil {
		return err
	}

	var result ClassificationResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return err
	}

	fmt.Printf("Acción clasificada: %s\n", result.Action)
	return nil
}
