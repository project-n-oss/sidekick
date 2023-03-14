package boltrouter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"

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
// It garantees that only one instance of BoltVars exists.
// This method is thread safe.
func GetBoltVars(logger *zap.Logger) (*BoltVars, error) {
	once.Do(func() {
		instance, instanceErr = newBoltVars(logger)
	})
	return instance, instanceErr
}

// BoltEndpointsMap is the type returned by quicksilver.
// It follows this schema:
//
//	{
//	  "main_write_endpoints": [],
//	  "failover_write_endpoints": [],
//	  "main_read_endpoints": [],
//	  "failover_read_endpoints": []
//	}
type BoltEndpointsMap map[string][]string

// BoltVars is a singleton struct keeping track of Bolt variables accross threads.
// This is used in BoltRouter to route requests appropriately.
// You should only access BoltVars with GetBoltVars().
type BoltVars struct {
	ReadOrderEndpoints  AtomicVar[[]string]
	WriteOrderEndpoints AtomicVar[[]string]
	HttpReadMethodTypes AtomicVar[[]string]

	Region           AtomicVar[string]
	ZoneId           AtomicVar[string]
	BoltCustomDomain AtomicVar[string]
	AuthBucket       AtomicVar[string]
	UserAgentPrefix  AtomicVar[string]
	BoltHostname     AtomicVar[string]
	QuicksilverURL   AtomicVar[string]
	BoltEndpoints    AtomicVar[BoltEndpointsMap]
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

func newBoltVars(logger *zap.Logger) (*BoltVars, error) {
	logger.Debug("initializing BoltVars...")
	ret := &BoltVars{}

	ret.ReadOrderEndpoints.Set([]string{"main_read_endpoints", "main_write_endpoints", "failover_read_endpoints", "failover_write_endpoints"})
	ret.WriteOrderEndpoints.Set([]string{"main_write_endpoints", "failover_write_endpoints"})
	ret.HttpReadMethodTypes.Set([]string{http.MethodGet, http.MethodHead}) // S3 operations get converted to one of the standard HTTP request methods https://docs.aws.amazon.com/apigateway/latest/developerguide/integrating-api-with-aws-services-s3.html

	awsRegion, err := getAwsRegion(logger)
	if err != nil {
		return nil, err
	}
	ret.Region.Set(awsRegion)

	awsZoneId, err := getAwsZoneID(logger)
	if err != nil {
		return nil, err
	}
	ret.ZoneId.Set(awsZoneId)

	boltCustomDomain, ok := os.LookupEnv("BOLT_CUSTOM_DOMAIN")
	if !ok {
		return nil, fmt.Errorf("BOLT_CUSTOM_DOMAIN env variable is not set")
	}
	ret.BoltCustomDomain.Set(boltCustomDomain)
	ret.BoltHostname.Set(fmt.Sprintf("bolt.%s.%s", ret.Region.Get(), ret.BoltCustomDomain.Get()))

	ret.AuthBucket.Set(os.Getenv("BOLT_AUTH_BUCKET"))

	userAgentPrefix, ok := os.LookupEnv("USER_AGENT_PREFIX")
	if !ok {
		userAgentPrefix = "projectn/"
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

	ret.BoltEndpoints.Set(BoltEndpointsMap{})

	logger.Debug("done!", zap.Object("BoltVars", ret))
	return ret, nil
}

func getAwsRegion(logger *zap.Logger) (string, error) {
	ret, ok := os.LookupEnv("AWS_REGION")
	if ok {
		return ret, nil
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", nil
	}

	client := imds.NewFromConfig(cfg)

	output, err := client.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return "", nil
	}

	return output.Region, nil
}

func getAwsZoneID(logger *zap.Logger) (string, error) {
	ret := os.Getenv("AWS_ZONE_ID")
	if len(ret) > 0 {
		return ret, nil
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", nil
	}

	client := imds.NewFromConfig(cfg)

	output, err := client.GetMetadata(ctx, &imds.GetMetadataInput{
		Path: "placement/availability-zone-id",
	})
	if err != nil {
		return "", nil
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Content)
	ret = buf.String()

	return ret, nil
}
