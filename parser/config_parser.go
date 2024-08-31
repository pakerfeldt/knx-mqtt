// config_parser.go
package parser

import (
	"fmt"
	"os"

	"github.com/pakerfeldt/knx-mqtt/models"
	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses the YAML configuration file.
func LoadConfig(filePath string) (*models.Config, error) {
	// Open the YAML file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a decoder
	decoder := yaml.NewDecoder(file)

	// Decode the YAML file into the Config struct
	var config models.Config
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding YAML: %w", err)
	}

	return &config, nil
}
