# Apache NiFi integration

Integrate Apache NiFi with Sidekick by overriding the endpoint URL in the `*S3` processors. This integration has been validated with Granica Crunch running in AWS. It should also work with Granica Crunch in GCP.

## VPC Peering

You need to allow vpc peering from your application vpc to the Granica vpc. You can follow the tutorial [here](https://granica.ai/docs/vpc-peering/) to do so.

## Configuration

### Sidekick

To ensure compatibility with Apache NiFi, run Sidekick with the --aws-ignore-auth-header-region flag.

If you're not running on an EC2 or Google Compute Engine instance in the same region as the Granica Crunch cluster, set the GRANICA_REGION environment variable to the Granica Crunch cluster's region.

The --aws-ignore-auth-header-region flag is crucial when using endpoint URL overrides. Without it, the AWS SDKs used by Apache NiFi default to signing requests for us-east-1, which may not always be appropriate. This flag instructs Sidekick to ignore the region specified in the Authorization header and to use the region of the underlying EC2/Google Compute Engine instance, or the GRANICA_REGION environment variable, instead.

### Apache NiFi Processors

Configure the Endpoint Override URL property in your Apache NiFi processors to point to the Sidekick address, typically http://localhost:7075.

## Limitations

Be aware that when running Sidekick with the --aws-ignore-auth-header-region flag, accessing buckets in regions different from the Granica Crunch region is not supported.
