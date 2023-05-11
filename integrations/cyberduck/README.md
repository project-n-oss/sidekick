# CyberDuck integration

This guide will walk you through how to seamlessly integrate SideKick with your CyberDuck instance running on your machine.

## SideKick configuration

To get started, you will need to run SideKick on your machine. Head over to the [releases page](https://github.com/project-n-oss/sidekick/releases) and download the package for your OS and architecture.

Next, extract the binary from the downloaded file:

#### Linux / macOS
```bash
tar -xvf <artifact>.tar.gz
```

#### Windows
Right-click on the downloaded file in the File Explorer and click "Extract All".

Before running the executable, make sure the following environment variables are set:

* `BOLT_CUSTOM_DOMAIN`: This is the custom domain you chose during setup.
* `AWS_REGION`: This is the region of your Bolt deployment.

#### Linux / macOS

```bash
export BOLT_CUSTOM_DOMAIN="your-custom-domain.com"
export AWS_REGION="your-bolt-region"
```

#### Windows
```bash
set BOLT_CUSTOM_DOMAIN=your-bolt-domain.com
set AWS_REGION=your-bolt-region
```

## CyberDuck configuration

Now that you have SideKick running, download the SideKick CyberDuck bookmark file from [here](./sidekick.duck). Drag it onto your CyberDuck window and enter your AWS credentials.
