package boltrouter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// SHA value for empty payload. As head object request is with empty payload
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/signer/v4#Signer.SignHTTP
const emptyPayloadHash string = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func (br *BoltRouter) NewBoltRequest(ctx context.Context, req *http.Request) (*http.Request, error) {
	authPrefix := randString(4)
	headReq, err := signedAwsHeadRequest(ctx, req, br.awsCred, br.boltVars.Region.Get(), authPrefix)
	if err != nil {
		return nil, fmt.Errorf("could not make signed aws head request")
	}

	boltURI, err := br.SelectBoltEndpoint(ctx, req.Method)
	if err != nil {
		return nil, err
	}
	req.URL.Host = boltURI.Host

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
	req.Header.Set("X-Bolt-Auth-Prefic", authPrefix)
	req.Header.Set("User-Agent", fmt.Sprintf("%s%s", br.boltVars.UserAgentPrefix.Get(), req.Header.Get("User-Agent")))

	if !br.config.Passthrough {
		req.Header.Set("X-Bolt-Passthrough-Read", "disable")
	}

	return req.Clone(ctx), nil
}

func signedAwsHeadRequest(ctx context.Context, req *http.Request, awsCred aws.Credentials, region string, authPrefix string) (*http.Request, error) {
	var sourceBucket string
	if paths := strings.Split(req.URL.EscapedPath(), "/"); len(paths) > 1 {
		sourceBucket = paths[1]
	}
	if sourceBucket == "" {
		sourceBucket = "n-auth-dummy"
	}

	headObjectURL := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s/auth", region, sourceBucket, authPrefix)
	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, headObjectURL, nil)
	if err != nil {
		return nil, err
	}

	// TODO: change signing time/skew clock to take advantage of bolt caching
	awsSigner := v4.NewSigner()
	if err := awsSigner.SignHTTP(ctx, awsCred, headReq, emptyPayloadHash, "s3", region, time.Now()); err != nil {
		return nil, err
	}
	return headReq, nil
}

func (br *BoltRouter) DoBoltRequest(req *http.Request) (*http.Response, error) {
	return br.httpClient.StandardClient().Do(req)
}
