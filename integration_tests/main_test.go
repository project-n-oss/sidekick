package integration_tests

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/project-n-oss/sidekick/api"
	"github.com/project-n-oss/sidekick/boltrouter"
	"github.com/project-n-oss/sidekick/cmd"
	"github.com/project-n-oss/sidekick/integration_tests/aws"
	"github.com/project-n-oss/sidekick/integration_tests/aws/utils"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("could not load .env file")
		}
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}

var boltIntegration = flag.Bool("bi", false, "run bolt integration test suite")
var sidekickURL = flag.String("sidekick-url", "", "the url sidekick is listening on if ran seperately. If not set, sidekick will be started in a goroutine")
var port = flag.Int("p", cmd.DEFAULT_PORT, "the port for sidekick to listen on")

func TestAws(t *testing.T) {
	if !*boltIntegration {
		t.Skip("skipping bolt integration test suite")
	}

	ctx := context.Background()
	if *sidekickURL == "" {
		SetupSidekick(t, ctx)
	} else {
		utils.SidekickURL = *sidekickURL
	}

	suite.Run(t, new(aws.AwsSuite))
}

func SetupSidekick(t *testing.T, ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	logger := cmd.NewLogger("debug")

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
		sig := <-sigs
		logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
	}()

	cfg := api.Config{
		BoltRouter: boltrouter.Config{
			Passthrough: false,
			Failover:    true,
		},
	}

	api, err := api.New(ctx, logger, cfg)
	require.NoError(t, err)
	handler := api.CreateHandler()

	listenCfg := net.ListenConfig{}
	addr := ":" + strconv.Itoa(*port)
	listener, err := listenCfg.Listen(ctx, "tcp", addr)
	require.NoError(t, err)
	utils.SidekickURL = "http://localhost" + addr

	go func() {
		<-ctx.Done()
		err := listener.Close()
		require.NoError(t, err)
		logger.Sync()
	}()

	go func() {
		err := http.Serve(listener, handler)
		if err != nil {
			logger.Error(fmt.Errorf("server err: %w", err).Error())
		}
	}()

	time.Sleep(1 * time.Second)
}
