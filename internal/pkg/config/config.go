package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Mqtt struct {
		Host            string `yaml:"host"`
		Port            string `yaml:"port"`
		Username        string `yaml:"username"`
		Password        string `yaml:"password"`
		DiscoveryPrefix string `yaml:"discoveryPrefix"`
	} `yaml:"server"`
	Healthbox struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"healthbox"`
}

func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
