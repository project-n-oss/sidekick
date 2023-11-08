# Dataproc integration

You can integrate sidekick with Dataproc by adding the [sidekick_service_init.sh](./sidekick_service_init.sh) as an [initialization action](https://cloud.google.com/dataproc/docs/concepts/configuring-clusters/init-actions) in your cluster.

## VPC Peering

You need to allow vpc peering from your Dataproc vpc to the Granica vpc. You can follow the tutorial [here](https://xyz.projectn.co/vpc-peering) to do so.

## Configuration

In order for sidekick to work with you clusters, you need define the `GRANICA_CUSTOM_DOMAIN`, `GRANICA_CLOUD_PLATFORM` environment variables.

### Init script configuration

You can define the environment variables by adding these lines to the [sidekick service init script](./sidekick_service_init.sh):

```bash
export GRANICA_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN>
export GRANICA_CLOUD_PLATFORM="<aws|gcp>"
```
