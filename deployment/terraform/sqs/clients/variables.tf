variable "environment" {
  type = string
  validation {
    condition     = can(regex("^(dev|prod|dqs)$", var.environment))
    error_message = "This variable must be either 'dev', 'prod', or 'dqs'."
  }
}

variable "sqs_arn" {
  type = string
}

variable "clients" {
  type = set(string)
}
