package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Conf from config yml
type Conf struct {
	Mongo  *MongoConfig  `yaml:"mongo"`
	System *SystemConfig `yaml:"system"`
}

// MongoDB parameters
type MongoConfig struct {
	URI                string `yaml:"uri"`
	Database           string `yaml:"database"`
	CollectionMessages string `yaml:"collection_messages"`
}

// System parameters
type SystemConfig struct {
	DataPath string `yaml:"data_path"`
}

// Load config from file
func Load(fname string) (res *Conf, err error) {
	res = &Conf{}
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, res); err != nil {
		return nil, err
	}

	return res, nil
}
