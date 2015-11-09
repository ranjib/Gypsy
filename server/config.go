package server

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
