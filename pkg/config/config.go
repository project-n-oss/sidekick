package config

import (
	"context"
)

type UnmarshalConfigOptions struct {
	filePath string
}

// UnmarshalConfig populates config with values from environment variables and a yaml file.
func UnmarshalConfig(ctx context.Context, prefix string, config any, opts ...func(*UnmarshalConfigOptions)) error {
	err := UnmarshalConfigFromFile(config, opts...)
	if err != nil {
		return err
	}

	err = UnmarshalConfigFromEnv(ctx, prefix, config)
	if err != nil {
		return err
	}

	return nil
}
