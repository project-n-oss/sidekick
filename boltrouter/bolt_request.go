package boltrouter

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"go.uber.org/zap"
)

type ErrUnknownCloudPlatform error

type BoltRequest struct {
	Bolt    *http.Request
	Aws     *http.Request
	Gcp     *http.Request
	crcHash uint32
}

type BoltRequestAnalytics struct {
	ObjectKey                     string
	RequestBodySize               uint32
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

// NewBoltRequest transforms the passed in intercepted aws or gcp http.Request and returns
// a new http.Request Ready to be sent to Bolt.
// This new http.Request is routed to the correct Bolt endpoint and signed correctly.
func (br *BoltRouter) NewBoltRequest(ctx context.Context, logger *zap.Logger, req *http.Request) (*BoltRequest, error) {
	var boltRequest *BoltRequest
	var err error

	switch br.config.CloudPlatform {
	case AwsCloudPlatform:
		boltRequest, err = br.newBoltRequestForAws(ctx, logger, req)
	case GcpCloudPlatform:
		boltRequest, err = br.newBoltRequestForGcp(ctx, logger, req)
	}
	return boltRequest, err
}

func (br *BoltRouter) newBoltRequestForAws(ctx context.Context, logger *zap.Logger, req *http.Request) (*BoltRequest, error) {
	sourceBucket, err := extractSourceBucket(logger, req, br.boltVars.Region.Get())
	if err != nil {
		return nil, fmt.Errorf("could not extract source bucket: %w", err)
	}

	awsCred, err := getAwsCredentialsFromRegion(ctx, sourceBucket.Region)
	if err != nil {
		return nil, fmt.Errorf("could not get aws credentials: %w", err)
	}

	awsRequest, err := newFailoverAwsRequest(ctx, req, awsCred, sourceBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to make failover request: %w", err)
	}

	authPrefix := randString(4)
	headReq, err := signedAwsHeadRequest(ctx, req, awsCred, sourceBucket.Bucket, sourceBucket.Region, authPrefix)
	if err != nil {
		return nil, fmt.Errorf("could not make signed aws head request: %w", err)
	}
	bucketAndObjPath := ""
	if sourceBucket.Bucket == "n-auth-dummy" {
		// Special Case to handle dummy bucket
		bucketAndObjPath, _ = url.JoinPath(sourceBucket.Bucket, req.URL.Path)
	} else {
		bucketAndObjPath = req.URL.Path
	}

	crcHash := crc32.ChecksumIEEE([]byte(bucketAndObjPath))
	BoltURL, err := br.SelectBoltEndpoint(req.Method)

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
	if br.config.Local {
		req.URL.Scheme = "http"
	} else {
		req.URL.Scheme = "https"
	}
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

	boltRequest := &BoltRequest{
		Bolt:    req.Clone(ctx),
		Aws:     awsRequest.Clone(ctx),
		Gcp:     nil,
		crcHash: crcHash,
	}
	return boltRequest, nil
}

func (br *BoltRouter) newBoltRequestForGcp(ctx context.Context, logger *zap.Logger, req *http.Request) (*BoltRequest, error) {
	BoltURL, err := br.SelectBoltEndpoint(req.Method)
	if err != nil {
		return nil, err
	}

	req.RequestURI = ""
	logger.Debug("req.URL.Path", zap.String("path", req.URL.Path))
	logger.Debug("req.URL.RawPath", zap.String("path", req.URL.RawPath))
	logger.Debug("req.URL.EscapedPath", zap.String("path", req.URL.EscapedPath()))

	escapedPath, escapePathErr := escapeGCSPath(req.URL.Path, logger)
	logger.Debug("escapedPath", zap.String("path", escapedPath))
	if escapePathErr != nil {
		logger.Debug("Error escaping path, using original path")
		BoltURL.Path = req.URL.Path
	} else {
		BoltURL.Path = escapedPath
	}
	// BoltURL.Path = req.URL.EscapedPath()

	logger.Debug("BoltURL.Path", zap.String("path", BoltURL.Path))

	// if !strings.HasPrefix(BoltURL.Path, "/") {
	// 	BoltURL.Path = "/" + BoltURL.Path
	// }

	// BoltURL.RawQuery = req.URL.RawQuery
	// req.URL = BoltURL
	// req.Header.Set("Host", br.boltVars.BoltHostname.Get())
	// req.Host = br.boltVars.BoltHostname.Get()

	// req.Header.Set("User-Agent", fmt.Sprintf("%s%s", br.boltVars.UserAgentPrefix.Get(), req.Header.Get("User-Agent")))
	// req.Header.Set("X-Bolt-Availability-Zone", br.boltVars.ZoneId.Get())
	// if !br.config.Passthrough {
	// 	req.Header.Set("X-Bolt-Passthrough-Read", "disable")
	// }

	logger.Debug("req URI", zap.String("uri", req.RequestURI))
	logger.Debug("req URL.URI", zap.String("url", req.URL.RequestURI()))

	gcpRequest := req.Clone(ctx)
	gcsUrl, _ := url.Parse("https://storage.googleapis.com")
	gcsUrl.Path = req.URL.Path
	gcsUrl.RawPath = req.URL.RawPath
	// if escapePathErr != nil {
	// 	logger.Debug("gcp: Error escaping path, using original path")
	// 	gcsUrl.Path = req.URL.Path
	// } else {
	// 	gcsUrl.Path = escapedPath
	// }

	logger.Debug("gcsUrl.Path", zap.String("path", gcsUrl.Path))

	logger.Debug("gcsRequest URI", zap.String("uri", gcpRequest.RequestURI))
	logger.Debug("gcsRequest URL.URI", zap.String("url", gcpRequest.URL.RequestURI()))

	// if !strings.HasPrefix(gcsUrl.Path, "/") {
	// 	gcsUrl.Path = "/" + gcsUrl.Path
	// }

	gcsUrl.RawQuery = req.URL.RawQuery
	gcpRequest.URL = gcsUrl
	gcpRequest.Header.Set("Host", "storage.googleapis.com")
	gcpRequest.Host = "storage.googleapis.com"

	CopyReqBody(req, gcpRequest)

	boltRequest := &BoltRequest{
		Bolt:    req.Clone(ctx),
		Aws:     nil,
		Gcp:     gcpRequest.Clone(ctx),
		crcHash: 0,
	}
	return boltRequest, nil
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
// DoRequest will failover to AWS if the Bolt request panics for any reason
// DoboltRequest will return a bool indicating if the request was a failover.
// DoRequest will return a BoltRequestAnalytics struct with analytics about the request.
func (br *BoltRouter) DoRequest(logger *zap.Logger, boltReq *BoltRequest) (*http.Response, bool, *BoltRequestAnalytics, error) {
	initialRequestTarget, reason, err := br.SelectInitialRequestTarget(boltReq)
	boltRequestAnalytics := &BoltRequestAnalytics{
		ObjectKey:                     boltReq.Bolt.URL.Path,
		RequestBodySize:               uint32(boltReq.Bolt.ContentLength),
		Method:                        boltReq.Bolt.Method,
		InitialRequestTarget:          InitialRequestTargetMap[initialRequestTarget],
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

	logger.Debug("initial request target", zap.String("target", InitialRequestTargetMap[initialRequestTarget]), zap.String("reason", reason))

	initialRequestTarget = InitialRequestTargetFallback
	if initialRequestTarget == InitialRequestTargetBolt {
		resp, isFailoverRequest, err := br.doBoltRequest(logger, boltReq, false, boltRequestAnalytics)
		// if nothing during br.doBoltRequest panics, err will not be of type ErrPanicDuringBoltRequest so failover was
		// handled inside the function as needed, and we can just return
		// If the err is of type ErrPanicDuringBoltRequest then we need to failover to AWS manually since .doBoltRequest
		// halted execution before it could failover
		if err != nil && errors.Is(err, ErrPanicDuringBoltRequest) && br.config.Failover {
			switch br.config.CloudPlatform {
			case AwsCloudPlatform:
				logger.Error("panic occurred during Bolt request, failing over to AWS", zap.Error(err))
				resp, isFailoverRequest, err = br.doAwsRequest(logger, boltReq, true, boltRequestAnalytics)
			case GcpCloudPlatform:
				logger.Error("panic occurred during Bolt request, failing over to GCP", zap.Error(err))
				resp, isFailoverRequest, err = br.doGcpRequest(logger, boltReq, true, boltRequestAnalytics)
			}
		}
		return resp, isFailoverRequest, boltRequestAnalytics, err
	} else {
		var resp *http.Response
		var isFailoverRequest bool
		var err error

		switch br.config.CloudPlatform {
		case AwsCloudPlatform:
			resp, isFailoverRequest, err = br.doAwsRequest(logger, boltReq, false, boltRequestAnalytics)
		case GcpCloudPlatform:
			resp, isFailoverRequest, err = br.doGcpRequest(logger, boltReq, false, boltRequestAnalytics)
		}
		return resp, isFailoverRequest, boltRequestAnalytics, err
	}
}

// doBoltRequest sends an HTTP Bolt request and returns an HTTP response, following policy (such as redirects, cookies, auth) as configured on the client.
// doBoltRequest will failover to AWS if the Bolt request fails, config.Failover is set to true, and the request itself is not a failover request.
// doBoltRequest will return a bool indicating if the request was a failover.
// doBoltRequest will return a BoltRequestAnalytics struct with analytics about the request.
// doBoltRequest will catch any panics and return with error ErrPanicDuringBoltRequest
func (br *BoltRouter) doBoltRequest(logger *zap.Logger, boltReq *BoltRequest, isFailover bool, analytics *BoltRequestAnalytics) (resp *http.Response, isFailoverRequest bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic occurred during Bolt request", zap.Any("panic", r))
			resp = nil
			err = ErrPanicDuringBoltRequest
		}
	}()

	beginTime := time.Now()
	resp, err = br.boltHttpClient.Do(boltReq.Bolt)
	duration := time.Since(beginTime)
	analytics.BoltRequestDuration = duration

	statusCode := -1
	if resp != nil {
		statusCode = resp.StatusCode
	}

	logger.Debug("bolt resp",
		zap.Any("code", statusCode),
		zap.Bool("failover", br.config.Failover),
		zap.Bool("isFailover", isFailover),
		zap.Error(err))

	if err != nil {
		br.MaybeMarkOffline(boltReq.Bolt.URL, err)
	}

	if !isFailover {
		failover := false

		// Fallback on 404 errors
		if !br.config.NoFallback404 && statusCode == 404 {
			failover = true
		} else if br.config.Failover &&
			(err != nil || !StatusCodeIs2xx(statusCode)) {
			// Attempt to failover on error or based on response status code
			failover = true
		}

		if failover {
			if resp != nil {
				b, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				logger.Error("bolt request failed", zap.Int("code", statusCode), zap.String("body", string(b)))
			} else {
				logger.Error("bolt request failed", zap.Error(err))
			}

			switch br.config.CloudPlatform {
			case AwsCloudPlatform:
				return br.doAwsRequest(logger, boltReq, true, analytics)
			case GcpCloudPlatform:
				return br.doGcpRequest(logger, boltReq, true, analytics)
			default:
				return nil, false, fmt.Errorf("unknown cloud platform")
			}
		}
	}

	if resp != nil {
		analytics.BoltRequestResponseStatusCode = statusCode
	}
	return resp, isFailover, err
}

// doAwsRequest sends an HTTP Bolt request and returns an HTTP response, following policy (such as redirects, cookies, auth) as configured on the client.
// doAwsRequest will failover to Bolt if the AWS request fails, response status code is not 404, and the request itself is not a failover request.
// doAwsRequest will return a bool indicating if the request was a failover.
// doAwsRequest will return a BoltRequestAnalytics struct with analytics about the request.
// doAwsRequest will catch any panics and return with error ErrPanicDuringAwsRequest
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

	statusCode := -1
	if resp != nil {
		statusCode = resp.StatusCode
	}

	logger.Debug("aws resp",
		zap.Any("code", statusCode),
		zap.Bool("failover", br.config.Failover),
		zap.Bool("isFailover", isFailover),
		zap.Error(err))

	if !isFailover {

		// Fallback on 404 errors
		// For other AWS errors, we will return that error back to client to retry as necessary.
		if !br.config.NoFallback404 && resp != nil && statusCode == 404 {
			logger.Info("aws request failed, falling back to bolt on 404")

			return br.doBoltRequest(logger, boltReq, true, analytics)
		}
	}

	if resp != nil {
		analytics.AwsRequestResponseStatusCode = statusCode
	}
	return resp, isFailover, err
}

func (br *BoltRouter) doGcpRequest(logger *zap.Logger, boltReq *BoltRequest, isFailover bool, analytics *BoltRequestAnalytics) (resp *http.Response, isFailoverRequest bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic occurred during GCP request", zap.Any("panic", r))
			resp = nil
			err = ErrPanicDuringAwsRequest
		}
	}()

	beginTime := time.Now()
	resp, err = br.gcpHttpClient.Do(boltReq.Gcp)
	duration := time.Since(beginTime)
	analytics.AwsRequestDuration = duration

	statusCode := -1
	if resp != nil {
		statusCode = resp.StatusCode
	}

	logger.Debug("gcp resp",
		zap.Any("code", statusCode),
		zap.Bool("failover", br.config.Failover),
		zap.Bool("isFailover", isFailover),
		zap.Error(err))

	// if !isFailover {
	// 	// Fallback on 404 errors
	// 	// For other AWS errors, we will return that error back to client to retry as necessary.
	// 	if !br.config.NoFallback404 && resp != nil && statusCode == 404 {
	// 		logger.Info("gcp request failed, falling back to bolt on 404")

	// 		return br.doGcpRequest(logger, boltReq, true, analytics)
	// 	}
	// }

	if resp != nil {
		analytics.AwsRequestResponseStatusCode = statusCode
	}
	return resp, isFailover, err
}

func StatusCodeIs2xx(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// escapeGCSPath escapes special characters in the Google Cloud Storage object path.
//
// According to Google Cloud Storage documentation, certain characters in the object name
// or query string of a request URI must be percent-encoded to ensure compatibility across
// Cloud Storage tools. This includes characters such as !, #, $, &, ', (, ), *, +, ,, /, :, ;, =, ?, @, [, ], and space.
//
// This function specifically focuses on encoding the object name part of the URI path that follows
// after "/o/". It splits the given path at "/o/", escapes the object name part, and then reassembles the path.
// This ensures that special characters in the object name are correctly percent-encoded as per the standards
// outlined in RFC 3986, Section 3.3.
//
// Example:
// For an object named "foo??bar" in the bucket "example-bucket", the path "/b/example-bucket/o/foo??bar"
// would be transformed to "/b/example-bucket/o/foo%3F%3Fbar".
//
// https://cloud.google.com/storage/docs/request-endpoints#encoding
func escapeGCSPath(path string, logger *zap.Logger) (string, error) {
	parts := strings.SplitN(path, "/o/", 2)
	if len(parts) != 2 {
		return path, fmt.Errorf("path does not contain '/o/'")
	}

	basePart := parts[0] + "/o/"
	pathPart := parts[1]

	unescapedPathPart, err := url.QueryUnescape(pathPart)
	if err != nil {
		logger.Debug("Error unescaping path, using original path")
		// If there's an error in unescaping, it means the path wasn't escaped
		unescapedPathPart = pathPart
	}

	reEscapedPathPart := url.PathEscape(unescapedPathPart)
	if reEscapedPathPart == pathPart {
		logger.Debug("Path was already escaped")
		return basePart + pathPart, nil
	}

	return basePart + reEscapedPathPart, nil
}
