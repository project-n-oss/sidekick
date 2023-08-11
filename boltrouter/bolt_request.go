package boltrouter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"go.uber.org/zap"
)

type BoltRequest struct {
	Bolt *http.Request
	Aws  *http.Request
}

type BoltRequestAnalytics struct {
	ObjectKey                     string
	RequestBodySize               int
	Method                        string
	InitialRequestTarget          string
	InitialRequestTargetReason    string
	BoltRequestUrl                string
	BoltRequestDuration           time.Duration
	BoltRequestResponseStatusCode int
	AwsRequestDuration            time.Duration
	AwsRequestResponseStatusCode  int
}

var (
	ErrPanicDuringBoltRequest = errors.New("panic occurred during Bolt request")
	ErrPanicDuringAwsRequest  = errors.New("panic occurred during AWS request")
)

// NewBoltRequest transforms the passed in intercepted aws http.Request and returns
// a new http.Request Ready to be sent to Bolt.
// This new http.Request is routed to the correct Bolt endpoint and signed correctly.
func (br *BoltRouter) NewBoltRequest(ctx context.Context, logger *zap.Logger, req *http.Request) (*BoltRequest, error) {
	sourceBucket, err := extractSourceBucket(ctx, logger, req, br.boltVars.Region.Get())
	if err != nil {
		return nil, fmt.Errorf("could not extract source bucket: %w", err)
	}

	awsCred, err := getAwsCredentialsFromRegion(ctx, sourceBucket.Region)
	if err != nil {
		return nil, fmt.Errorf("could not get aws credentials: %w", err)
	}

	failoverRequest, err := newFailoverAwsRequest(ctx, req, awsCred, sourceBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to make failover request: %w", err)
	}

	authPrefix := randString(4)
	headReq, err := signedAwsHeadRequest(ctx, req, awsCred, sourceBucket.Bucket, sourceBucket.Region, authPrefix)
	if err != nil {
		return nil, fmt.Errorf("could not make signed aws head request: %w", err)
	}

	BoltURL, err := br.SelectBoltEndpoint(ctx, req.Method)
	if err != nil {
		return nil, err
	}

	// RequestURI is the unmodified request-target of the Request-Line (RFC 7230, Section 3.1.1) as sent by the client to a server.
	//  It is an error to set this field in an HTTP client request.
	req.RequestURI = ""
	if sourceBucket.Style == virtualHostedStyle {
		BoltURL = BoltURL.JoinPath(sourceBucket.Bucket, req.URL.EscapedPath())

	} else {
		BoltURL = BoltURL.JoinPath(req.URL.Path)
	}
	// Bolt will only accept query if path starts with "/".
	// Bolt will return a 400 error otherwise
	BoltURL.Path = "/" + BoltURL.Path
	BoltURL.RawQuery = req.URL.RawQuery
	req.URL = BoltURL
	req.URL.Scheme = "https"

	if v := headReq.Header.Get("X-Amz-Security-Token"); v != "" {
		req.Header.Set("X-Amz-Security-Token", v)
	}
	if v := headReq.Header.Get("X-Amz-Date"); v != "" {
		req.Header.Set("X-Amz-Date", v)
	}
	if v := headReq.Header.Get("Authorization"); v != "" {
		req.Header.Set("Authorization", v)
	}
	if v := headReq.Header.Get("X-Amz-Content-Sha256"); v != "" {
		req.Header.Set("X-Amz-Content-Sha256", v)
	}

	req.Header.Set("Host", br.boltVars.BoltHostname.Get())
	req.Host = br.boltVars.BoltHostname.Get()
	req.Header.Set("X-Bolt-Auth-Prefix", authPrefix)
	req.Header.Set("User-Agent", fmt.Sprintf("%s%s", br.boltVars.UserAgentPrefix.Get(), req.Header.Get("User-Agent")))
	req.Header.Set("X-Bolt-Availability-Zone", br.boltVars.ZoneId.Get())

	if !br.config.Passthrough {
		req.Header.Set("X-Bolt-Passthrough-Read", "disable")
	}

	return &BoltRequest{
		Bolt: req.Clone(ctx),
		Aws:  failoverRequest.Clone(ctx),
	}, nil
}

// SHA value for empty payload. As head object request is with empty payload
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/signer/v4#Signer.SignHTTP
const emptyPayloadHash string = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// signedAwsHeadRequest returns a new Head http.Request signed by AWS v4 signer.
func signedAwsHeadRequest(ctx context.Context, req *http.Request, awsCred aws.Credentials, sourceBucket string, region string, authPrefix string) (*http.Request, error) {
	headObjectURL := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s/auth", region, sourceBucket, authPrefix)
	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, headObjectURL, nil)
	if err != nil {
		return nil, err
	}

	if v := req.Header.Get("X-Amz-Security-Token"); v != "" {
		headReq.Header.Set("X-Amz-Security-Token", v)
	}
	headReq.Header.Set("X-Amz-Content-Sha256", emptyPayloadHash)

	// TODO: change signing time/skew clock to take advantage of bolt caching
	awsSigner := v4.NewSigner()
	if err := awsSigner.SignHTTP(ctx, awsCred, headReq, emptyPayloadHash, "s3", region, time.Now()); err != nil {
		return nil, err
	}
	return headReq, nil
}

// newFailoverAwsRequest creates a standard aws s3 request that can be used as a failover if the Bolt request fails.
func newFailoverAwsRequest(ctx context.Context, req *http.Request, awsCred aws.Credentials, sourceBucket SourceBucket) (*http.Request, error) {
	var host string
	switch sourceBucket.Style {
	case virtualHostedStyle:
		host = fmt.Sprintf("%s.s3.%s.amazonaws.com", sourceBucket.Bucket, sourceBucket.Region)
	// default to path style
	default:
		host = fmt.Sprintf("s3.%s.amazonaws.com", sourceBucket.Region)

	}

	clone := req.Clone(ctx)

	clone.Header.Del("Authorization")
	clone.Header.Del("X-Amz-Security-Token")

	clone.URL.Host = host
	clone.Host = host
	clone.URL.Scheme = "https"
	clone.RequestURI = ""
	// This needs to be set to "" in order to fix unicode errors in RawPath
	// This forces to use the well formated req.URL.Path value instead
	clone.URL.RawPath = ""

	// req.Clone(ctx) does not clone Body, need to clone body manually
	CopyReqBody(req, clone)

	payloadHash := req.Header.Get("X-Amz-Content-Sha256")

	awsSigner := v4.NewSigner()
	if err := awsSigner.SignHTTP(ctx, awsCred, clone, payloadHash, "s3", sourceBucket.Region, time.Now()); err != nil {
		return nil, err
	}

	return clone, nil
}

// DoRequest sends an HTTP Bolt request and returns an HTTP response, following policy (such as redirects, cookies, auth) as configured on the client.
// DoRequest will failover to AWS if the Bolt request fails and the config.Failover is set to true.
// DoboltRequest will return a bool indicating if the request was a failover.
// DoRequest will return a BoltRequestAnalytics struct with analytics about the request.
func (br *BoltRouter) DoRequest(logger *zap.Logger, boltReq *BoltRequest) (*http.Response, bool, *BoltRequestAnalytics, error) {
	initialRequestTarget, reason, err := br.SelectInitialRequestTarget()

	boltRequestAnalytics := &BoltRequestAnalytics{
		ObjectKey:                     boltReq.Bolt.URL.Path,
		RequestBodySize:               int(boltReq.Bolt.ContentLength),
		Method:                        boltReq.Bolt.Method,
		InitialRequestTarget:          initialRequestTarget,
		InitialRequestTargetReason:    reason,
		BoltRequestUrl:                boltReq.Bolt.URL.Hostname(),
		BoltRequestDuration:           time.Duration(0),
		BoltRequestResponseStatusCode: -1,
		AwsRequestDuration:            -1,
		AwsRequestResponseStatusCode:  -1,
	}

	if err != nil {
		return nil, false, boltRequestAnalytics, err
	}

	logger.Debug("initial request target", zap.String("target", initialRequestTarget), zap.String("reason", reason))

	if initialRequestTarget == "bolt" {
		resp, isFailoverRequest, err := br.doBoltRequest(logger, boltReq, false, boltRequestAnalytics)
		// if nothing during br.doBoltRequest panics, err will not be of type ErrPanicDuringBoltRequest so failover was
		// handled inside the function as needed and we can just return
		// If the err is of type ErrPanicDuringBoltRequest then we need to failover to AWS manually since .doBoltRequest
		// halted execution before it could failover
		if err != nil && err == ErrPanicDuringBoltRequest && br.config.Failover {
			logger.Error("panic occurred during Bolt request, failing over to AWS", zap.Error(err))
			resp, isFailoverRequest, err = br.doAwsRequest(logger, boltReq, true, boltRequestAnalytics)
		}
		return resp, isFailoverRequest, boltRequestAnalytics, err
	} else {
		resp, isFailoverRequest, err := br.doAwsRequest(logger, boltReq, false, boltRequestAnalytics)
		return resp, isFailoverRequest, boltRequestAnalytics, err
	}
}

func (br *BoltRouter) doBoltRequest(logger *zap.Logger, boltReq *BoltRequest, isFailover bool, analytics *BoltRequestAnalytics) (resp *http.Response, isFailoverRequest bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			resp = nil
			err = ErrPanicDuringBoltRequest
		}
	}()

	beginTime := time.Now()
	resp, err = br.boltHttpClient.Do(boltReq.Bolt)
	duration := time.Since(beginTime)
	analytics.BoltRequestDuration = duration
	if err != nil {
		if resp != nil {
			analytics.BoltRequestResponseStatusCode = resp.StatusCode
		}
		return resp, isFailoverRequest, err
	} else if !StatusCodeIs2xx(resp.StatusCode) && br.config.Failover && !isFailoverRequest {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logger.Error("bolt request failed", zap.Int("statusCode", resp.StatusCode), zap.String("body", string(b)))
		return br.doAwsRequest(logger, boltReq, true, analytics)
	}
	analytics.BoltRequestResponseStatusCode = resp.StatusCode
	return resp, isFailoverRequest, nil
}

func (br *BoltRouter) doAwsRequest(logger *zap.Logger, boltReq *BoltRequest, isFailover bool, analytics *BoltRequestAnalytics) (resp *http.Response, isFailoverRequest bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic occurred during AWS request", zap.Any("panic", r))
			resp = nil
			err = ErrPanicDuringAwsRequest
		}
	}()

	beginTime := time.Now()
	resp, err = http.DefaultClient.Do(boltReq.Aws)
	duration := time.Since(beginTime)
	analytics.AwsRequestDuration = duration
	if err != nil {
		if resp != nil {
			analytics.AwsRequestResponseStatusCode = resp.StatusCode
		}
		return resp, isFailoverRequest, err
	} else if !StatusCodeIs2xx(resp.StatusCode) && resp.StatusCode == 404 && !isFailoverRequest {
		// failover to bolt
		logger.Error("aws request failed, falling back to bolt", zap.Int("statusCode", resp.StatusCode))
		return br.doBoltRequest(logger, boltReq, true, analytics)
	}
	analytics.AwsRequestResponseStatusCode = resp.StatusCode
	return resp, isFailoverRequest, nil
}

func StatusCodeIs2xx(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
