package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Mqtt struct {
		Host            string `yaml:"Host"`
		Port            string `yaml:"Port"`
		Username        string `yaml:"Username"`
		Password        string `yaml:"Password"`
		DiscoveryPrefix string `yaml:"DiscoveryPrefix"`
	} `yaml:"Mqtt"`
	Healthbox struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"Healthbox"`
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
