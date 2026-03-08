package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const configFileName = ".gatorconfig.json"

type Config struct{
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(name string, id uuid.UUID) error{
	c.CurrentUserName = name
	err := write(*c)
	return err
}

func Read() Config{
	homeDir, err := os.UserHomeDir()
	if err != nil{
		fmt.Println("Error getting home dir:", err)
		return Config{}
	}
	data, err := os.ReadFile(filepath.Join(homeDir, configFileName))
	if err != nil{
		fmt.Println("Error reading file:", err)
		return Config{}
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil{
		fmt.Println("Error reading file:", err)
		return Config{}
	}
	return  config
}

func write(cfg Config) error{
	data, err := json.Marshal(cfg)
	if err != nil{
		return fmt.Errorf("Error marshalling config: %w", err)
		
	}
	homeDir, err := os.UserHomeDir()
	if err != nil{
		return fmt.Errorf("Error getting home dir: %w", err)
	}
	err = os.WriteFile(filepath.Join(homeDir, configFileName), data, 0644)
	if err != nil{
		return fmt.Errorf("Error writing file: %w", err) 
	}
	return nil
}