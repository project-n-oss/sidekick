package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	require.NotEmpty(t, version())
}
