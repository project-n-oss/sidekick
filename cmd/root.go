package cmd

import (
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

var logger *zap.Logger

var rootCmd = &cobra.Command{
	Use:           "sidekick",
	Version:       getVersion(),
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		logger = NewLogger(verbose)

		if _, err := os.Stat(".env"); err == nil {
			err := godotenv.Load()
			if err != nil {
				return fmt.Errorf("could not load .env file")
			}
		}

		// wait forever for sig signal
		go func() {
			WaitForTermSignal()
		}()

		fmt.Println(asciiArt)
		logger.Sugar().Infof("Version: %s", getVersion())

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
	logger.Info("received signal, shutting down", zap.String("signal", sig.String()))

	// Do a graceful shutdown
	Shutdown()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
