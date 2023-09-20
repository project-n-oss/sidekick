# Kubernetes integration

This guide will walk you through how to seamlessly integrate SideKick with your Kubernetes applications by adding a SideKick sidecar container to your application pod.

## SideKick sidecar container configuration

To get started, you will need to configure your SideKick sidecar container. A sample Kubernetes manifest file is provided [here](./sidekick_sidecar.yaml), which you can use as a reference. When configuring the SideKick sidecar container, we recommend that you specify a specific container image tag for production deployments, rather than using the default "latest" tag. Available tags can be found on our container package [page](https://github.com/project-n-oss/sidekick/pkgs/container/sidekick).

Additionally, the following environment variables must be set:

- `GRANICA_CUSTOM_DOMAIN`: This is the custom domain you chose during setup.
- `GRANICA_REGION`: This is the region of your Bolt deployment.

## Application configuration

Now that you have configured your SideKick sidecar container, you need to point the S3 clients in your application to `http://localhost:7075` as the S3 endpoint. To achieve this, we suggest editing your application code to read the S3 endpoint URL from an environment variable and then setting that environment variable to `http://localhost:7075` in the application container.

## (Optional) VPC Peering

If you have configured Bolt in the private hosted zone, you will need to configure VPC peering between your Kubernetes cluster VPC and the Bolt VPC to enable communication between the two. VPC peering provides a secure and direct connection between the two VPCs, allowing you to seamlessly integrate Bolt with your Kubernetes applications.

To configure VPC peering, follow the tutorial provided [here](https://xyz.projectn.co/vpc-peering). The tutorial provides step-by-step instructions on how to set up VPC peering.
