package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	EnvPath    string           `yaml:"env_path"`
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	GRPCServer GRPCServerConfig `yaml:"grpc_server"`
}

type HTTPServerConfig struct {
	Host    string `yaml:"host"`
	Timeout int    `yaml:"timeout"`
}

type GRPCServerConfig struct {
	Port    int    `yaml:"port"`
	Host    string `yaml:"host"`
	Timeout int    `yaml:"timeout"`
}

func ParseConfig(log *slog.Logger, configPath string) (Config, error) {
	var config Config

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Error("error with reading yaml file: ", "error", err)
		return Config{}, err
	}

	if err = yaml.Unmarshal(file, &config); err != nil {
		log.Error("error with unmarshling yaml file: ", "error", err)
		return Config{}, err
	}

	if err := godotenv.Load(config.EnvPath); err != nil {
		log.Error("error with loading .env variable: ", "error", err)
		return Config{}, err
	}

	return config, nil
}
