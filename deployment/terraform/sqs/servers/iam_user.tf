locals {
  prefix = "dyndns-server"
}

resource "aws_iam_user" "dyndns_server" {
  for_each = var.servers
  name     = "${local.prefix}-${var.environment}-${each.value}"
}

resource "aws_iam_user_policy_attachment" "dyndns_server" {
  for_each   = var.servers
  user       = aws_iam_user.dyndns_server[each.value].name
  policy_arn = aws_iam_policy.dyndns_server.arn
}

resource "aws_iam_access_key" "dyndns_server" {
  for_each = var.servers
  user     = aws_iam_user.dyndns_server[each.value].name
}
