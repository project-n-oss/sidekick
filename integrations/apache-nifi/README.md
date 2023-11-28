# Apache NiFi integration

You can integrate Sidekick with Apache NiFi by overriding the endpoint URL in the `*S3` processors.

This has been validated only against Granica Crunch running in AWS but Granica Crunch in GCP should work as well.

## VPC Peering

You need to allow vpc peering from your application vpc to the Granica vpc. You can follow the tutorial [here](https://granica.ai/docs/vpc-peering/) to do so.

## Configuration

### Sidekick

In order for Sidekick to work with Apache NiFi you need to run it with the `--aws-ignore-auth-header-region` flag.

If not running on an ec2 or google compute engine instance in the same region as the Granica Crunch cluster, you'll also need to set the `GRANICA_REGION` environment variable to the region of the Granica Crunch cluster.

When running with the endpoint URL override, the underlying AWS SDKs that Apache NiFi uses sign requests targeted at `us-east-1`, however, this is not always desired. Using the `--aws-ignore-auth-header-region` flag will prevent Sidekick from inferring the bucket region from the `Authorization` header and instead cause it to use the region of the underlying ec2/google compute engine instance or the value of the `GRANICA_REGION` environment variable.

### Apache NiFi Processors

Set the `Endpoint Override URL` property to the Sidekick address. This is usually `http://localhost:7075`.

## Limitations

When running Sidekick with the `--aws-ignore-auth-header-region` flag accessing buckets in regions different than the Granica Crunch region is not possible.
