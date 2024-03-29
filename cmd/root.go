package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

//go:embed ascii.txt
var asciiArt string

//go:embed version.md
var versionFile string

func version() string {
	return strings.Split(versionFile, " ")[1]
}

func init() {
	rootLogger = NewLogger(zapcore.InfoLevel)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringP("log-level", "", "info", "log level. one of: debug, info, warn, error, fatal, panic")
	rootCmd.PersistentFlags().StringP("config", "c", "", "read configuration from this file")
}

var (
	rootLogger *zap.Logger
	rootConfig = DefaultConfig
)

var rootCmd = &cobra.Command{
	Use:           "sidekick",
	Version:       version(),
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logLevel, _ := cmd.Flags().GetString("log-level")
		zapLogLevel, _ := zapcore.ParseLevel(logLevel)
		rootLogger = NewLogger(zapLogLevel)

		OnShutdown(func() {
			_ = rootLogger.Sync()
		})

		if _, err := os.Stat(".env"); err == nil {
			err := godotenv.Load()
			if err != nil {
				return fmt.Errorf("could not load .env file. %v", err)
			}
		}

		if config, _ := cmd.Flags().GetString("config"); config != "" {
			f, err := os.Open(config)
			if err != nil {
				return err
			}
			defer f.Close()
			if err := yaml.NewDecoder(f).Decode(&rootConfig); err != nil {
				return fmt.Errorf("failed to decode config: %w", err)
			}
		} else if f, err := os.Open("config.yaml"); err == nil {
			defer f.Close()
			if err := yaml.NewDecoder(f).Decode(&rootConfig); err != nil {
				return fmt.Errorf("failed to decode config: %w", err)
			}
		}

		if err := unmarshalConfigFromEnv(&rootConfig); err != nil {
			return err
		}

		// wait forever for sig signal
		go func() {
			waitForTermSignal()
		}()

		fmt.Println(asciiArt)
		fmt.Printf("Version: %s\n", version())

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Do a graceful shutdown
		Shutdown()
		return nil
	},
}

func waitForTermSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	sig := <-sigs
	rootLogger.Info("received signal, shutting down", zap.String("signal", sig.String()))

	// Do a graceful shutdown
	Shutdown()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		rootLogger.Fatal(err.Error())
	}
}
