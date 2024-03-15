resource "aws_sqs_queue" "dyndns" {
  name                      = var.queue_name
  delay_seconds             = 0
  message_retention_seconds = 86400
}