package boltrouter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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
func (br *BoltRouter) NewBoltRequest(ctx context.Context, req *http.Request) (*BoltRequest, error) {
	sourceBucket := extractSourceBucket(req)
	failoverRequest, err := newFailoverAwsRequest(ctx, req.Clone(ctx), br.awsCred, sourceBucket, br.boltVars.Region.Get())
	if err != nil {
		return nil, fmt.Errorf("failed to make failover request: %w", err)
	}

	authPrefix := randString(4)
	headReq, err := signedAwsHeadRequest(ctx, req, br.awsCred, sourceBucket.bucket, br.boltVars.Region.Get(), authPrefix)
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
	if sourceBucket.style == virtualHostedStyle {
		BoltURL = BoltURL.JoinPath(sourceBucket.bucket, req.URL.EscapedPath())

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
func newFailoverAwsRequest(ctx context.Context, req *http.Request, awsCred aws.Credentials, sourceBucket SourceBucket, region string) (*http.Request, error) {
	var host string
	switch sourceBucket.style {
	case virtualHostedStyle:
		host = fmt.Sprintf("%s.s3.%s.amazonaws.com", sourceBucket.bucket, region)
	// default to path style
	default:
		host = fmt.Sprintf("s3.%s.amazonaws.com", region)

	}

	req.Header.Del("Authorization")
	req.Header.Del("X-Amz-Security-Token")

	req.URL.Host = host
	req.Host = host
	req.URL.Scheme = "https"
	req.RequestURI = ""

	payloadHash := req.Header.Get("X-Amz-Content-Sha256")

	awsSigner := v4.NewSigner()
	if err := awsSigner.SignHTTP(ctx, awsCred, req, payloadHash, "s3", region, time.Now()); err != nil {
		return nil, err
	}

	return req.Clone(ctx), nil
}

type s3RequestStyle string

const (
	virtualHostedStyle s3RequestStyle = "virtual-hosted-style"
	pathStyle          s3RequestStyle = "path-style"
	nAuthDummy         s3RequestStyle = "n-auth-dummy"
)

type SourceBucket struct {
	bucket string
	style  s3RequestStyle
}

// extractSourceBucket extracts the aws request bucket using Path-style or Virtual-hosted-style requests.
// https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html
// This method will "n-auth-dummy" if nothing is found
func extractSourceBucket(req *http.Request) SourceBucket {
	// virtual-hosted-style
	if split := strings.Split(req.Host, "."); len(split) > 1 {
		bucket := split[0]
		return SourceBucket{bucket: bucket, style: virtualHostedStyle}
	}

	// path-style request
	if paths := strings.Split(req.URL.EscapedPath(), "/"); len(paths) > 1 {
		bucket := paths[1]
		return SourceBucket{bucket: bucket, style: pathStyle}
	}

	return SourceBucket{bucket: "n-auth-dummy", style: nAuthDummy}
}

// DoBoltRequest sends an HTTP Bolt request and returns an HTTP response, following policy (such as redirects, cookies, auth) as configured on the client.
// DoBoltRequest will failover to AWS if the Bolt request fails and the config.Failover is set to true.
// DoboltRequest will return a bool indicating if the request was a failover.
func (br *BoltRouter) DoBoltRequest(logger *zap.Logger, boltReq *BoltRequest) (*http.Response, bool, error) {
	resp, err := br.boltHttpClient.Do(boltReq.Bolt)
	if err != nil {
		return resp, false, err
	} else if resp.StatusCode != 200 && br.config.Failover {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		logger.Warn("bolt request failed", zap.Int("status code", resp.StatusCode), zap.String("msg", string(b)))
		resp, err := http.DefaultClient.Do(boltReq.Aws)
		return resp, true, err
	}

	return resp, false, nil
}
