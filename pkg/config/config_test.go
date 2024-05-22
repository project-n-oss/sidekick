package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Config struct {
	Foo    string `yaml:"Foo"`
	SubFoo SubFoo `yaml:"SubFoo"`
}

type SubFoo struct {
	SubBar        int `yaml:"SubBar"`
	PointerSubBar int `yaml:"PointerSubBar"`
}

func TestUnmarshalConfig(t *testing.T) {
	for name, tc := range map[string]struct {
		Environment map[string]string
		In          Config
		Expected    *Config
	}{
		"EnvOnly": {
			Environment: map[string]string{
				"TEST_FOO":                  "foo",
				"TEST_SUBFOO_SUBBAR":        "1",
				"TEST_SUBFOO_POINTERSUBBAR": "2",
			},
			Expected: &Config{
				Foo: "foo",
				SubFoo: SubFoo{
					SubBar:        1,
					PointerSubBar: 2,
				},
			},
		},
		"Override": {
			Environment: map[string]string{
				"TEST_FOO": "foo",
			},
			In: Config{
				Foo: "bar",
			},
			Expected: &Config{
				Foo: "foo",
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
