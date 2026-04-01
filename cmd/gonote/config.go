package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Difficulty int

const (
	OnlyNotesWithoutSign Difficulty = iota + 1
	AllNotes
)

type Config struct {
	ScoreToWin int        `yaml:"score_to_win"`
	Difficulty Difficulty `yaml:"difficulty"`
	Port       int        `yaml:"default_port"`
}

func DefaultConfig() *Config {
	return &Config{
		ScoreToWin: 10,
		Difficulty: AllNotes,
		Port:       9090,
	}
}

type YamlStruct struct {
	ScoreToWin  int        `yaml:"score_to_win"`
	Difficulty  Difficulty `yaml:"difficulty"`
	DefaultPort int        `yaml:"default_port"`
}

func ParseConfig() (*Config, error) {

	cfg := DefaultConfig()

	data, err := os.ReadFile("../../config.yaml")
	if err == nil {
		var yamlCfg YamlStruct
		if err := yaml.Unmarshal(data, &yamlCfg); err != nil {
			return nil, err
		}
		cfg.ScoreToWin = yamlCfg.ScoreToWin
		cfg.Difficulty = yamlCfg.Difficulty
		cfg.Port = yamlCfg.DefaultPort
	}

	_ = godotenv.Load()

	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}

	return cfg, nil
}
