variable "hosted_zone" {
  type = string
  description = "AWS Route53 hosted zone that holds the DNS records"
}

variable "env_vars" {
  type = map
  description = "Environment variables for Dyndns server"
}