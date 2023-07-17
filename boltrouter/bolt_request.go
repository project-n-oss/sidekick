package boltrouter

import (
	"context"
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

// DoBoltRequest sends an HTTP Bolt request and returns an HTTP response, following policy (such as redirects, cookies, auth) as configured on the client.
// DoBoltRequest will failover to AWS if the Bolt request fails and the config.Failover is set to true.
// DoboltRequest will return a bool indicating if the request was a failover.
func (br *BoltRouter) DoBoltRequest(logger *zap.Logger, boltReq *BoltRequest) (*http.Response, bool, error) {
	resp, err := br.boltHttpClient.Do(boltReq.Bolt)
	if err != nil {
		return resp, false, err
	} else if !StatusCodeIs2xx(resp.StatusCode) && br.config.Failover {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logger.Warn("bolt request failed", zap.Int("statusCode", resp.StatusCode), zap.String("body", string(b)))
		resp, err := http.DefaultClient.Do(boltReq.Aws)
		return resp, true, err
	}

	return resp, false, nil
}

func StatusCodeIs2xx(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
