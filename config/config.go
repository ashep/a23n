package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Database struct {
	DSN string `yaml:"dsn"`
}

type Config struct {
	DB              Database `yaml:"db"`
	Address         string   `yaml:"address"`
	Secret          string   `yaml:"secret"`
	AccessTokenTTL  uint     `yaml:"access_token_ttl"`
	RefreshTokenTTL uint     `yaml:"refresh_token_ttl"`
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
