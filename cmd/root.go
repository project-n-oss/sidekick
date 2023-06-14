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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

//go:embed ascii.txt
var asciiArt string

//go:embed version.md
var versionFile string

func getVersion() string {
	return strings.Split(versionFile, " ")[1]
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "make output more verbose")
	rootCmd.PersistentFlags().StringP("config", "c", "", "read configuration from this file")
}

var rootLogger *zap.Logger
var rootConfig = NewConfig()

var rootCmd = &cobra.Command{
	Use:           "sidekick",
	Version:       getVersion(),
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		rootLogger = NewLogger(verbose)
		OnShutdown(func() {
			_ = rootLogger.Sync()
		})

		if _, err := os.Stat(".env"); err == nil {
			err := godotenv.Load()
			if err != nil {
				return fmt.Errorf("could not load .env file")
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

		if err := UnmarshalConfigFromEnv(context.Background(), &rootConfig); err != nil {
			return err
		}

		// wait forever for sig signal
		go func() {
			WaitForTermSignal()
		}()

		fmt.Println(asciiArt)
		rootLogger.Sugar().Infof("Version: %s", getVersion())

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Do a graceful shutdown
		Shutdown()
		return nil
	},
}

func WaitForTermSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	sig := <-sigs
	rootLogger.Info("received signal, shutting down", zap.String("signal", sig.String()))

	// Do a graceful shutdown
	Shutdown()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
