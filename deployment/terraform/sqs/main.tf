module "clients" {
  source      = "./clients"
  clients     = var.clients
  sqs_arn     = aws_sqs_queue.dyndns.arn
  environment = var.environment
}

module "servers" {
  source      = "./servers"
  servers     = var.servers
  hosted_zone = var.hosted_zone
  sqs_arn     = aws_sqs_queue.dyndns.arn
  environment = var.environment
}