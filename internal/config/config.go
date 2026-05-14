package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {

	configlocation, err := getConfigFilePath()
	if err != nil {
		fmt.Printf("Error reading config file %s", err)
	}

	configfile, err := os.ReadFile(configlocation)

	if err != nil {
		fmt.Println("Error reading config file:", err)
	}

	var config Config
	if err := json.Unmarshal(configfile, &config); err != nil {
		fmt.Println("Error unmarshaling:", err)
	}

	return config, err
}

func (cfg Config) SetUser(name string) error {
	cfg.CurrentUserName = name
	err := write(cfg)

	return err
}

func write(cfg Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		fmt.Printf("Error getting config file %s", err)
	}

	updatedConfigJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("Error marhaling json %s", err)
	}

	err = os.WriteFile(path, updatedConfigJSON, 0644)
	if err != nil {
		fmt.Printf("Error writing to file %s", err)
	}
	return err
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	configlocation := home + "/" + configFileName
	return configlocation, err
}
