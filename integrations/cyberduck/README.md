# CyberDuck integration

This guide will walk you through how to seamlessly integrate SideKick with your CyberDuck instance running on your machine.

## SideKick configuration

To get started, you will need to run SideKick on your machine. Head over to the [releases page](https://github.com/project-n-oss/sidekick/releases) and download the package for your OS and architecture.

Next, extract the binary from the downloaded file:

**macOS**

```bash
tar -xvf <artifact>.tar.gz
```

**Windows**
Right-click on the downloaded file in the File Explorer and click "Extract All".

Before running the executable, make sure the following environment variables are set:

- `BOLT_CUSTOM_DOMAIN`: This is the custom domain you chose during setup.
- `AWS_REGION`: This is the region of your Bolt deployment.

**macOS**

```bash
export BOLT_CUSTOM_DOMAIN="your-custom-domain.com"
export AWS_REGION="your-bolt-region"
```

**Windows**

```bash
set BOLT_CUSTOM_DOMAIN=your-bolt-domain.com
set AWS_REGION=your-bolt-region
```

## CyberDuck configuration

Now that you have SideKick running, download the SideKick CyberDuck bookmark file, drag it onto your CyberDuck Bookmarks view and enter your AWS credentials.

Available CyberDuck profiles

- Single-bucket profile - configured to access a single crunched bucket via sidekick - [profile](./sidekick-single-bucket.duck)
- Multi-bucket profile - configure to access all accessible buckets (crunched and not) - tbd
