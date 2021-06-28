package config

import (
	_ "embed"
	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var yamlConfig []byte
var cfg KeptnGoUtilsConfig

type KeptnGoUtilsConfig struct {
	ShKeptnSpecVersion string `yaml:"shkeptnspecversion"`
}

func GetKeptnGoUtilsConfig() (*KeptnGoUtilsConfig, error) {
	err := yaml.Unmarshal(yamlConfig, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
