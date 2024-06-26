package app

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	sidekickAws "github.com/project-n-oss/sidekick/app/aws"
)

// DoRequest makes a request to the cloud platform
// Does a request to the source bucket and if it returns 404, tries the crunched bucket
// Returns the response and a boolean indicating if the response is from the crunched bucket
func (sess *Session) DoRequest(req *http.Request) (*http.Response, bool, error) {
	switch sess.app.cfg.CloudPlatform {
	case AwsCloudPlatform.String():
		return sess.DoAwsRequest(req)
	default:
		return nil, false, fmt.Errorf("CloudPlatform %s not supported", sess.app.cfg.CloudPlatform)
	}
}

const crunchFileFoundErrStatus = "409 Src file not found, but crunched file also found"
const crunchFileFoundStatusCode = 409

// DoAwsRequest makes a request to AWS
// If a crunched version of the source file exists, returns a 500 response
// Returns the response and a boolean indicating if a crunched file was found
// You can disable this behavior by setting NoCrunchErr to true in the config
func (sess *Session) DoAwsRequest(req *http.Request) (*http.Response, bool, error) {
	sourceBucket, err := sidekickAws.ExtractSourceBucket(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to extract source bucket from request: %w", err)
	}

	cloudRequest, err := sidekickAws.NewRequest(sess.Context(), sess.Logger(), req, sourceBucket)
	if err != nil {
		return nil, false, fmt.Errorf("failed to make aws request: %w", err)
	}

	resp, err := http.DefaultClient.Do(cloudRequest)
	if err != nil {
		return nil, false, fmt.Errorf("failed to do aws request: %w", err)
	}

	// if the source file is not already a crunched file, check if the crunched file exists
	if !sess.app.cfg.NoCrunchErr && !isCrunchedFile(cloudRequest.URL.Path) {
		objectKey := makeCrunchFilePath(sourceBucket.Bucket, cloudRequest.URL.Path)

		s3Client, err := sidekickAws.GetS3ClientFromRegion(sess.Context(), sourceBucket.Region)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get s3 client for region '%s': %w", sourceBucket.Region, err)
		}

		// ignore errors, we only want to check if the object exists
		headResp, _ := s3Client.HeadObject(sess.Context(), &s3.HeadObjectInput{
			Bucket: aws.String(sourceBucket.Bucket),
			Key:    aws.String(objectKey),
		})

		// found crunched file, return 500 to client
		if headResp != nil && headResp.ETag != nil {
			resp.StatusCode = crunchFileFoundStatusCode
			resp.Status = crunchFileFoundErrStatus
		}
		return resp, true, nil
	}

	return resp, false, err
}
