# Databricks integration

You can integrate sidekick with databricks by adding the [sidekick_service_init.sh](./sidekick_service_init.sh) as an [init script](https://docs.databricks.com/clusters/init-scripts.html) in your clusters/workspace. Both Global and Cluster based init scripts work.


## Configuration

In order for sidekick to work with you clusters, you need define the `SIDEKICK_APP_CLOUDPLATFORM` environment variable and edit spark s3 hadoop plugin to point bucket endpoints to sidekick.

These can be configured in two ways:

### Init script configuration

You can add edit any bucket endpoints and regions by adding this section in the init_script

```bash
cat >/databricks/driver/conf/sidekick-spark-conf.conf <<EOL
[driver] {
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET>.endpoint" = "http://localhost:7075"
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET>.endpoint.region" = <AWS_REGION_OF_BUCKET>
}
EOL
```

You can define the environment variables by adding these lines to the [sidekick service init script](./sidekick_service_init.sh):

```bash
export SIDEKICK_APP_CLOUDPLATFORM=<AWS|GCP>
```

### Cluster base configuration

You can also set the bucket endpoints and environment variables by changing you cluster settings.

You can use the [spark configuration](https://docs.databricks.com/clusters/configure.html#spark-configuration) option in your cluster settings to set the bucket endpoint and region:

```
spark.hadoop.fs.s3a.bucket.<MY_BUCKET>.endpoint http://localhost:7075
spark.hadoop.fs.s3a.bucket.<MY_BUCKET>.endpoint.region <AWS_REGION_OF_BUCKET>
```

You can then also use the [env configuration](https://docs.databricks.com/clusters/configure.html#environment-variables) option in your cluster settings to set the environment variables.

```
SIDEKICK_APP_CLOUDPLATFORM=<AWS|GCP>
```
