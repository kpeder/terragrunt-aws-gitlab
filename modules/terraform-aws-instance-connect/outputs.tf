output "availability_zone" {
  description = "The availability zone in which the instance connect endpoint is deployed."
  value       = aws_ec2_instance_connect_endpoint.this.availability_zone
}

output "dns_name" {
  description = "DNS name of the instance connect endpoint."
  value       = aws_ec2_instance_connect_endpoint.this.dns_name
}

output "preserve_client_ip" {
  description = "Whether client source addresses will be preserved during connection forwarding."
  value       = aws_ec2_instance_connect_endpoint.this.preserve_client_ip
}

output "security_group_ids" {
  description = "List of security group ids associated with the instance connect endpoint."
  value       = aws_ec2_instance_connect_endpoint.this.security_group_ids
}

output "subnet_id" {
  description = "Id of the subnet in which the instance connect endpoint is deployed."
  value       = aws_ec2_instance_connect_endpoint.this.subnet_id
}

output "tags" {
  description = "Dictionary of tags applied to the instance connect endpoint."
  value       = aws_ec2_instance_connect_endpoint.this.tags
}

output "vpc_id" {
  description = "Id of the VPC in which the instance connect endpoint is deployed."
  value       = aws_ec2_instance_connect_endpoint.this.vpc_id
}
