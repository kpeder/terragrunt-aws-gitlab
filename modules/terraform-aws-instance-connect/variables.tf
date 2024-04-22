variable "preserve_client_ip" {
    description = "Whether to preserve client source addresses during connection forwarding."
    type        = bool

    default = true
}

variable "security_group_ids" {
    description = "List of security group ids to associate with the instance connect endpoint."
    type        = list(string)

    default = null
}

variable "subnet_id" {
    description = "Id of the subnet in which to deploy the instance connect endpoint."
    type        = string
}

variable "tags" {
    description = "Dictionary of tags to apply to the instance connect endpoint."
    type        = map(string)

    default = {}
}
