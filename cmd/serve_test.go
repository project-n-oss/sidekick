package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testServeCmd = &cobra.Command{
	Use: "test",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func TestServeCmdConfig(t *testing.T) {
	initServerFlags(testServeCmd)
	rootCmd.AddCommand(testServeCmd)

	t.Run("Default", func(t *testing.T) {
		rootCmd.SetArgs([]string{"test"})
		err := rootCmd.Execute()
		require.NoError(t, err)

		cfg, err := getBoltRouterConfig(testServeCmd)
		require.NoError(t, err)
		assert.Equal(t, false, cfg.Local)
		assert.Equal(t, false, cfg.Passthrough)
		assert.Equal(t, false, cfg.Failover)
	})

	t.Run("EnvOnly", func(t *testing.T) {
		os.Setenv("SIDEKICK_BOLTROUTER_LOCAL", "true")
		os.Setenv("SIDEKICK_BOLTROUTER_PASSTHROUGH", "true")
		os.Setenv("SIDEKICK_BOLTROUTER_FAILOVER", "true")

		rootCmd.SetArgs([]string{"test"})
		err := rootCmd.Execute()
		require.NoError(t, err)

		cfg, err := getBoltRouterConfig(testServeCmd)
		require.NoError(t, err)
		assert.Equal(t, true, cfg.Local)
		assert.Equal(t, true, cfg.Passthrough)
		assert.Equal(t, true, cfg.Failover)
	})

	t.Run("FlagOnly", func(t *testing.T) {
		rootCmd.SetArgs([]string{"test", "--local", "--passthrough", "--failover"})
		err := rootCmd.Execute()
		require.NoError(t, err)

		cfg, _ := getBoltRouterConfig(testServeCmd)
		require.NoError(t, err)
		assert.Equal(t, true, cfg.Local)
		assert.Equal(t, true, cfg.Passthrough)
		assert.Equal(t, true, cfg.Failover)
	})

	// Flags should override env vars
	t.Run("EnvAndFlag", func(t *testing.T) {
		os.Setenv("SIDEKICK_BOLTROUTER_LOCAL", "true")
		os.Setenv("SIDEKICK_BOLTROUTER_PASSTHROUGH", "true")
		os.Setenv("SIDEKICK_BOLTROUTER_FAILOVER", "true")

		rootCmd.SetArgs([]string{"test", "--local=false", "--passthrough=false", "--failover=false"})
		err := rootCmd.Execute()
		require.NoError(t, err)

		cfg, err := getBoltRouterConfig(testServeCmd)
		require.NoError(t, err)
		assert.Equal(t, false, cfg.Local)
		assert.Equal(t, false, cfg.Passthrough)
		assert.Equal(t, false, cfg.Failover)
	})
}
