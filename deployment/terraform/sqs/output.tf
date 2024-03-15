output "sqs_url" {
  value = aws_sqs_queue.dyndns.url
}

output "sqs_arn" {
  value = aws_sqs_queue.dyndns.arn
}

output "ids_server" {
  value     = module.servers.iam_keys
  sensitive = true
}

output "ids_client" {
  value     = module.clients.iam_keys
  sensitive = true
}