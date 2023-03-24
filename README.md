![projectn-sidekick.png](projectn-sidekick.png)
# Sidekick

Sidekick is a [sidecar](https://learn.microsoft.com/en-us/azure/architecture/patterns/sidecar) proxy process that allows your applications to talk with a Bolt cluster through any AWS SDK.

## Getting started

- go1.20
- A bolt cluster
- A cloud instance that has a vpc connection with the bolt cluster

## Running Sidekick

### Local

You can run sidekick directly from the command line:

```bash
go run main.go serve
```

This will run sidekick localy on your machine on `localhost:7071`.

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
docker run -p 7071:7071 --env BOLT_CUSTOM_DOMAIN=rvh.bolt.projectn.co sidekick serve
```


## Using Sidekick

In order to use sidekick with your aws sdk, you need to update the S3 Client hostname to point to the sidekick url (ex: `localhost:8081`). 

Currently you also need to set your s3 client to use `pathStyle` to work.

### AWS cli

```bash
aws s3api get-object --bucket <YOUR_BUCKET> --key <YOUR_OBJECT_KEY>  delete_me.csv --endpoint-url http://localhost:7071
aws s3api get-object --bucket sidekick-test-rvh2 --key animals/1.csv delete_me.csv --endpoint-url http://localhost:7071
```

### Go 

```Go
func main() {
	ctx := context.Background()
	sidekickURL := "http://localhost:7071"
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
	s3c := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = true
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
