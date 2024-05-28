![release](https://img.shields.io/github/v/release/project-n-oss/sidekick)

![projectn-sidekick.png](granica-sidekick.png)

# Sidekick

Sidekick is a [sidecar](https://learn.microsoft.com/en-us/azure/architecture/patterns/sidecar) proxy process that helps you integrate with the granica crunch platform.

## What it does

Sidekick runs as a sidecar next to you application code and acts as a proxy to S3. If sidecar finds a crunched version of the file you are trying to query it will always return a 409. This garantees an error on the client side during the crunching of a file.

## Getting started

### Perequisites

- go1.21

## Running Sidekick

### Local

You will need to create a config.yml file in the root of the project. You can use the following template:

```yaml
App:
  CloudPlatform: AWS
```

These config values can also be set from ENV variable like so:

```bash
export SIDEKICK_APP_CLOUDPLATFORM=AWS
```

You can then run sidekick directly from the command line:

```bash
go run main.go serve
```

This will run sidekick localy on your machine on `localhost:7075`.

run the following command to learn more about the options:

```bash
go run main.go serve --help
```

## Using Sidekick

### Docker

You can pull the docker image from the [containers page](https://github.com/project-n-oss/sidekick/pkgs/container/sidekick)

You can then run the docker image with the following command:

```bash
docker run -p 7075:7075 --env SIDEKICK_APP_CLOUDPLATFORM=AWS <sidekick-image> serve 
```

### Pre Built binaries

Sidekick binaries are hosted and released from GitHub. Please check our [releases page](https://github.com/project-n-oss/sidekick/releases).
To download any release of our linux amd64 binary run:

```bash
wget https://github.com/project-n-oss/sidekick/releases/download/${release}/sidekick-linux-amd64.tar.gz
```

You can then run the binary directly:

```bash
SIDEKICK_APP_CLOUDPLATFORM=AWS ./sidekick serve
```

### Integrations

Document on how to integrate sidekick with various services can be found in the [integrations](./integrations) folder.

## Contributing

### Versioning

This repository uses [release-please](https://github.com/google-github-actions/release-please-action) to create and manage release.

### Commits

We follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for our commits and PR titles. This allows us to use release-please to manage our releases.

The most important prefixes you should have in mind are:

- fix: which represents bug fixes, and correlates to a SemVer patch.
- feat: which represents a new feature, and correlates to a SemVer minor.
- feat!:, or fix!:, refactor!:, etc., which represent a breaking change (indicated by the !) and will result in a SemVer major.
