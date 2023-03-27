package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/project-n-oss/sidekick/api"
	"github.com/project-n-oss/sidekick/boltrouter"

	"github.com/spf13/cobra"
)

const DEFAULT_PORT = 7071

func init() {
	serveCmd.Flags().IntP("port", "p", DEFAULT_PORT, "The port for sidekick to listen on.")
	serveCmd.Flags().BoolP("local", "l", false, "Run sidekick in local (non cloud) mode. This is mostly use for testing locally.")
	serveCmd.Flags().Bool("passthrough", false, "Set passthrough flag to bolt requests.")
	serveCmd.Flags().BoolP("failover", "f", true, "Enables aws request failover if bolt request fails.")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serves the sidekick api",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		OnShutdown(cancel)

		local, _ := cmd.Flags().GetBool("local")
		passthrough, _ := cmd.Flags().GetBool("passthrough")
		failover, _ := cmd.Flags().GetBool("failover")
		cfg := api.Config{
			BoltRouter: boltrouter.Config{
				Local:       local,
				Passthrough: passthrough,
				Failover:    failover,
			},
		}

		api, err := api.New(ctx, logger, cfg)
		if err != nil {
			return err
		}

		handler := api.CreateHandler()
		port, _ := cmd.Flags().GetInt("port")
		server := &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: handler,
		}

		go func() {
			<-ctx.Done()
			if err := server.Shutdown(context.Background()); err != nil {
				logger.Error(err.Error())
			}
		}()

		logger.Info(fmt.Sprintf("listening at http://127.0.0.1:%v", port))
		if err := server.ListenAndServe(); err != nil {
			return err
		}

		return nil
	},
}
