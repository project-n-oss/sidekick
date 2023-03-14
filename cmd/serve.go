package cmd

import (
	"context"
	"fmt"
	"net/http"
	"sidekick/api"
	"sidekick/boltrouter"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "the port to listen on")
	serveCmd.Flags().BoolP("local", "l", false, "running sidekick locally")
	serveCmd.Flags().Bool("passthrough", false, "set passthrough flag to try in bolt requests")
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
		cfg := api.Config{
			BoltRouter: boltrouter.Config{
				Local:       local,
				Passthrough: passthrough,
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

func init() {
	rootCmd.AddCommand(serveCmd)
}
