package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/project-n-oss/sidekick/api"
	"github.com/project-n-oss/sidekick/app"
	"github.com/project-n-oss/sidekick/pkg/shutdown"
	"github.com/spf13/cobra"
)

// DEFAULT_PORT
// From Unassigned https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=104
const DEFAULT_PORT = 7075

func init() {
	serveCmd.Flags().IntP("port", "p", DEFAULT_PORT, "the port to listen on")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serves the proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		shutdown.OnShutdown(cancel)

		app, err := app.New(ctx, rootLogger, rootConfig.App)
		if err != nil {
			return fmt.Errorf("could not create app: %w", err)
		}

		port, _ := cmd.Flags().GetInt("port")
		api, err := api.New(ctx, rootConfig.Api, app)
		if err != nil {
			return err
		}

		handler := api.CreateHandler()
		server := &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: handler,
		}

		go func() {
			<-ctx.Done()
			if err := server.Shutdown(context.Background()); err != nil {
				rootLogger.Error("error shutting down server")
				rootLogger.Error(err.Error())
			}
		}()

		rootLogger.Sugar().Infof("listening at http://localhost:%v", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}

		return nil
	},
}
