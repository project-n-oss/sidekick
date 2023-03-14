package boltrouter

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ZONE_ID", "use1-az1")
	os.Setenv("BOLT_CUSTOM_DOMAIN", "test.bolt.projectn.co")

	exitVal := m.Run()
	os.Exit(exitVal)
}
