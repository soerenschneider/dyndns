output "api_gateway_invoke_url" {
  value = aws_api_gateway_deployment.dyndns-server_v1.invoke_url
}
