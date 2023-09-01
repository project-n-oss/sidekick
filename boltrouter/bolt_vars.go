package boltrouter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once        sync.Once
	instance    *BoltVars
	instanceErr error
)

// GetBoltVars acts as a singleton method wrapper around BoltVars.
// It guarantees that only one instance of BoltVars exists.
// This method is thread safe.
func GetBoltVars(ctx context.Context, logger *zap.Logger, cloudPlatform CloudPlatformType) (*BoltVars, error) {
	once.Do(func() {
		instance, instanceErr = newBoltVars(ctx, logger, cloudPlatform)
	})
	return instance, instanceErr
}

// BoltInfo is the type returned by quicksilver.
// It follows this schema:

//	{
//		"main_write_endpoints": [],
//		"failover_write_endpoints": [],
//		"main_read_endpoints": [],
//		"failover_read_endpoints": [],
//		"cluster_healthy": bool,
//		"client_behavior_params": {
//			"cleaner_on": bool
//			"crunch_traffic_percent": int
//		}
//	}
type BoltInfo map[string]interface{}

// BoltVars is a singleton struct keeping track of Bolt variables across threads.
// This is used in BoltRouter to route requests appropriately.
// You should only access BoltVars with GetBoltVars().
type BoltVars struct {
	ReadOrderEndpoints  AtomicVar[[]string]
	WriteOrderEndpoints AtomicVar[[]string]
	HttpReadMethodTypes AtomicVar[[]string]
	offlineEndpoints    map[string]bool
	livenessLock        sync.RWMutex

	Region           AtomicVar[string]
	ZoneId           AtomicVar[string]
	BoltCustomDomain AtomicVar[string]
	AuthBucket       AtomicVar[string]
	UserAgentPrefix  AtomicVar[string]
	BoltHostname     AtomicVar[string]
	QuicksilverURL   AtomicVar[string]
	BoltInfo         AtomicVar[BoltInfo]
}

func (bv *BoltVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	fields := reflect.TypeOf(bv).Elem()
	values := reflect.ValueOf(bv).Elem()
	num := fields.NumField()
	for i := 0; i < num; i++ {
		name := fields.Field(i).Name
		v := values.Field(i).Addr().MethodByName("String").Call([]reflect.Value{})[0].String()
		enc.AddString(name, v)
	}

	return nil
}

func newBoltVars(ctx context.Context, logger *zap.Logger, cloudPlatform CloudPlatformType) (*BoltVars, error) {
	logger.Debug("initializing BoltVars...")
	ret := &BoltVars{
		offlineEndpoints: make(map[string]bool),
	}

	ret.ReadOrderEndpoints.Set([]string{"main_read_endpoints", "main_write_endpoints", "failover_read_endpoints", "failover_write_endpoints"})
	ret.WriteOrderEndpoints.Set([]string{"main_write_endpoints", "failover_write_endpoints"})
	ret.HttpReadMethodTypes.Set([]string{http.MethodGet, http.MethodHead}) // S3 operations get converted to one of the standard HTTP request methods https://docs.aws.amazon.com/apigateway/latest/developerguide/integrating-api-with-aws-services-s3.html

	if cloudPlatform == AwsCloudPlatform {
		isEc2, err := isEc2Instance(ctx, logger)
		if err != nil {
			return nil, err
		}

		awsRegion, err := getAwsRegion(ctx, logger, isEc2)
		if err != nil {
			return nil, err
		}
		ret.Region.Set(awsRegion)

		awsZoneId, err := getAwsZoneID(ctx, logger, isEc2)
		if err != nil {
			return nil, err
		}
		ret.ZoneId.Set(awsZoneId)
	} else if cloudPlatform == GcpCloudPlatform {
		isComputeEngine, err := isComputeEngineInstance(ctx, logger)
		if err != nil {
			return nil, err
		}

		gcpRegion, err := getGcpRegion(ctx, logger, isComputeEngine)
		if err != nil {
			return nil, err
		}
		ret.Region.Set(gcpRegion)

		gcpZoneId, err := getGcpZone(ctx, logger, isComputeEngine)
		if err != nil {
			return nil, err
		}
		ret.ZoneId.Set(gcpZoneId)
	}

	var boltCustomDomain string
	boltCustomDomain, ok := os.LookupEnv("GRANICA_CUSTOM_DOMAIN")
	if !ok {
		// Keeping this for backwards compatibility
		boltCustomDomain, ok = os.LookupEnv("BOLT_CUSTOM_DOMAIN")
		if !ok {
			return nil, fmt.Errorf("GRANICA_CUSTOM_DOMAIN or BOLT_CUSTOM_DOMAIN env variable is not set")
		}
	}
	ret.BoltCustomDomain.Set(boltCustomDomain)
	ret.BoltHostname.Set(fmt.Sprintf("bolt.%s.%s", ret.Region.Get(), ret.BoltCustomDomain.Get()))

	ret.AuthBucket.Set(os.Getenv("BOLT_AUTH_BUCKET"))

	userAgentPrefix, ok := os.LookupEnv("USER_AGENT_PREFIX")
	if !ok {
		userAgentPrefix = "granica-sidekick/"
	} else {
		userAgentPrefix = fmt.Sprintf("%s/", userAgentPrefix)
	}
	ret.UserAgentPrefix.Set(userAgentPrefix)

	ret.QuicksilverURL.Set("")
	quicksilverURL := fmt.Sprintf("https://quicksilver.%s.%s/services/bolt", ret.Region.Get(), ret.BoltCustomDomain.Get())
	if ret.ZoneId.Get() != "" {
		quicksilverURL += fmt.Sprintf("?az=%s", ret.ZoneId.Get())
	}
	ret.QuicksilverURL.Set(quicksilverURL)

	ret.BoltInfo.Set(BoltInfo{})

	if logger.Level() == zap.DebugLevel {
		logger.Debug(fmt.Sprintf("done! BoltVars %v", ret))
	}

	return ret, nil
}

func getGcpRegion(ctx context.Context, logger *zap.Logger, isComputeEngine bool) (string, error) {
	ret, ok := os.LookupEnv("GRANICA_REGION")
	if ok {
		return ret, nil
	} else if !isComputeEngine {
		return "", fmt.Errorf("GRANICA_REGION env variable is not set")
	}

	zone, err := getGcpZone(ctx, logger, isComputeEngine)
	if err != nil {
		return "", err
	}
	split := strings.Split(zone, "-")
	if len(split) < 2 {
		return "", fmt.Errorf("could not parse gcp region from zone %s", zone)
	}
	// concat all but last item in split into a string separated by dashes
	retval := strings.Join(split[:len(split)-1], "-")
	return retval, nil
}

func getGcpZone(ctx context.Context, logger *zap.Logger, isComputeEngine bool) (string, error) {
	ret, ok := os.LookupEnv("GCP_ZONE")
	if ok {
		return ret, nil
	} else if !isComputeEngine {
		return "", nil
	}

	zone, err := metadata.Zone()
	if err != nil {
		return "", fmt.Errorf("could not get gcp zone from metadata service: %w", err)
	}
	return zone, nil
}

func getAwsRegion(ctx context.Context, logger *zap.Logger, isEc2 bool) (string, error) {
	ret, ok := os.LookupEnv("GRANICA_REGION")
	if ok {
		return ret, nil
	} else if !isEc2 {
		return "", fmt.Errorf("GRANICA_REGION env variable is not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("could not load aws default config: %w", err)
	}

	client := imds.NewFromConfig(cfg)

	output, err := client.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return "", fmt.Errorf("could not get aws region from ec2 metadata service: %w", err)
	}

	return output.Region, nil
}

func getAwsZoneID(ctx context.Context, logger *zap.Logger, isEc2 bool) (string, error) {
	ret := os.Getenv("AWS_ZONE_ID")
	if len(ret) > 0 {
		return ret, nil
	} else if !isEc2 {
		return "", nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("could not load aws default config: %w", err)
	}

	client := imds.NewFromConfig(cfg)

	output, err := client.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "placement/availability-zone-id",
	})
	if err != nil {
		return "", fmt.Errorf("could not get aws zone id from ec2 metadata service: %w", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Content)
	ret = buf.String()

	return ret, nil
}

func isEc2Instance(ctx context.Context, logger *zap.Logger) (bool, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return false, fmt.Errorf("could not load aws default config: %w", err)
	}
	client := imds.NewFromConfig(cfg)
	_, err = client.GetMetadata(ctx, &imds.GetMetadataInput{})
	if err != nil {
		logger.Warn("not running on ec2 instance", zap.Error(err))
		return false, nil
	}

	return true, nil
}

func isComputeEngineInstance(ctx context.Context, logger *zap.Logger) (bool, error) {
	_, err := metadata.ProjectID()
	if err != nil {
		logger.Warn("not running on compute engine instance", zap.Error(err))
		return false, nil
	}

	return true, nil
}
