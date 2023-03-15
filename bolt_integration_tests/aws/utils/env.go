package utils

import (
	"log"
	"os"
	"strconv"
)

func GetEnv(key string, defaultVal int) int {
	var err error
	out := defaultVal
	val := os.Getenv(key)
	if len(val) > 0 {
		out, err = strconv.Atoi(val)
		if err != nil {
			log.Fatalf("invalid env val %v for key %v", val, key)
		}
	}
	return out
}

func GetEnvStr(key, defaultVal string) string {
	val := os.Getenv(key)
	if len(val) > 0 {
		return val
	}
	return defaultVal
}

func GetEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)

	if val == "true" || val == "1" {
		return true
	}

	if val == "false" || val == "0" {
		return false
	}

	if len(val) > 0 {
		log.Fatalf("unhandled val %v for key %v", val, key)
	}

	return defaultVal
}
