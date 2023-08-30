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

// DEFAULT_PORT
// From Unassigned https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=104
const DEFAULT_PORT = 7075

func init() {
	initServerFlags(serveCmd)
	rootCmd.AddCommand(serveCmd)
}

func initServerFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("port", "p", DEFAULT_PORT, "The port for sidekick to listen on.")
	cmd.Flags().BoolP("local", "l", false, "Run sidekick in local (non cloud) mode. This is mostly use for testing locally.")
	cmd.Flags().String("bolt-endpoint-override", "", "Specify the local bolt endpoint with port to override in local mode. e.g: <local-bolt-ip>:9000")
	cmd.Flags().Bool("passthrough", false, "Set passthrough flag to bolt requests.")
	cmd.Flags().BoolP("failover", "f", false, "Enables aws request failover if bolt request fails.")
	cmd.Flags().String("crunch-traffic-split", "objectkeyhash", "Specify the crunch traffic split strategy: random or objectkeyhash")
	cmd.Flags().StringP("cloud-platform", "", "", "cloud platform to use. one of: aws, gcp")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serves the sidekick api",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		OnShutdown(cancel)

		boltRouterConfig := getBoltRouterConfig(cmd)

		cfg := api.Config{
			BoltRouter: boltRouterConfig,
		}

		// Create api service to handle HTTP requests
		api, err := api.New(ctx, rootLogger, cfg)
		if err != nil {
			return err
		}
		handler := api.CreateHandler()

		// Start HTTP server and listen on the HTTP requests
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
	if cmd.Flags().Lookup("local").Changed {
		boltRouterConfig.Local, _ = cmd.Flags().GetBool("local")
	}
	if cmd.Flags().Lookup("cloud-platform").Changed {
		boltRouterConfig.CloudPlatform, _ = cmd.Flags().GetString("cloud-platform")
	}
	if cmd.Flags().Lookup("bolt-endpoint-override").Changed {
		boltRouterConfig.BoltEndpointOverride, _ = cmd.Flags().GetString("bolt-endpoint-override")
	}
	if cmd.Flags().Lookup("passthrough").Changed {
		boltRouterConfig.Passthrough, _ = cmd.Flags().GetBool("passthrough")
	}
	if cmd.Flags().Lookup("failover").Changed {
		boltRouterConfig.Failover, _ = cmd.Flags().GetBool("failover")
	}
	if cmd.Flags().Lookup("crunch-traffic-split").Changed {
		crunchTrafficSplitStr, _ := cmd.Flags().GetString("crunch-traffic-split")
		boltRouterConfig.CrunchTrafficSplit = boltrouter.CrunchTrafficSplitType(crunchTrafficSplitStr)
	}
	return boltRouterConfig
}
