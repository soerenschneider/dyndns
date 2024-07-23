variable "queue_name" {
  type    = string
  default = "dyndns"
}

variable "environment" {
  type = string
  validation {
    condition     = can(regex("^(dev|prod|dqs)$", var.environment))
    error_message = "This variable must be either 'dev', 'prod', or 'dqs'."
  }
}

variable "clients" {
  type = set(string)
}

variable "servers" {
  type = set(string)
}

variable "hosted_zone" {
  type = string
}
