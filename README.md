![release](https://img.shields.io/github/v/release/project-n-oss/sidekick)

![projectn-sidekick.png](projectn-sidekick.png)

# Sidekick

Sidekick is a [sidecar](https://learn.microsoft.com/en-us/azure/architecture/patterns/sidecar) proxy process that allows your applications to talk with a Bolt cluster through any AWS SDK.

## Getting started

- go1.20
- A bolt cluster
- A cloud instance that has a vpc connection with the bolt cluster

## Running Sidekick

### Env Variables

In order to run Sidekick, you first need to set some ENV variables

```bash
export GRANICA_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN>
# Optional if not running on a ec2 instance or running in a different region
export AWS_REGION=<YOUR_BOLT_CLUSTER_REGION>
# Optional if not running on a ec2 instance to force read from a read-replica in this az
export AWS_ZONE_ID=<AWS_ZONE_ID>
```

### Traffic Splitting

Traffic splitting provides a mechanism to precisely control how traffic is distributed between Bolt (Crunch) and S3, enabling gradual rollouts and offering numerous benefits. This capability allows you to onboard your applications in a safe and controlled manner, minimizing risks and ensuring smooth transitions.

Traffic splitting configuration is managed through the `client-behavior-params` ConfigMap in the Bolt (Crunch) Kubernetes cluster. This ConfigMap can be edited on your behalf by the Granica team or by you via `custom.vars` or directly editing the ConfigMap (the latter option is not recommended as it can cause state drift.) For further guidance on the traffic splitting configuration, reach out to the Granica team.

### Fallback on 404
If sidekick can't find your object in Bolt, sidekick tries to find the object in S3. This happens transparently to the client, and it doesn't need any client retries.

### Failover
Sidekick has a failover mechanism that comes into play when there are network failures or when Bolt returns 500s
In these situations sidekick returns a 500 to the client, that may cause the client to retry. The retry request will make sidekick request Bolt again but a different endpoint. Eventually, either client will exhaust all retries to sidekick or sidekick will exhaust all replica endpoints. When one of these happens, the request has failed
The above failover mechanism is recommended, but can be changed. Sidekick can instead failover to S3 transparently to the Client immediately on noticing a network failure or 500 error. This can be set with either by a command line argument or environment variable

```bash
# Using flag
go run main serve --failover=false
# Using binary
./sidekick serve --failover=false
```

```bash
# Using env variable
export SIDEKICK_BOLTROUTER_FAILOVER=true  # Not recommended
go run main serve
```

### Local

You can run sidekick directly from the command line:

```bash
go run main.go serve
```

This will run sidekick localy on your machine on `localhost:7075`.

run the following command to learn more about the options:

```bash
go run main.go serve --help
```

### Logging

Sidekick supports a `--log-level` argument to control the logging level. By default, the logging level is set to `info`. However, you can set a more verbose log level, such as debug, to enable detailed debugging information. For all available logging options run `./sidekick --help`.

### Docker

Build the docker image:

```bash
docker build -t sidekick .
```

or pull one from the [containers page](https://github.com/project-n-oss/sidekick/pkgs/container/sidekick)

#### Running on an EC2 Instance using instance profile credentials

```bash
docker run -p 7075:7075 --env GRANICA_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN> -env AWS_REGION=<YOUR_BOLT_CLUSTER_REGION> <sidekick-image> sidekick serve
```

#### Running on any machine using environment variable credentials

```bash
docker run -p 7075:7075 --env GRANICA_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN> -env AWS_REGION=<YOUR_BOLT_CLUSTER_REGION> --env AWS_ACCESS_KEY_ID=<YOUR_AWS_ACCESS_KEY> --env AWS_SECRET_ACCESS_KEY="<YOUR_AWS_SECRET_KEY>" <sidekick-image> serve -v
```

If using temporary credentials, add `--env AWS_SESSION_TOKEN=<YOUR_SESSION_TOKEN>` to the command above. However, this is not recommended since credentials will expire. Instead, consider using the credentials profiles file with role assumption directives.

#### Running on any machine using the credential profiles file

```bash
docker run -p 7075:7075 --env GRANICA_CUSTOM_DOMAIN=<YOUR_CUSTOM_DOMAIN> --env AWS_REGION=<YOUR_BOLT_CLUSTER_REGION> -v ~/.aws/:/root/.aws/ <sidekick-image> serve
```

By default, the `default` profile from the credentials file will be used. If you want to use another profile from the credentials file add `--env AWS_DEFAULT_PROFILE=<YOUR_PROFILE>` to the command above.

## Using Sidekick

### AWS sdks

You can find examples on how to setup your aws sdk clients to work with sidekick [here](./integrations/AWS_SDK.md)

### 3rd Party Integrations

You can find more information to integrated sidekick with 3 party tools/frameworks/services [here](./integrations)

### Pre Built binaries

Sidekick binaries are hosted and released from GitHub. Please check our [releases page](./releases).
To download any release of our linux amd64 binary run:

```bash
wget https://github.com/project-n-oss/sidekick/releases/${release}/download/sidekick-linux-amd64.tar.gz
```

## Contributing

### Versioning

This repository uses [release-please](https://github.com/google-github-actions/release-please-action) to create and manage release.

### Commits

We follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) for our commits and PR titles. This allows us to use release-please to manage our releases.

The most important prefixes you should have in mind are:

- fix: which represents bug fixes, and correlates to a SemVer patch.
- feat: which represents a new feature, and correlates to a SemVer minor.
- feat!:, or fix!:, refactor!:, etc., which represent a breaking change (indicated by the !) and will result in a SemVer major.
