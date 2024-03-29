package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/project-n-oss/sidekick/boltrouter"

	"go.uber.org/zap"
)

type Api struct {
	logger *zap.Logger

	br *boltrouter.BoltRouter
}

func New(ctx context.Context, logger *zap.Logger, cfg Config) (*Api, error) {
	br, err := boltrouter.NewBoltRouter(ctx, logger, cfg.BoltRouter)
	if err != nil {
		return nil, err
	}

	if cfg.BoltRouter.Local {
		logger.Info("running sidekick locally")
	}

	// force refresh endpoints at the start
	if err := br.RefreshBoltInfo(ctx); err != nil {
		return nil, err
	}
	// Refresh endpoints periodically
	br.RefreshBoltInfoPeriodically(ctx)
	br.RefreshAWSCredentialsPeriodically(ctx, logger)

	return &Api{
		br:     br,
		logger: logger,
	}, nil
}

// CreateHandler creates the http.Handler for the sidekick api
func (a *Api) CreateHandler() http.Handler {
	handler := http.HandlerFunc(a.routeBase)
	handler = a.healthMiddleware(handler)
	handler = a.sessionMiddleware(handler)

	return handler
}

func (a *Api) routeBase(w http.ResponseWriter, req *http.Request) {
	sess := CtxSession(req.Context())
	ctx := req.Context()

	boltReq, err := sess.br.NewBoltRequest(ctx, sess.Logger(), req.Clone(ctx))
	if err != nil {
		if strings.Contains(err.Error(), "no auth header in request") {
			a.BadRequest(sess.Logger(), w, err)
			return
		}
		a.InternalError(sess.Logger(), w, err)
		return
	}

	if sess.Logger().Level() == zap.DebugLevel {
		dumpRequest(sess.Logger(), boltReq)
	}

	resp, failover, analytics, err := sess.br.DoRequest(sess.Logger(), boltReq)

	if sess.Logger().Level() == zap.DebugLevel {
		dumpAnalytics(sess.Logger(), analytics, err)
	}

	if err != nil {
		a.InternalError(sess.Logger(), w, err)
		return
	}

	sess.WithLogger(sess.Logger().
		With(zap.Int("statusCode", resp.StatusCode)).
		With(zap.Bool("failover", failover)))
	if analytics.AwsRequestResponseStatusCode > 0 {
		sess.WithLogger(sess.Logger().With(zap.String("endpt", "aws")))
	} else if analytics.BoltRequestResponseStatusCode > 0 {
		sess.WithLogger(sess.Logger().With(zap.String("endpt", analytics.BoltRequestUrl)))
	}

	// Convert the response headers to lower case, as Python etc libraries expect lower case.
	for k, values := range resp.Header {
		lowK := strings.ToLower(k)
		if strings.HasPrefix(lowK, "x-amz-meta") {
			w.Header()[lowK] = values
		} else {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}

	if !boltrouter.StatusCodeIs2xx(resp.StatusCode) {
		body := boltrouter.CopyRespBody(resp)
		b, _ := io.ReadAll(body)
		body.Close()
		sess.Logger().Debug("Status code is not 2xx in s3 response", zap.String("body", string(b)))
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		a.InternalError(sess.Logger(), w, err)
		return

	}
}

func dumpRequest(logger *zap.Logger, boltReq *boltrouter.BoltRequest) {
	boltDump, err := httputil.DumpRequest(boltReq.Bolt, true)
	if err != nil {
		logger.Error("dumping bolt request", zap.Error(err))
		return
	}

	var awsDump []byte
	if boltReq.Aws != nil {
		awsDump, err = httputil.DumpRequest(boltReq.Aws, true)
		if err != nil {
			logger.Error("dumping aws request", zap.Error(err))
			return
		}
	}

	var gcpDump []byte
	if boltReq.Gcp != nil {
		gcpDump, err = httputil.DumpRequest(boltReq.Gcp, true)
		if err != nil {
			logger.Error("dumping gcp request", zap.Error(err))
			return
		}
	}

	logger.Debug("BoltRequest dump", zap.String("bolt", string(boltDump)), zap.String("aws", string(awsDump)), zap.String("gcp", string(gcpDump)))
}

func dumpAnalytics(logger *zap.Logger, analytics *boltrouter.BoltRequestAnalytics, err error) {
	defaultValue := "N/A"

	logger.Debug("BoltRequestAnalytics dump",
		zap.Any("ObjectKey", orDefault(analytics.ObjectKey, defaultValue)),
		zap.Any("RequestBodySize", orDefault(analytics.RequestBodySize, defaultValue)),
		zap.Any("Method", orDefault(analytics.Method, defaultValue)),
		zap.Any("InitialRequestTarget", orDefault(analytics.InitialRequestTarget, defaultValue)),
		zap.Any("InitialRequestTargetReason", orDefault(analytics.InitialRequestTargetReason, defaultValue)),
		zap.Any("BoltReplicaIp", orDefault(analytics.BoltRequestUrl, defaultValue)),
		zap.Any("BoltRequestDuration", orDefault(analytics.BoltRequestDuration, defaultValue)),
		zap.Any("BoltRequestResponseStatusCode", orDefault(analytics.BoltRequestResponseStatusCode, defaultValue)),
		zap.Any("AwsRequestDuration", orDefault(analytics.AwsRequestDuration, defaultValue)),
		zap.Any("AwsRequestResponseStatusCode", orDefault(analytics.AwsRequestResponseStatusCode, defaultValue)),
		zap.Any("Error", orDefault(err, defaultValue)),
	)
}

func orDefault(value interface{}, defaultValue interface{}) interface{} {
	if value == nil || value == "" || value == 0 {
		return defaultValue
	}
	return value
}
