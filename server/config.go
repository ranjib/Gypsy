package server

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	DataDir          string `yaml:"data_dir"`
	ArtifactDir      string `yaml:"artifact_dir"`
	BindAddr         string `yaml:"bind_addr"`
	PollingFrequency int    `yaml:"polling_frequency"`
}

func DefaultConfig() *Config {
	return &Config{
		DataDir:          "data",
		ArtifactDir:      "data/artifacts",
		BindAddr:         "127.0.0.1:5678",
		PollingFrequency: 300,
	}
}

func ConfigFomeFile(file string) (*Config, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read configuration file %s. Error: %s\n", file, err)
		return nil, err
	}
	config := DefaultConfig()
	if err := yaml.Unmarshal(content, config); err != nil {
		log.Errorf("Failed to desrialize configuration file %s. Error: %s\n", file, err)
		return nil, err
	}
	return config, nil
}
