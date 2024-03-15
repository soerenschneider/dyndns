locals {
  prefix = "dyndns-client"
}

resource "aws_iam_user" "dyndns_client" {
  for_each = var.clients
  name     = "${local.prefix}-${var.environment}-${each.value}"
}

resource "aws_iam_user_policy_attachment" "dyndns_client" {
  for_each   = var.clients
  user       = aws_iam_user.dyndns_client[each.value].name
  policy_arn = aws_iam_policy.dyndns_client_policy.arn
}

resource "aws_iam_access_key" "dyndns_client" {
  for_each = var.clients
  user     = aws_iam_user.dyndns_client[each.value].name
}