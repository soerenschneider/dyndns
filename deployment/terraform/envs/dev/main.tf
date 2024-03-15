locals {
  hosted_zone = "ABCDEF12345678"
  env_vars = {
    "DYNDNS_HOSTED_ZONE_ID" = local.hosted_zone,
    "DYNDNS_KNOWN_HOSTS" = replace(replace(<<EOT
      {
        "your.record.tld": [
          "IyXH8z/+vRsIUEAldlGgKKFcVHoll8w2tzC6o9717m8="
        ]
      }
EOT
    , "\n", ""), " ", "")
  }
}

module "dyndns-server" {
  source      = "../../lambda-deployment-module"
  env_vars    = local.env_vars
  hosted_zone = local.hosted_zone
}

output "gateway-url" {
  value = module.dyndns-server.api_gateway_invoke_url
}