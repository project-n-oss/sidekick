package boltrouter

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
func extractSourceBucket(ctx context.Context, req *http.Request) (SourceBucket, error) {
	region, err := getRegionForBucket(ctx, req.Header.Get("Authorization"))
	if err != nil {
		return SourceBucket{}, fmt.Errorf("could not get region for bucket: %w", err)
	}

	ret := SourceBucket{
		Region: region,
		Bucket: "n-auth-dummy",
		Style:  nAuthDummy,
	}

	isVirtualHostedStyle := false
	split := strings.Split(req.Host, ".")
	if len(split) > 1 {
		if _, err := strconv.Atoi(split[0]); err != nil {
			// is not a number, so it is a bucket name
			isVirtualHostedStyle = true
		}
	}

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

	return ret, nil
}

var credentialRegexp = regexp.MustCompile(`Credential=([^,]*)`)

func getRegionForBucket(ctx context.Context, authHeader string) (string, error) {
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
