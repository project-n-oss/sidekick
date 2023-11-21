package boltrouter

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type s3RequestStyle string

const (
	virtualHostedStyle s3RequestStyle = "virtual-hosted-style"
	pathStyle          s3RequestStyle = "path-style"
	nAuthDummy         s3RequestStyle = "n-auth-dummy"
)

type SourceBucket struct {
	Bucket string
	Region string
	Style  s3RequestStyle
}

// extractSourceBucket extracts the aws request bucket using Path-style or Virtual-hosted-style requests.
// https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html
// This method will "n-auth-dummy" if nothing is found
func extractSourceBucket(logger *zap.Logger, req *http.Request, defaultRegionFallback string, ignoreAuthHeaderRegion bool) (SourceBucket, error) {
	region, err := getRegionForBucket(req.Header.Get("Authorization"))
	if err != nil {
		return SourceBucket{}, fmt.Errorf("could not get region for bucket: %w", err)
	}
	logger.Debug("extracted region", zap.String("region", region))

	if region == "" {
		logger.Warn("could not get region from auth header, using default region fallback", zap.String("defaultRegionFallback", defaultRegionFallback))
		region = defaultRegionFallback
	}

	if ignoreAuthHeaderRegion {
		region = defaultRegionFallback
	}

	ret := SourceBucket{
		Region: region,
		Bucket: "n-auth-dummy",
		Style:  nAuthDummy,
	}

	isVirtualHostedStyle := false
	split := strings.Split(req.Host, ".")
	logger.Debug("split host", zap.Strings("split", split))
	// assuming running on http://sidekick.us-west-2.km-nov21-aws.bolt.projectn.co
	if len(split) > 6 {
		if _, err := strconv.Atoi(split[0]); err != nil {
			// is not a number, so it is a bucket name
			isVirtualHostedStyle = true
		}
	}
	logger.Debug("isVirtualHostedStyle", zap.Bool("isVirtualHostedStyle", isVirtualHostedStyle))

	if isVirtualHostedStyle {
		bucket := split[0]
		ret.Bucket = bucket
		ret.Style = virtualHostedStyle
	} else if paths := strings.Split(req.URL.EscapedPath(), "/"); len(paths) > 1 {
		// path-style request
		bucket := paths[1]
		ret.Bucket = bucket
		ret.Style = pathStyle
	}

	logger.Debug("extracted source bucket", zap.String("bucket", ret.Bucket), zap.String("style", string(ret.Style)), zap.String("region", ret.Region))

	return ret, nil
}

var credentialRegexp = regexp.MustCompile(`Credential=([^,]*)`)

func getRegionForBucket(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("no auth header in request, cannot extract region")
	}

	matches := credentialRegexp.FindStringSubmatch(authHeader)
	if len(matches) != 2 {
		return "", fmt.Errorf("could not extract credential from auth header, matches: %v", matches)
	}

	// format: AKIA3Y7DLM2EYWSYCN5P/20230511/us-east-1/s3/aws4_request
	credentialStr := matches[1]
	credentialSplit := strings.Split(credentialStr, "/")
	if len(credentialSplit) != 5 {
		return "", fmt.Errorf("could not extract region from credential, credential: %v", credentialStr)
	}
	region := credentialSplit[2]
	return region, nil
}
