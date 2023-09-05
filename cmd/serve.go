package cmd

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/project-n-oss/sidekick/api"
	"github.com/project-n-oss/sidekick/boltrouter"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"
)

//go:embed sidekick-local.granica.ai.pem
var sslCertCrt string

//go:embed sidekick-local.granica.ai-key.pem
var sslCertKey string

// DEFAULT_PORT
// From Unassigned https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=104
const DEFAULT_PORT = 7075

// DEFAULT_HTTPS_PORT
// From Unassigned https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=104
const DEFAULT_HTTPS_PORT = 7076

func init() {
	initServerFlags(serveCmd)
	rootCmd.AddCommand(serveCmd)
}

func initServerFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("port", "p", DEFAULT_PORT, "The port for sidekick to listen on.")
	cmd.Flags().IntP("https-port", "", DEFAULT_HTTPS_PORT, "The port for sidekick to listen on for https.")
	cmd.Flags().BoolP("local", "l", false, "Run sidekick in local (non cloud) mode. This is mostly use for testing locally.")
	cmd.Flags().String("bolt-endpoint-override", "", "Specify the local bolt endpoint with port to override in local mode. e.g: <local-bolt-ip>:9000")
	cmd.Flags().Bool("passthrough", false, "Set passthrough flag to bolt requests.")
	cmd.Flags().BoolP("failover", "f", false, "Enables aws request failover if bolt request fails.")
	cmd.Flags().String("crunch-traffic-split", "objectkeyhash", "Specify the crunch traffic split strategy: random or objectkeyhash")
	cmd.Flags().StringP("cloud-platform", "", "", "Cloud platform to use. one of: aws, gcp")
	cmd.Flags().BoolP("gcp-replicas", "", false, "Whether to query Quicksilver for replica IPs in GCP mode")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serves the sidekick api",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		OnShutdown(cancel)

		// validate cloud-platform is one of aws or gcp
		cloudPlatform, _ := cmd.Flags().GetString("cloud-platform")
		rootLogger.Info("cloud platform", zap.String("cloud-platform", cloudPlatform))
		if cloudPlatform != "aws" && cloudPlatform != "gcp" {
			return fmt.Errorf("cloud-platform must be one of: aws, gcp")
		}

		if cloudPlatform == "gcp" {
			gcpReplicas, _ := cmd.Flags().GetBool("gcp-replicas")
			rootLogger.Info("gcp replicas enabled", zap.Bool("gcp-replicas", gcpReplicas))
		}

		boltRouterConfig, err := getBoltRouterConfig(cmd)
		if err != nil {
			return err
		}

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

		httpsPort, _ := cmd.Flags().GetInt("https-port")
		cert, err := tls.X509KeyPair([]byte(sslCertCrt), []byte(sslCertKey))
		if err != nil {
			return err
		}
		httpsServer := &http.Server{
			Addr:    ":" + strconv.Itoa(httpsPort),
			Handler: handler,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}

		go func() {
			<-ctx.Done()
			if err := server.Shutdown(context.Background()); err != nil {
				rootLogger.Error("error shutting down server")
				rootLogger.Error(err.Error())
			}
			if err := httpsServer.Shutdown(context.Background()); err != nil {
				rootLogger.Error("error shutting down https server")
				rootLogger.Error(err.Error())
			}
		}()

		errs, ctx := errgroup.WithContext(ctx)
		errs.Go(func() error {
			rootLogger.Info(fmt.Sprintf("listening at http://127.0.0.1:%v", port))
			return server.ListenAndServe()
		})
		errs.Go(func() error {
			rootLogger.Info(fmt.Sprintf("listening at http://127.0.0.1:%d", httpsPort))
			return httpsServer.ListenAndServeTLS("", "")
		})
		return errs.Wait()
	},
}

func getBoltRouterConfig(cmd *cobra.Command) (boltrouter.Config, error) {
	boltRouterConfig := rootConfig.BoltRouter
	if cmd.Flags().Lookup("local").Changed {
		boltRouterConfig.Local, _ = cmd.Flags().GetBool("local")
	}
	if cmd.Flags().Lookup("cloud-platform").Changed {
		cp, _ := cmd.Flags().GetString("cloud-platform")
		cp = strings.ToLower(cp)
		boltRouterConfig.CloudPlatform = boltrouter.CloudPlatformType(boltrouter.CloudPlatformStrToTypeMap[cp])
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
	if cmd.Flags().Lookup("gcp-replicas").Changed {
		boltRouterConfig.GcpReplicasEnabled, _ = cmd.Flags().GetBool("gcp-replicas")
	}
	return boltRouterConfig, nil
}
