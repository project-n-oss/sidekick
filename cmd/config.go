package cmd

import (
	"github.com/project-n-oss/sidekick/api"
	"github.com/project-n-oss/sidekick/app"
)

const cfgEnvPrefix = "SIDEKICK"

type Config struct {
	Api api.Config `yaml:"Api"`
	App app.Config `yaml:"App"`
}

var DefaultConfig = Config{}
