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
  validation {
    condition     = length(var.clients) > 0 && length([for s in var.clients : s if can(regex("^[a-z]{3,}$", s))]) == length(var.clients)
    error_message = "clients must not be empty, each client name must be lowercase and contain at least three characters."
  }
}

variable "servers" {
  type = set(string)
  validation {
    condition     = length(var.servers) > 0 && length([for s in var.servers : s if can(regex("^[a-z]{3,}$", s))]) == length(var.servers)
    error_message = "servers must not be empty, each server name must be lowercase and contain at least three characters."
  }
}

variable "hosted_zone" {
  type = string
}