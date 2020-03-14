# Amazon ECS for Open Application Model

The oam-ecs CLI is a proof-of-concept that partially implements the [Open Application Model](https://oam.dev/) (OAM) specification, version v1alpha1.

The oam-ecs CLI provisions two of the core OAM workload types as Amazon ECS services running on AWS Fargate using AWS CloudFormation.  A workload of type `core.oam.dev/v1alpha1.Worker` will deploy a CloudFormation stack containing an ECS service running in private VPC subnets with no accessible endpoint.  A workload of type `core.oam.dev/v1alpha1.Server` will deploy a CloudFormation stack containing an ECS service running in private VPC subnets, behind a publicly-accessible network load balancer.

For a full comparison with the OAM specification, see the [Compatibility](COMPATIBILITY.md) page.

>⚠️ Note that this project is a proof-of-concept and should not be used with production workloads.

## Build & test

```
make
make test
export PATH="$PATH:./bin/local"
oam-ecs --help
```

## Deploy an oam-ecs environment

The oam-ecs environment deployment creates a VPC with public and private subnets where OAM workloads can be deployed.

```
oam-ecs env deploy
```

The environment attributes like VPC ID and ECS cluster name can be described.

```
oam-ecs env show
```

The CloudFormation template deployed by this command can be [seen here](templates/environment/cf.yml).

## Deploy OAM workloads with oam-ecs

The dry-run step outputs the CloudFormation template that represents the given OAM workloads.  The CloudFormation templates are written to the `./oam-ecs-dry-run-results` directory.

```
oam-ecs app deploy --dry-run \
  -f examples/example-app.yaml \
  -f examples/worker-component.yaml \
  -f examples/server-component.yaml
```

Then the CloudFormation resources, including load balancers and ECS services running on Fargate, can be deployed:

```
oam-ecs app deploy \
  -f examples/example-app.yaml \
  -f examples/worker-component.yaml \
  -f examples/server-component.yaml
```

The application component instances' attributes like ECS service name and endpoint DNS name can be described.

```
oam-ecs app show -f examples/example-app.yaml
```

## Tear down

To delete all infrastructure provisioned by oam-ecs, first delete the deployed applications:

```
oam-ecs app delete -f examples/example-app.yaml
```

Then delete the environment infrastructure:

```
oam-ecs env delete
```

## Credentials and Region

oam-ecs will look for credentials in the following order, using the [default provider chain](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) in the AWS SDK for Go.

1. Environment variables.
1. Shared credentials file. Profiles can be specified using the `AWS_PROFILE` environment variable.
1. If running on Amazon ECS (with task role) or AWS CodeBuild, IAM role from the container credentials endpoint.
1. If running on an Amazon EC2 instance, IAM role for Amazon EC2.

No credentials are required for dry-runs of the oam-ecs tool.

oam-ecs will determine the region in the following order, using the [default behavior](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-the-region) in the AWS SDK for Go.

1. From the `AWS_REGION` environment variable.
1. From the `config` file in the `.aws/` folder in your home directory.

## Security Disclosures

If you would like to report a potential security issue in this project, please do not create a GitHub issue.  Instead, please follow the instructions [here](https://aws.amazon.com/security/vulnerability-reporting/) or [email AWS Security directly](mailto:aws-security@amazon.com).

## License

This project is licensed under the Apache-2.0 License.
