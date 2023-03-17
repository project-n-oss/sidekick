# Integration Tests

This directory contains a suite of integration tests for sidekick and bolt.

## Getting Started

### Login into aws account

Choose a aws account to create a test cluster, bucket and run the tests.

:warning: DO NOT use the "root" aws account. Please use a sandbox account or "data-plane". Preferred option is sandbox

### Create a test cluster

Follow the [tutorial](https://docs.google.com/document/d/1SK3gg7th5UbXQpzzhAgvms-Yq-IhHq6JRt8z_2gyoGg/edit#heading=h.bafob67q0tz0) to create a test cluster in your sandbox account

### Create a test ec2 instance

[Create](https://docs.google.com/document/d/1SK3gg7th5UbXQpzzhAgvms-Yq-IhHq6JRt8z_2gyoGg/edit#heading=h.b3g5yr5tcsus) a dedicated test ec2 instance to run the tests.

ssh into the instance and clone the [sidekick repository](https://github.com/project-n-oss/sidekick)

### Create a test bucket

:warning: Make sure you select the same region as the one your cluster is running in.

1. Go to the s3 console in the same account as your cluster.
2. Create a bucket for the tests. The bucket name should be something like: `sidekick-tests-rvh` where rvh is your initials.
3. Add the following bucket policies to the bucket:

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
            "Resource": [
                "arn:aws:s3:::{BUCKET_NAME}/*",
                "arn:aws:s3:::{BUCKET_NAME}"
            ]
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
```

### Crunch Bucket

ssh into your cluster's admin server and crunch your new bucket:

```bash
projectn crunch s3://YOUR_BUCKET
projectn status
```

Wait for 100% progress on the status board

## Running the tests

Create a `.env` file in `sidekick/integration_tests`:

```bash
BUCKET=YOU_BUCKET
BOLT_CUSTOM_DOMAIN=YOUR_CLUSTER_DOMAIN
AWS_REGION=REGION_OF_BUCKET_AND_CLUSTER
```

You can run test from this directory using the following command:

```bash
go test -i -v 
```

In order to specify a specifc test or series of tests you want to run, you can use the `-run=` arg like so:

```bash
go test -i -v -run=TestAws/TestAws/TestGetObject
```
