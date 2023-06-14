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

// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=104
const DEFAULT_PORT = 7075

func init() {
	initServerFlags(serveCmd)
	rootCmd.AddCommand(serveCmd)
}

func initServerFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("port", "p", DEFAULT_PORT, "The port for sidekick to listen on.")
	cmd.Flags().BoolP("local", "l", false, "Run sidekick in local (non cloud) mode. This is mostly use for testing locally.")
	cmd.Flags().Bool("passthrough", false, "Set passthrough flag to bolt requests.")
	cmd.Flags().BoolP("failover", "f", true, "Enables aws request failover if bolt request fails.")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serves the sidekick api",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		OnShutdown(cancel)

		boltRouterConfig := getBoltRouterConfig(cmd)

		cfg := api.Config{BoltRouter: boltRouterConfig}

		api, err := api.New(ctx, rootLogger, cfg)
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
				rootLogger.Error(err.Error())
			}
		}()

		rootLogger.Info(fmt.Sprintf("listening at http://127.0.0.1:%v", port))
		if err := server.ListenAndServe(); err != nil {
			return err
		}

		return nil
	},
}

func getBoltRouterConfig(cmd *cobra.Command) boltrouter.Config {
	boltRouterConfig := rootConfig.BoltRouter
	boltRouterConfig.Local, _ = cmd.Flags().GetBool("local")
	boltRouterConfig.Passthrough, _ = cmd.Flags().GetBool("passthrough")
	boltRouterConfig.Failover, _ = cmd.Flags().GetBool("failover")
	return boltRouterConfig
}
