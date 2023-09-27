package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	cDDLC_YAML string = "cddlc.yaml"
	cDDLC_JSON string = "cddlc.json"
)

type PluginRemote struct {
	URL string `json:"url" yaml:"url"`
}

type PluginLocal struct {
	Path string `json:"path" yaml:"path"`
}

func (pl *PluginLocal) Valid() error {
	stat, err := os.Stat(pl.Path)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return errors.New("plugin path " + pl.Path + " not found")
	}
	return nil
}

type Plugin struct {
	Name   string       `json:"name" yaml:"name"`
	Remote PluginRemote `json:"remote" yaml:"remote"`
	Local  PluginLocal  `json:"local" yaml:"local"`
}

func (p *Plugin) Valid() error {
	err := p.Local.Valid()
	return err
}

type BuildConfig struct {
	Package   string `json:"package" yaml:"package"`
	SourceDir string `json:"src" yaml:"src"`
	OutDir    string `json:"output" yaml:"output"`

	// TODO: Implement overrides
	// Exclude   []string `json:"exclude" yaml:"exclude"`
}

// Config contains configuration values for the cddlc tool
type Config struct {
	Version string         `json:"version" yaml:"version"`
	Plugins []*Plugin      `json:"plugins" yaml:"plugins"`
	Builds  []*BuildConfig `json:"builds" yaml:"builds"`

	Options map[string]interface{} `json:"options" yaml:"options"`
}

func (c *Config) Valid() error {
	for _, plugin := range c.Plugins {
		err := plugin.Valid()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewDefaultConfig() *Config {
	builds := []*BuildConfig{newDefaultBuildConfig()}

	return &Config{
		Version: "1",
		Builds:  builds,
		Plugins: []*Plugin{},
	}
}

func newDefaultBuildConfig() *BuildConfig {
	return &BuildConfig{
		SourceDir: ".",
		OutDir:    "dist/",
	}
}

func LoadConfig(path string) (*Config, error) {
	base := filepath.Base(path)
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	switch base {
	case cDDLC_JSON:
		err = json.NewDecoder(fh).Decode(config)
	case cDDLC_YAML:
		err = yaml.NewDecoder(fh).Decode(config)
	default:
		err = errors.New("unknown config file base: " + base)
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}
