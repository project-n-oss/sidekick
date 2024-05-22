package aws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/project-n-oss/sidekick/pkg/copybody"
	"go.uber.org/zap"
)

type Options struct {
	path *string
}

func WithPath(path string) func(*Options) {
	return func(o *Options) {
		o.path = &path
	}
}

// NewRequest creates a standard aws s3 request
func NewRequest(ctx context.Context, logger *zap.Logger, req *http.Request, sourceBucket SourceBucket, opts ...func(*Options)) (*http.Request, error) {
	options := &Options{
		path: &req.URL.Path,
	}
	for _, opt := range opts {
		opt(options)
	}

	awsCred, err := getCredentialsFromRegion(ctx, sourceBucket.Region)
	if err != nil {
		return nil, fmt.Errorf("could not get aws credentials: %w", err)
	}

	var host string
	switch sourceBucket.Style {
	case VirtualHostedStyle:
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
	clone.URL.Path = *options.path
	// This needs to be set to "" in order to fix unicode errors in RawPath
	// This forces to use the well formated req.URL.Path value instead
	clone.URL.RawPath = ""

	// req.Clone(ctx) does not clone Body, need to clone body manually
	copybody.CopyReqBody(req, clone)

	payloadHash := req.Header.Get("X-Amz-Content-Sha256")

	awsSigner := v4.NewSigner()
	if err := awsSigner.SignHTTP(ctx, awsCred, clone, payloadHash, "s3", sourceBucket.Region, time.Now()); err != nil {
		return nil, err
	}

	return clone, nil
}
