variable "aws_region" {
  description = "The AWS region to deploy to"
  default     = "eu-west-1"
}

variable "availability_zone" {
  description = "Availability zone of resources"
  type        = string
}
 
variable "instance_ami" {
  description = "ID of the AMI used"
  type        = string
}
 
variable "instance_type" {
  description = "Type of the instance"
  type        = string
}
 
variable "ssh_public_key" {
  description = "Public SSH key for logging into EC2 instance"
  type        = string
}

# TODO: to be completed. check aws specific documentation
