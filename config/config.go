package config

import "flag"

var (
	dataDir = flag.String("data-dir", "./data", "Path to the data directory.")
)

var (
	config *Config
)

func Get() *Config {
	if config == nil {
		flag.Parse()
	}

	config = &Config{
		dataDir: *dataDir,
	}

	return config
}

type Config struct {
	dataDir string
}

func (c *Config) DataDir() string {
	return c.dataDir
}
