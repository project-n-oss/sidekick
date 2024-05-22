package app

import "fmt"

type CloudPlatformType string

const (
	AwsCloudPlatform CloudPlatformType = "AWS"
	GcpCloudPlatform CloudPlatformType = "GCP"
)

func (c CloudPlatformType) String() string {
	return string(c)
}

type Config struct {
	CloudPlatform string `yaml:"CloudPlatform"`
	NoCrunchErr   bool   `yaml:"NoCrunchErr"`
}

func (c Config) Validate() error {
	if c.CloudPlatform != AwsCloudPlatform.String() && c.CloudPlatform != GcpCloudPlatform.String() {
		return fmt.Errorf("CloudPlatform must be one of: %s, %s", AwsCloudPlatform, GcpCloudPlatform)
	}

	return nil
}
