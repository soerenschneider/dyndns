data "sops_file" "vars" {
  source_file = "vars.sops.json"
}

module "dyndns" {
  source = "../../lambda-deployment-module"
  env_vars = jsondecode(data.sops_file.vars.raw).env_vars
  hosted_zone = data.sops_file.vars.data["hosted_zone"]
}

output "gateway-url" {
  value = module.dyndns.api_gateway_invoke_url
}
