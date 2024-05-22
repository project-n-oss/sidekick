package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/project-n-oss/sidekick/pkg/config"
	"github.com/project-n-oss/sidekick/pkg/logger"
	"github.com/project-n-oss/sidekick/pkg/shutdown"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

//go:embed ascii.txt
var bannerArt string

//go:embed version.md
var versionFile string

func version() string {
	return strings.Split(versionFile, " ")[1]
}

func init() {
	rootLogger, _ = logger.NewLogger(false)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "make output more verbose")
	rootCmd.PersistentFlags().StringP("config", "c", "", "read configuration from this file")
}

var (
	rootLogger *zap.Logger
	rootConfig = DefaultConfig
)

var rootCmd = &cobra.Command{
	Use:           "sidekick-router",
	Version:       version(),
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		rootLogger, _ = logger.NewLogger(verbose)
		shutdown.OnShutdown(func() {
			rootLogger.Sync()
		})

		if _, err := os.Stat(".env"); err == nil {
			err := godotenv.Load()
			if err != nil {
				return fmt.Errorf("could not load .env file. %v", err)
			}
		}

		configFilePath, _ := cmd.Flags().GetString("config")
		cfgOpts := []func(*config.UnmarshalConfigOptions){}
		if configFilePath != "" {
			cfgOpts = append(cfgOpts, config.WithFilePath(configFilePath))
		}

		if err := config.UnmarshalConfig(context.Background(), cfgEnvPrefix, &rootConfig, cfgOpts...); err != nil {
			return err
		}

		// wait forever for sig signal
		go func() {
			waitForTermSignal()
		}()

		fmt.Println(bannerArt)
		fmt.Printf("Version: %s\n", version())

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Do a graceful shutdown
		shutdown.Shutdown()
		return nil
	},
}

func waitForTermSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	sig := <-sigs
	rootLogger.Info("received signal, shutting down", zap.String("signal", sig.String()))

	// Do a graceful shutdown
	shutdown.Shutdown()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		rootLogger.Fatal(err.Error())
	}
}
