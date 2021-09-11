package core

import (
	"gopkg.in/yaml.v2"
	"os"
)

type namedQueryConfig struct {
	Name    string `yaml:"name"`
	Query   string `yaml:"query"`
	Timeout int32  `yaml:"timeout"`
}

func loadQueryConfig(path string) ([]namedQueryConfig, error) {
	cfg := make([]namedQueryConfig, 0)
	if len(path) == 0 {
		path = "queries.yml"
	}

	f, err := os.Open(path)
	if err == os.ErrNotExist {
		return cfg, nil
	}

	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func prepareQueryConfigMap(cfg []namedQueryConfig) map[string]namedQueryConfig {
	ret := make(map[string]namedQueryConfig)
	for _, c := range cfg {
		ret[c.Name] = c
	}

	return ret
}
