package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Props struct {
	Spanner Spanner `yaml:"spanner"`
	PubSub  PubSub  `yaml:"pubsub"`
	Server  Server  `yaml:"server"`
}

type Server struct {
	ShutdownDelay int32 `yaml:"shutdown_delay"`
}

type PubSub struct {
	ProjectID       string `yaml:"project_id"`
	CredentialsFile string `yaml:"credentials_file"`
	EmulatorEnabled bool   `yaml:"emulator_enabled"`
	EmulatorHost    string `yaml:"emulator_host"`
	SubscriptionID  string `yaml:"subscription_id"`
}

type Spanner struct {
	ProjectID       string `yaml:"project_id"`
	InstanceID      string `yaml:"instance_id"`
	DatabaseID      string `yaml:"database_id"`
	CredentialsFile string `yaml:"credentials_file"`
	EmulatorEnabled bool   `yaml:"emulator_enabled"`
	EmulatorHost    string `yaml:"emulator_host"`
}

func LoadConfig() (*Props, error) {
	propFile := "config.yaml"
	if profile := os.Getenv("GO_PROFILE"); profile != "" {
		propFile = fmt.Sprintf("config-%s.yaml", profile)
	}
	data, err := os.ReadFile(propFile)
	if err != nil {
		return nil, err
	}

	props := Props{}
	if err := yaml.Unmarshal(data, &props); err != nil {
		return nil, err
	}
	return &props, nil
}
