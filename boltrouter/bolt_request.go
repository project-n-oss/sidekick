package boltrouter

import (
	"context"
	fipssha256 "crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	md5simd "github.com/minio/md5-simd"
	"github.com/minio/minio-go/v7/pkg/signer"
	awsSigner "github.com/project-n-oss/sidekick/boltrouter/signer/aws"
	"github.com/project-n-oss/sidekick/common"
	"go.uber.org/zap"
)

// hashWrapper implements the md5simd.Hasher interface.
type hashWrapper struct {
	hash.Hash
}

func newSHA256Hasher() md5simd.Hasher {
	return &hashWrapper{Hash: fipssha256.New()}
}

func (m *hashWrapper) Close() {
	m.Hash = nil
}

type BoltRequest struct {
	Bolt *http.Request
	Aws  *http.Request
}

// NewBoltRequest transforms the passed in intercepted aws http.Request and returns
// a new http.Request Ready to be sent to Bolt.
// This new http.Request is routed to the correct Bolt endpoint and signed correctly.
func (br *BoltRouter) NewBoltRequest(ctx context.Context, req *http.Request) (*BoltRequest, error) {
	boltReq := req.Clone(ctx)
	sourceBucket, err := extractSourceBucket(ctx, boltReq)
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
	headReq, err := awsSigner.SignedAwsHeadRequest(ctx, boltReq, awsCred, sourceBucket.Bucket, sourceBucket.Region, authPrefix)
	if err != nil {
		return nil, fmt.Errorf("could not make signed aws head request: %w", err)
	}

	BoltURL, err := br.SelectBoltEndpoint(ctx, boltReq.Method)
	if err != nil {
		return nil, err
	}

	// RequestURI is the unmodified request-target of the Request-Line (RFC 7230, Section 3.1.1) as sent by the client to a server.
	//  It is an error to set this field in an HTTP client request.
	boltReq.RequestURI = ""
	if sourceBucket.Style == virtualHostedStyle {
		BoltURL = BoltURL.JoinPath(sourceBucket.Bucket, boltReq.URL.EscapedPath())

	} else {
		BoltURL = BoltURL.JoinPath(boltReq.URL.Path)
	}
	// Bolt will only accept query if path starts with "/".
	// Bolt will return a 400 error otherwise
	BoltURL.Path = "/" + BoltURL.Path
	BoltURL.RawQuery = boltReq.URL.RawQuery
	boltReq.URL = BoltURL
	boltReq.URL.Scheme = "https"

	if v := headReq.Header.Get("X-Amz-Security-Token"); v != "" {
		boltReq.Header.Set("X-Amz-Security-Token", v)
	}
	if v := headReq.Header.Get("X-Amz-Date"); v != "" {
		boltReq.Header.Set("X-Amz-Date", v)
	}
	if v := headReq.Header.Get("Authorization"); v != "" {
		boltReq.Header.Set("Authorization", v)
	}
	if v := headReq.Header.Get("X-Amz-Content-Sha256"); v != "" {
		boltReq.Header.Set("X-Amz-Content-Sha256", v)
	}

	boltReq.Header.Set("Host", br.boltVars.BoltHostname.Get())
	boltReq.Host = br.boltVars.BoltHostname.Get()
	boltReq.Header.Set("X-Bolt-Auth-Prefix", authPrefix)
	boltReq.Header.Set("User-Agent", fmt.Sprintf("%s%s", br.boltVars.UserAgentPrefix.Get(), boltReq.Header.Get("User-Agent")))

	if !br.config.Passthrough {
		boltReq.Header.Set("X-Bolt-Passthrough-Read", "disable")
	}

	dataLen := boltReq.ContentLength
	boltReq = signer.StreamingSignV4(boltReq, awsCred.AccessKeyID, awsCred.SecretAccessKey, awsCred.SessionToken, br.boltVars.Region.Get(), dataLen, time.Now(), newSHA256Hasher())

	return &BoltRequest{
		Bolt: boltReq.Clone(ctx),
		Aws:  failoverRequest.Clone(ctx),
	}, nil
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
	common.CopyReqBody(req, clone)

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
