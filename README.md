## Terragrunt Deployment Example
Example Terragrunt deployment that includes Go test routines using Terratest.

### Decision Records
This repository uses architecture decision records to record design decisions about important elements of the solution.

The ADR index is available [here](./docs/decisions/index.md).

### Requirements
Tested on Go version 1.21 on Ubuntu Linux.

Uses installed packages:
```
awscli
golangci-lint
make
pre-commit
terraform
terragrunt
```

### Configuration
1. Install the packages listed above.
1. Make a copy of the aws/aws.yaml file, named local.aws.yaml, and fill in the fields with configuration values for the target platform.
1. Configure an AWS access key for Terraform provider to use to access the platform:
    ```
    $ aws configure
    ```
1. It's recommended to deploy a build project and folder first, and to use this project in the configuration for additional deployments. The build project can be deployed and managed using this framework with a couple of additional steps to create a local state configuration and then to migrate state to a remote state configuration after the project and bucket are created. ALternatively, the build resources can be pre-created and then imported into the framework using the import command. These considerations are not addressed in this example.

### Deployment
Automated installation configuration, and deployment steps are managed using Makefile targets. Use ```make help``` for a list of configured targets:
```
$ make help
make <target>

Targets:
    help                 Show this help
    pre-commit           Run pre-commit checks

    aws_gitlab_clean     Clean up state files
    aws_gitlab_configure Configure the deployment
    aws_gitlab_deploy    Deploy configured resources
    aws_gitlab_init      Initialize modules, providers
    aws_gitlab_install   Install Terraform, Terragrunt
    aws_gitlab_lint      Run Go linters
    aws_gitlab_plan      Show deployment plan
    aws_gitlab_test      Run deployment tests and clean up (CI loop)
```

Note that configuration targets will be specific to the target; particularly domain names, route53 zones and tags. Update these configuration values before deploying to your own environment by editing the aws_gitlab_configure target in the Makefile.

Additional targets can be added in order to configure multiple environments, for example to create development and production environments.
