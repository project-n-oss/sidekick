# Spark integration

You can integrate sidekick with Apache Spark by adding the [sidekick_service_init.sh](./sidekick_service_init.sh) as an [init script]() in your Spark clusters. This init-script should be configured to run on all Spark nodes.

Briefly, the init script does the following:
    - Install sidekick on the Spark node
    - Configure S3 endpoint (for specific buckets) to point to sidekick
    - Setup a sytstemctl service to run sidekick as a daemon

## Configuration

To get started, download the sample [init script]() and make the following changes.

1. 
Add bucket endpoints and regions which will be accessed via sidekick by adding to this section in the init_script.

```bash
cat >/databricks/driver/conf/sidekick-spark-conf.conf <<EOL
[driver] {
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET1>.endpoint" = "http://localhost:7075"
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET1>.endpoint.region" = <AWS_REGION_OF_BUCKET1>
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET2>.endpoint" = "http://localhost:7075"
  "spark.hadoop.fs.s3a.bucket.<MY_BUCKET2>.endpoint.region" = <AWS_REGION_OF_BUCKET2>
}
EOL
```

2.
Define the environment variables by adding these lines to the [sidekick service init script](./sidekick_service_init.sh):

```bash
export SIDEKICK_APP_CLOUDPLATFORM=<AWS|GCP>
```
