package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func WithFilePath(filePath string) func(*UnmarshalConfigOptions) {
	return func(options *UnmarshalConfigOptions) {
		options.filePath = filePath
	}
}

// UnmarshalConfigFromFile populates config with values from a yaml file.
// It will by default look for a file named config.yml in the current working directory.
// You can override this by passing the WithFilePath option.
func UnmarshalConfigFromFile(config any, opts ...func(*UnmarshalConfigOptions)) error {
	options := UnmarshalConfigOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	if options.filePath != "" {
		f, err := os.Open(options.filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := yaml.NewDecoder(f).Decode(config); err != nil {
			return fmt.Errorf("failed to decode config: %w", err)
		}
	} else if f, err := os.Open("config.yml"); err == nil {
		defer f.Close()
		if err := yaml.NewDecoder(f).Decode(config); err != nil {
			return fmt.Errorf("error decoding config: %w", err)
		}
	}

	return nil
}
