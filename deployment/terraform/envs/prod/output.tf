
output "gateway-url" {
  value = module.dyndns.api_gateway_invoke_url
}

output "ids_server" {
  value     = module.sqs.ids_server
  sensitive = true
}

output "ids_client" {
  value     = module.sqs.ids_client
  sensitive = true
}

output "sqs_queue" {
  value = module.sqs.sqs_url
}