# S3 Browser integration

This guide will walk you through how to seamlessly integrate SideKick with your [S3 Browser](https://s3browser.com/) instance running on your machine.

## SideKick configuration

To get started, you will need to run SideKick on your machine. Head over to the [releases page](https://github.com/project-n-oss/sidekick/releases) and download the package for your OS and architecture.

Next, extract the binary from the downloaded file:

**Windows**
Right-click on the downloaded file in the File Explorer and click "Extract All".

Before running the executable, make sure the following environment variables are set:

- `GRANICA_CUSTOM_DOMAIN`: This is the custom domain you chose during setup.
- `AWS_REGION`: This is the region of your Bolt deployment.

**Windows**

```bash
set GRANICA_CUSTOM_DOMAIN=your-bolt-domain.com
set AWS_REGION=your-bolt-region
```

## S3 Browser configuration

1. Add a new account with "S3 Compatible Storage" type
2. REST Endpoint: `localhost:7075`
3. Enter AWS Access Key, AWS Secret Key
4. Make sure "Use secure transfer" is **not checked**
5. Click on “Advanced Settings”
6. Under "Signature Version” select "Signature V4"
7. Configure all regions where your buckets live like below, save changes and refresh the bucket list.

```
DefaultRegion=bolt-region
USE1=us-east-1
USW1=us-west-1
```
