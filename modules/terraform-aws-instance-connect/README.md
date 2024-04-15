# Terraform AWS Instance Connect

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 5.6 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | >= 5.6 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_ec2_instance_connect_endpoint.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ec2_instance_connect_endpoint) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_subnet_id"></a> [subnet\_id](#input\_subnet\_id) | Id of the subnet in which to deploy the instance connect endpoint. | `string` | n/a | yes |
| <a name="input_tags"></a> [tags](#input\_tags) | Dictionary of tags to apply to the instance connect endpoint. | `map(string)` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_dns_name"></a> [dns\_name](#output\_dns\_name) | DNS name of the instance connect endpoint. |
| <a name="output_subnet_id"></a> [subnet\_id](#output\_subnet\_id) | Id of the subnet in which the instance connect endpoint is deployed. |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | Id of the VPC in which the instance connect endpoint is deployed. |
<!-- END_TF_DOCS -->
