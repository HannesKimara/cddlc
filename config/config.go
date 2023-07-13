package config

import (
	"errors"
	"os"
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
	SourceDir string `json:"src" yaml:"src"`
	OutDir    string `json:"output" yaml:"output"`
}

// Config contains configuration values for the cddlc tool
type Config struct {
	Version string                  `json:"version" yaml:"version"`
	Plugins []Plugin                `json:"plugins" yaml:"plugins"`
	Builds  map[string]*BuildConfig `json:"builds" yaml:"builds"`
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
	builds := make(map[string]*BuildConfig)
	builds["golang"] = newDefaultBuildConfig()

	return &Config{
		Version: "1",
		Builds:  builds,
		Plugins: []Plugin{},
	}
}

func newDefaultBuildConfig() *BuildConfig {
	return &BuildConfig{
		SourceDir: ".",
		OutDir:    "dist/",
	}
}
