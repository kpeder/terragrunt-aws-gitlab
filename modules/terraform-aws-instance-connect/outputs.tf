output "dns_name" {
  description = "DNS name of the instance connect endpoint."
  value       = aws_ec2_instance_connect_endpoint.this.dns_name
}

output "subnet_id" {
  description = "Id of the subnet in which the instance connect endpoint is deployed."
  value       = aws_ec2_instance_connect_endpoint.this.subnet_id
}

output "vpc_id" {
  description = "Id of the VPC in which the instance connect endpoint is deployed."
  value       = aws_ec2_instance_connect_endpoint.this.vpc_id
}
