![projectn-sidekick.png](projectn-sidekick.png)
# Sidekick

Sidekick a [sidecar](https://learn.microsoft.com/en-us/azure/architecture/patterns/sidecar) ([ambassador](https://learn.microsoft.com/en-us/azure/architecture/patterns/ambassador)) process that allows your applications to talk with a Bolt cluster through any AWS SDK.

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

This will run sidekick localy on your machine on `localhost:8081`.

run the following command to learn more about the options:

```bash
go run main.go serve --help
```

### Docker

Todo

## Using Sidekick

In order to use sidekick with your aws sdk, you need to update the S3 Client hostname to point to the sidekick url (ex: `localhost:8081`). 

Currently you also need to set your s3 client to use `pathStyle` to work.

### Go 

```Go
sidekickURL := "http://localhost:8081"
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
cfg, _:= config.LoadDefaultConfig(ctx, config.WithEndpointResolverWithOptions(customResolver))
s3c := s3.NewFromConfig(cfg, func(options *s3.Options) {
    options.UsePathStyle = true
})
```
