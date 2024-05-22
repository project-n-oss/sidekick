# Dataproc integration

You can integrate Sidekick with Dataproc by adding the [sidekick_service_init.sh](./sidekick_service_init.sh) as an [initialization action](https://cloud.google.com/dataproc/docs/concepts/configuring-clusters/init-actions) in your cluster. You'll want to download the init script, edit it to your needs based on the instructions below and make it available to Dataproc in a Google Cloud Storage bucket in your project.


## Configuration

In order for Sidekick to work with you clusters, you need define the `SIDEKICK_APP_CLOUDPLATFORM` environment variable in the init script since Dataproc doesn't have first-class support for OS level custom environment variables which apply to all processes.

### Init script configuration

Define the environment variables by adding these lines to the [sidekick service init script](./sidekick_service_init.sh):

```bash
export SIDEKICK_APP_CLOUDPLATFORM=<AWS|GCP>
```
