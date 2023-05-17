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

Now that you have SideKick running, start CyberDuck.

First, you'll need to enable the "S3 (HTTP)" profile.

1. Navigate to Settings
   Windows: Click "Edit" and then "Preferences"
   macOS: Click "CyberDuck" in the top left corner and then "Settings"

2. Then navigate to "Profiles" and search for "s3 http"

3. Tick "S3 (HTTP)"

Now download the SideKick CyberDuck bookmark file, drag it onto your CyberDuck Bookmarks view and enter your AWS credentials.

Available CyberDuck profiles

- Single-bucket profile
  Configured to access a single crunched bucket via sidekick. Before dragging the bookmark file onto the CyberDuck window, open the file in a text editor and replace `bucket-name` with the name of your bucket.

[profile](./sidekick-single-bucket.duck)

- Multi-bucket profile - configure to access all accessible buckets (crunched and not) - tbd
