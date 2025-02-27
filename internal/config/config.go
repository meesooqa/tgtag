package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Conf from config yml
type Conf struct {
	Mongo  *MongoConfig  `yaml:"mongo"`
	System *SystemConfig `yaml:"system"`
	Server *ServerConfig `yaml:"server"`
}

// MongoConfig is a set of parameters for MongoDB
type MongoConfig struct {
	URI                string `yaml:"uri"`
	Database           string `yaml:"database"`
	CollectionMessages string `yaml:"collection_messages"`
}

// SystemConfig is the configuration for App
type SystemConfig struct {
	DataPath string `yaml:"data_path"`
}

// ServerConfig is a configuration for the server
type ServerConfig struct {
	Port int `yaml:"port"`
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
