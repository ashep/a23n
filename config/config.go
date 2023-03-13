package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type API struct {
	TokenTTL int `yaml:"token_ttl"`
}

type Database struct {
	DSN string `yaml:"dsn"`
}

type Server struct {
	Address string `yaml:"address"`
	Secret  string `yaml:"secret"`
}

type Config struct {
	API    API      `yaml:"api"`
	DB     Database `yaml:"db"`
	Server Server   `yaml:"server"`
}

func Parse(in []byte) (Config, error) {
	r := Config{}
	err := yaml.Unmarshal(in, &r)
	if err != nil {
		return r, err
	}

	return r, nil
}

func ParseFromPath(path string) (Config, error) {
	fp, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer fp.Close()

	b, err := io.ReadAll(fp)
	if err != nil {
		return Config{}, err
	}

	return Parse(b)
}
