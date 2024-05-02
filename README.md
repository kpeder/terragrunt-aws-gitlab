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

### GitLab Installation
Installation is decoupled from the instance deployment so that it can be customized, since GitLab installation may contain many environment-specific configuration options, including default root password, external database connection, external object storage location, external authentication, email forwarding configuration, and much more. Follow the instructions below for a vanilla, stand-alone implementation of the product.

1. Connect to the GitLab instance using the configured keypair.

    Update the deployed public key file if necessary by editing the value of pubkey_file value in the following file. It defaults to the ssh-keygen default location.

    ```aws/gitlab/reg-primary/keypairs/gitlab/inputs.yaml```

    Use the associated private key directly or via an SSH agent (recommended), such as ssh-agent (Linux) or Pageant (with PuTTY).

    Use the AWS CLI to connect via the instance connect endpoint.

    ```$ aws ec2-instance-connect ssh --instance-id i-0ea992f180c12345 --os-user ubuntu```

2. Once connected to the instance, follow the [instructions](https://about.gitlab.com/install/#ubuntu) for installing gitlab, generalized below.

    Note that the package repository is already configured on the host. Update the example domain name to match the deployment zone.

    ```$ sudo EXTERNAL_URL="http://gitlab.example.com" apt-get install gitlab-ee```

3. Edit the configuration file to allow health checks from the local VPC range.

    ```$ sudo vi /etc/gitlab/gitlab.rb```

    Uncomment and edit the following line to add the VPC range (the default VPC range for this deployment is 172.16.0.0/16).

    ```gitlab_rails['monitoring_whitelist] = ['127.0.0.1/8', '172.16.0.0/16']```

    Then reconfigure GitLab.

    ```$ sudo gitlab-ctl reconfigure```

### Connection
1. Copy the installation password from the GitLab instance.

    ```$ sudo cat /etc/gitlab/initial_root_password```

2. Connect to the endpoint using a browser and log in as user 'root' with the retrieved password. Update the example domain name to match the deployment zone.

    ```https://gitlab.example.com/```
