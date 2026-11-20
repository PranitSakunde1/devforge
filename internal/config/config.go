package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Stacks  struct {
		Go     bool `yaml:"go"`
		Java   bool `yaml:"java"`
		Python bool `yaml:"python"`
		Redis  bool `yaml:"redis"`
	} `yaml:"stacks"`
	Reports struct {
		Directory string `yaml:"directory"`
	} `yaml:"reports"`
}

func Default(name string) Config {
	c := Config{Name: name, Version: "1"}
	c.Stacks.Go, c.Stacks.Java, c.Stacks.Python, c.Stacks.Redis = true, true, true, true
	c.Reports.Directory = "reports"
	return c
}
func Save(dir string, c Config) error {
	b, e := yaml.Marshal(&c)
	if e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(dir, "devforge.yaml"), b, 0644)
}
func Load(dir string) (Config, error) {
	b, e := os.ReadFile(filepath.Join(dir, "devforge.yaml"))
	if e != nil {
		return Config{}, e
	}
	var c Config
	if e = yaml.Unmarshal(b, &c); e != nil {
		return Config{}, fmt.Errorf("parse devforge.yaml: %w", e)
	}
	if c.Name == "" {
		return Config{}, fmt.Errorf("devforge.yaml: name is required")
	}
	return c, nil
}
