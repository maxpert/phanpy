package core

import (
	"errors"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

type configRoot struct {
	NamedQueries []namedQueryConfig `hcl:"query,block"`
}

type namedQueryConfig struct {
	Name    string `yaml:"name" hcl:"name,label"`
	SQL     string `yaml:"sql" hcl:"sql"`
	Timeout int32  `yaml:"timeout" hcl:"timeout"`
}

func loadHCLConfig(path string) ([]namedQueryConfig, error) {
	var cfg configRoot
	if err := hclsimple.DecodeFile(path, nil, &cfg); err != nil {
		return nil, err
	}

	return cfg.NamedQueries, nil
}

func loadYAMLConfig(path string) ([]namedQueryConfig, error) {
	cfg := make([]namedQueryConfig, 0)
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

func loadQueryConfig(path string) ([]namedQueryConfig, error) {
	if len(path) == 0 {
		path = "queries.yml"
	}

	if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
		return loadYAMLConfig(path)
	}

	if strings.HasSuffix(path, ".hcl") || strings.HasSuffix(path, ".json") {
		return loadHCLConfig(path)
	}

	return nil, errors.New("invalid file type")
}

func prepareQueryConfigMap(cfg []namedQueryConfig) map[string]namedQueryConfig {
	ret := make(map[string]namedQueryConfig)
	for _, c := range cfg {
		ret[c.Name] = c
	}

	return ret
}
