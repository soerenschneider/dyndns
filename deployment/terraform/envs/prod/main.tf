data "sops_file" "vars" {
  source_file = "vars.sops.json"
}

module "dyndns" {
  source      = "../../lambda-deployment-module"
  env_vars    = jsondecode(data.sops_file.vars.raw).env_vars
  hosted_zone = data.sops_file.vars.data["hosted_zone"]
}

module "sqs" {
  source      = "../../sqs"
  hosted_zone = data.sops_file.vars.data["hosted_zone"]
  clients     = [
    "router-dd",
    "k8s-dd",
    "router-ez",
    "k8s-ez"
  ]
  servers     = ["k8s-dd"]
  environment = "prod"
  queue_name  = "dyndns-prod"
}
