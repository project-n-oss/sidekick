package bolt_integration_tests

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sidekick/api"
	"sidekick/bolt_integration_tests/aws"
	"sidekick/bolt_integration_tests/aws/utils"
	"sidekick/boltrouter"
	"strconv"
	"syscall"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
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

var boltIntegration = flag.Bool("i", false, "run bolt integration test suite")
var local = flag.Bool("l", false, "run sidekick locally")
var port = flag.Int("p", 8000, "the port for sidekick to listen on")

func TestAws(t *testing.T) {
	if !*boltIntegration {
		t.Skip("skipping bolt integration test suite")
	}

	ctx := context.Background()
	SetupSidekick(t, ctx)

	suite.Run(t, new(aws.AwsSuite))
}

func SetupSidekick(t *testing.T, ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	logger := zaptest.NewLogger(t)

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
		sig := <-sigs
		logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
	}()

	cfg := api.Config{
		BoltRouter: boltrouter.Config{
			Local:       *local,
			Passthrough: false,
		},
	}

	api, err := api.New(ctx, logger, cfg)
	require.NoError(t, err)
	handler := api.CreateHandler()

	listenCfg := net.ListenConfig{}
	addr := "localhost:" + strconv.Itoa(*port)
	listner, err := listenCfg.Listen(ctx, "tcp", addr)
	require.NoError(t, err)
	utils.SidekickURL = "http://" + addr

	go func() {
		<-ctx.Done()
		err := listner.Close()
		require.NoError(t, err)
	}()

	go func() {
		http.Serve(listner, handler)
	}()
}
