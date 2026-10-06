package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr string `yaml:"listen_addr"`
	DatabaseURL string `yaml:"database_url"`
	ValkeyAddr string `yaml:"valkey_addr"`
	DataDir    string `yaml:"data_dir"`
}

func Load() (*Config, error) {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "/etc/chatix/config.yaml"
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("чтение файла конфига %s: %w", cfgPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("парсинг YAML: %w", err)
	}

	// Переопределение из переменных окружения (удобно для docker/systemd)
	if v := os.Getenv("CHATIX_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("CHATIX_DB_URL"); v != "" {
		cfg.DatabaseURL = v
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:8080"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "/var/lib/chatix"
	}

	return &cfg, nil
}