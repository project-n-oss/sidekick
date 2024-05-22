package aws

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type s3RequestStyle string

const (
	// example: https://bucket-name.s3.region-code.amazonaws.com/key-name
	VirtualHostedStyle s3RequestStyle = "virtual-hosted-style"
	// example: https://s3.region-code.amazonaws.com/bucket-name/key-name
	PathStyle s3RequestStyle = "path-style"
)

type SourceBucket struct {
	Bucket string
	Region string
	Style  s3RequestStyle
}

// extractSourceBucket extracts the aws request bucket using Path-style or Virtual-hosted-style requests.
// https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html
// This method will error if the bucket cannot be extracted from the request.
func ExtractSourceBucket(req *http.Request) (SourceBucket, error) {
	region, err := getRegionForBucket(req.Header.Get("Authorization"))
	if err != nil {
		return SourceBucket{}, fmt.Errorf("could not get region for bucket: %w", err)
	}

	ret := SourceBucket{
		Region: region,
		Bucket: "",
		Style:  "",
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
		ret.Style = VirtualHostedStyle
	} else if paths := strings.Split(req.URL.EscapedPath(), "/"); len(paths) > 1 {
		// path-style request
		bucket := paths[1]
		ret.Bucket = bucket
		ret.Style = PathStyle
	} else {
		return SourceBucket{}, fmt.Errorf("could not extract bucket from request")
	}

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
