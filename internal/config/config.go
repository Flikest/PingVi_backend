package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MinIOClient MinIOClientConfig `yaml:"minio_client,omitempty"`
}
type MinIOClientConfig struct {
	Endpoint string `yaml:"endpoint,omitempty"`
	UseSSL   bool   `yaml:"use_ssl,omitempty"`
}

func ParseConfig(log *slog.Logger, env, configPath string) (Config, error) {
	var config Config

	fileNameWithExt := filepath.Base(configPath)

	ext := filepath.Ext(fileNameWithExt)

	fileName := strings.TrimSuffix(fileNameWithExt, ext)

	if fileName == "file_storage" {
		file, err := os.ReadFile(configPath)
		if err != nil {
			log.Error("error with reading yaml file: ", "error", err)
			return Config{}, err
		}

		if err = yaml.Unmarshal(file, &config); err != nil {
			log.Error("error with unmarshling yaml file: ", "error", err)
			return Config{}, err
		}
		return config, nil
	}

	if err := godotenv.Load(fmt.Sprintf("./%s.%s.env", fileName, env)); err != nil {
		log.Error("error with loading .env variable: ", "error", err)
		return Config{}, err
	}
	return Config{}, nil
}
