package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type SecretValue struct {
	Value string `json:"value"`
}

// UnmarshalConfigFromEnv populates config with values from environment variables. The names of the
// environment variables read are formed by joining the config struct's field names with underscores
// and adding a user defined prefix. for example consider this config struct:
//
//	type Config struct {
//		Foo    string `mapstructure:"Foo"`
//		SubFoo SubFoo `mapstructure:"SubFoo"`
//	}
//
//	type SubFoo struct {
//		SubBar        int  `mapstructure:"SubBar"`
//		PointerSubBar *int `mapstructure:"PointerBar"`
//	}
//
// If the prefix is "TEST", the following environment variables will be read:
// TEST_FOO
// TEST_SUBFOO_SUBBAR
// TEST_SUBFOO_POINTERBAR
func UnmarshalConfigFromEnv(ctx context.Context, prefix string, config any) error {
	_, err := unmarshalConfig(prefix, reflect.ValueOf(config), func(key string) (*string, error) {
		v, ok := os.LookupEnv(key)
		if !ok {
			return nil, nil
		}
		return &v, nil
	})
	return err
}

func unmarshalConfig(prefix string, v reflect.Value, lookup func(string) (*string, error)) (didUnmarshal bool, err error) {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		if env, err := lookup(prefix); err != nil {
			return false, err
		} else if env != nil {
			switch dest := v.Interface().(type) {
			case *bool:
				switch *env {
				case "false":
					*dest = false
				case "true":
					*dest = true
				default:
					return false, fmt.Errorf("boolean config must be \"true\" or \"false\"")
				}
			case *int:
				n, err := strconv.Atoi(*env)
				if err != nil {
					return false, fmt.Errorf("invalid value for integer config")
				}
				*dest = n
			case *string:
				*dest = *env
			case *[]byte:
				buf, err := base64.StdEncoding.DecodeString(*env)
				if err != nil {
					return false, fmt.Errorf("byte slice configs must be base64 encoded")
				}
				*dest = buf
			case *[]int:
				parts := strings.Split(*env, ",")
				intParts := make([]int, len(parts))
				for i, part := range parts {
					intParts[i], err = strconv.Atoi(strings.TrimSpace(part))
					if err != nil {
						return false, fmt.Errorf("invalid value for integer config")
					}
				}
				*dest = intParts
			case *[]string:
				parts := strings.Split(*env, ",")
				for i, part := range parts {
					parts[i] = strings.TrimSpace(part)
				}
				*dest = parts
			default:
				return false, fmt.Errorf("unsupported environment config type %T", v.Elem().Interface())
			}
			return true, nil
		}
	}

	if v.Kind() == reflect.Ptr {
		if !v.IsNil() {
			return unmarshalConfig(prefix, v.Elem(), lookup)
		}

		new := reflect.New(v.Type().Elem())
		didUnmarshal, err := unmarshalConfig(prefix, new, lookup)
		if err != nil {
			return false, err
		} else if didUnmarshal {
			v.Set(new)
		}
		return didUnmarshal, nil
	}

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			didUnmarshalField, err := unmarshalConfig(prefix+"_"+strings.ToUpper(t.Field(i).Name), v.Field(i).Addr(), lookup)
			if err != nil {
				return false, err
			}
			if didUnmarshalField {
				didUnmarshal = true
			}
		}
	}

	return didUnmarshal, nil
}
