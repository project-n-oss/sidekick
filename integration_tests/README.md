# Integration Tests

This directory contains a suite of integration tests for sidekick and bolt. Integration tests can be run againt Sidekick running in AWS or GCP mode.

## Getting Started

1. [AWS Setup Instructions](#aws)
2. [GCP Setup Instructions](#gcp)

<a name="aws"></a>

## AWS Setup Instructions

### Login into aws account

Choose a AWS account to create a test cluster, bucket and run the tests.

:warning: DO NOT use the "root" aws account. Please use a sandbox account or "data-plane". Preferred option is sandbox

### Create a test cluster

Follow the [tutorial](https://docs.google.com/document/d/1SK3gg7th5UbXQpzzhAgvms-Yq-IhHq6JRt8z_2gyoGg/edit#heading=h.bafob67q0tz0) to create a test cluster in your sandbox account

### Create a test ec2 instance

[Create](https://docs.google.com/document/d/1SK3gg7th5UbXQpzzhAgvms-Yq-IhHq6JRt8z_2gyoGg/edit#heading=h.b3g5yr5tcsus) a dedicated test ec2 instance to run the tests.

ssh into the instance and clone the [sidekick repository](https://github.com/project-n-oss/sidekick)

### Create test buckets

You need to create two buckets for the tests:

- A bucket that will be crunched by bolt
- A "failover" bucket that will not be touched by bolt, this will force sidekick to failover back to aws calls.
- A "failover" bucket in a different region, this is useful to make sure aws failover works in different regions.

:warning: Make sure you select the same region as the one your cluster is running in.

1. Go to the s3 console in the same account as your cluster.
2. Create a bucket for the tests. The bucket name should be something like: `sidekick-tests-rvh` where rvh is your initials.
3. Create anoter "failover" bucket for the tests. The bucket name should be something like: `sidekick-failover-tests-rvh` where rvh is your initials.
4. Add the following bucket policies to the buckets:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Statement1",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::{YOUR_ACCOUNT_ID}:role/{PROJECTN_ADMIN_ROLE_OF_CLUSTER}"
      },
      "Action": "s3:*",
      "Resource": ["arn:aws:s3:::{BUCKET_NAME}/*", "arn:aws:s3:::{BUCKET_NAME}"]
    }
  ]
}
```

Make sure to replace the following values:

- `YOUR_ACCOUNT_ID`
- `PROJECTN_ADMIN_ROLE_OF_CLUSTER` (This should be the admin role assumed by the bolt admin server created when you made the cluster)
- `BUCKET_NAME`

### Upload test data

:warning: Make sure you are running this commands on the test ec2 instance created earlier

Run the following command to cp the test data to your new bucket:
:warning: Make sure to be in the `sidekick/integration_tests/` directory before running the command.

```bash
aws s3 cp ./test_data/ s3://{YOUR_BUCKET}/ --recursive
aws s3 cp ./test_data/ s3://{YOUR_FAILOVER_BUCKET}/ --recursive
aws s3 cp ./test_data/ s3://{YOUR_FAILOVER_BUCKET_DIFF_REGION}/ --recursive
```

<a name="gcp"></a>

## GCP Setup Instructions

### Create a test cluster

Follow the GCP [instructions](https://granica.ai/docs/installation-guide#start-pilot) to create a test GCP cluster. The GCP cluster must be private and have read replicas enabled.

### Create a test Google Compute Engine instance

Create a Google Compute Engine instance in the default VPC subnet matching the Crunch region or in a new peered VPC in the same region as Crunch.

### Create an IAM role with full storage permissions

Go to the Google Cloud IAM console and create an IAM role with `storage.*` permissions.

### Create a Service Account and associated key

Go to the Google Cloud IAM console and create a Service Account. Give this service account access to the IAM role created in the previous steps.

Create a JSON key for this service account and download it to your machine.

### Create test buckets

You need to create two buckets for the tests:

- A bucket that will be crunched by bolt
- A "failover" bucket that will not be touched by bolt, this will force sidekick to failover back to aws calls.
- A "failover" bucket in a different region, this is useful to make sure aws failover works in different regions.

:warning: Make sure you select the same region as the one your cluster is running in.

1. Go to the Google Cloud console in the same project as your cluster
2. Create a bucket for the tests. The bucket name should be something like: `sidekick-tests-km` where `km` is your initials.
3. Create anoter "failover" bucket for the tests. The bucket name should be something like: `sidekick-failover-tests-km` where rvh is your initials.

### Crunch Bucket

ssh into your cluster's admin server and crunch your new bucket:

:warning: This step only applies to `<YOUR_BUCKET>`, **NOT** `<YOUR_FAILOVER_BUCKET>`.

```bash
projectn crunch s3://<YOUR_BUCKET>
projectn status
```

Wait for 100% progress on the status board

### Make sure `crunch_traffic_percent` is set to `"100"`

Current integration tests assume that all traffic will first hit Bolt, therefore, ensure that this setting is set.

On the Admin Server run:

```
kubectl describe cm client-behavior-params
```

Make sure `crunch_traffic_percent` is set to `"100"`. Edit the configmap appropriately (`kubectl edit cm client-behavior-params`)

## Running the tests

Create a `.env` file in `sidekick/integration_tests`:

```bash
BUCKET=<YOUR_BUCKET>
FAILOVER_BUCKET=<YOUR_FAILOVER_BUCKET>
FAILOVER_BUCKET_DIFF_REGION=<YOUR_FAILOVER_BUCKET_DIFF_REGION>
GRANICA_CUSTOM_DOMAIN=<YOUR_CLUSTER_DOMAIN>
GRANICA_REGION=<YOUR_CRUNCH_CLUSTER_REGION>
```

You can run test from this directory using the following command:

```bash
go test -bi -v --cloud-platform <aws|gcp>
```

In order to specify a specifc test or series of tests you want to run, you can use the `-run=` arg like so:

```bash
go test -bi -v --cloud-platform <aws|gcp> -run=TestAws/TestGetObject
```
