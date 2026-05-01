package main

import (
	"os"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Difficulty int

const (
	OnlyNotesWithoutSign Difficulty = iota + 1
	AllNotes
)

type Config struct {
	ScoreToWin int
	Difficulty Difficulty
	Port       int
}

type envStruct struct {
	Port int `env:"GONOTE_SERVER_PORT" envDefault:"9090"`
}

func DefaultConfig() *Config {
	return &Config{
		ScoreToWin: 10,
		Difficulty: AllNotes,
		Port:       9090,
	}
}

type YamlStruct struct {
	ScoreToWin int        `yaml:"scoreToWin"`
	Difficulty Difficulty `yaml:"difficulty"`
}

func ParseConfig() (*Config, error) {
	cfg := DefaultConfig()

	_ = godotenv.Load()

	envData := new(envStruct)

	err := env.Parse(envData)
	if err != nil {
		return cfg, err
	}

	cfg.Port = envData.Port

	data, err := os.ReadFile("../../config.yaml")
	if err != nil {
		return cfg, err
	}

	yamlCfg := new(YamlStruct)
	if err := yaml.Unmarshal(data, yamlCfg); err != nil {
		return nil, err
	}

	if yamlCfg.ScoreToWin != 0 {
		cfg.ScoreToWin = yamlCfg.ScoreToWin
	}

	if yamlCfg.Difficulty != 0 {
		cfg.Difficulty = yamlCfg.Difficulty
	}

	return cfg, nil
}
