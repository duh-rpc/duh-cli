package lint

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Lint LintConfig `yaml:"lint"`
}

type LintConfig struct {
	Disable []string `yaml:"disable"`
}

func LoadConfig() Config {
	data, err := os.ReadFile(".duh.yaml")
	if err != nil {
		return Config{}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}
	}

	return cfg
}
