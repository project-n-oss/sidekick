![release](https://img.shields.io/github/v/release/project-n-oss/sidekick)

![projectn-sidekick.png](projectn-sidekick.png)
# Sidekick

Sidekick is a [sidecar](https://learn.microsoft.com/en-us/azure/architecture/patterns/sidecar) proxy process that allows your applications to talk with a Bolt cluster through any AWS SDK.

## Getting started

- go1.20
- A bolt cluster
- A cloud instance that has a vpc connection with the bolt cluster

## Running Sidekick

### Env Variables

In order to run sidekick, you first need to set some ENV variables

```bash
export BOLT_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN>
# Optional if not running on a ec2 instance or running in a different region
export AWS_REGION=<YOUR_BOLT_CLUSTER_REGION>
# Optional if not running on a ec2 instance to force read from a read-replica in this az 
export AWS_ZONE_ID=<AWS_ZONE_ID> 
```

### Local

You can run sidekick directly from the command line:

```bash
go run main.go serve
```

This will run sidekick localy on your machine on `localhost:7075`.

run the following command to learn more about the options:

```bash
go run main.go serve --help
```

### Docker

Build the docker image:

```bash
docker build -t sidekick .
```

running:

```bash
docker run -p 7075:7075 --env BOLT_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN> -env AWS_REGION=<YOUR_BOLT_CLUSTER_REGION> sidekick serve
```

## Using Sidekick

In order to use sidekick with your aws sdk, you need to update the S3 Client hostname to point to the sidekick url (ex: `localhost:7075`). 
Currently you also need to set your s3 client to use `pathStyle` to work.

### AWS cli

```bash
aws s3api get-object --bucket <YOUR_BUCKET> --key <YOUR_OBJECT_KEY>  delete_me.csv --endpoint-url http://localhost:7075
```

### Go 

```Go
func main() {
	ctx := context.Background()
	sidekickURL := "http://localhost:7075"
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           sidekickURL,
				SigningRegion: region,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	cfg, _ := config.LoadDefaultConfig(ctx, config.WithEndpointResolverWithOptions(customResolver))
	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	awsResp, err := s3c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("foo"),
		Key:    aws.String("bar"),
	})
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(awsResp.Body)
	awsResp.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

```

## Contributing

### Versionning

This repository uses [release-please](https://github.com/google-github-actions/release-please-action) to create and manage release.

### Commits

We follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for our commits and PR titles. This allows us to use release-please to manage our releases.

The most important prefixes you should have in mind are:

- fix: which represents bug fixes, and correlates to a SemVer patch.
- feat: which represents a new feature, and correlates to a SemVer minor.
- feat!:, or fix!:, refactor!:, etc., which represent a breaking change (indicated by the !) and will result in a SemVer major.