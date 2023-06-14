package aws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// SHA value for empty payload. As head object request is with empty payload
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/signer/v4#Signer.SignHTTP
const emptyPayloadHash string = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// signedAwsHeadRequest returns a new Head http.Request signed by AWS v4 signer.
func SignedAwsHeadRequest(ctx context.Context, req *http.Request, awsCred aws.Credentials, sourceBucket string, region string, authPrefix string) (*http.Request, error) {
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
