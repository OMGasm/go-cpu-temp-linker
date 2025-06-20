package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Cpu_sensor           string
	Cpu_input_label      string
	Resolve_temp_hw_path bool
	Link_path            string
}

func Read_config() (*Config, error) {
	file, err := os.Open("config.json")
	defer file.Close()
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(file)
	var cfg Config
	err = dec.Decode(&cfg)
	return &cfg, err
}

func Default_Config() *Config {
	return &Config{
		Cpu_sensor:           "k10temp",
		Cpu_input_label:      "Tctl",
		Resolve_temp_hw_path: false,
		Link_path:            "temp_input",
	}
}
