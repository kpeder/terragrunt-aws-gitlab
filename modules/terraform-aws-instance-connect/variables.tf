variable "subnet_id" {
    description = "Id of the subnet in which to deploy the instance connect endpoint."
    type        = string
}

variable "tags" {
    description = "Dictionary of tags to apply to the instance connect endpoint."
    type        = map(string)
}
