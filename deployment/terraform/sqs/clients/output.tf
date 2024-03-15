output "iam_keys" {
  value     = aws_iam_access_key.dyndns_client[*]
  sensitive = true
}