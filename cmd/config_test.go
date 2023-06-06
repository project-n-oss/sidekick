package cmd

import (
	"reflect"
	"testing"

	"github.com/project-n-oss/sidekick/boltrouter"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalConfig(t *testing.T) {
	for name, tc := range map[string]struct {
		Environment map[string]string
		In          Config
		Expected    *Config
	}{
		"EnvOnly": {
			Environment: map[string]string{
				"TEST_BOLTROUTER_FAILOVER": "true",
			},
			Expected: &Config{
				BoltRouter: boltrouter.Config{
					Failover: true,
				},
			},
		},
		"Override": {
			Environment: map[string]string{
				"TEST_BOLTROUTER_FAILOVER": "false",
			},
			In: Config{
				BoltRouter: boltrouter.Config{
					Failover: true,
				},
			},
			Expected: &Config{
				BoltRouter: boltrouter.Config{
					Failover: false,
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			config := tc.In
			_, err := unmarshalConfig("TEST", reflect.ValueOf(&config), func(key string) (*string, error) {
				if v, ok := tc.Environment[key]; ok {
					return &v, nil
				}
				return nil, nil
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, &config)
		})
	}
}
