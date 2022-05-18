package utilconfig

import (
	"errors"
	"gopkg.in/yaml.v2"
	"os"
)

var (
	ErrMissedEnvConfig = errors.New("missed env var for a YAML config file")
)

func LoadYAMLFromEnvPath(cfg interface{}, envVar string) error {
	filename, ok := os.LookupEnv(envVar)
	if !ok {
		return ErrMissedEnvConfig
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return err
	}
	return nil
}
