locals {
  archive_file = "${path.module}/../../../dyndns-server-lambda.zip"
}

resource "aws_lambda_function" "dyndns-server" {
  architectures    = ["arm64"]
  function_name    = "dyndns-server"
  filename         = local.archive_file
  source_code_hash = filebase64sha256(local.archive_file)
  handler          = "bootstrap"
  role             = aws_iam_role.dyndns-server.arn
  runtime          = "provided.al2"
  memory_size      = 128
  timeout          = 5

  environment {
    variables = var.env_vars
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
  ]
}

resource "aws_iam_role" "dyndns-server" {
  name               = "lambda-dyndns-server"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }
}
POLICY
}

resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/lambda/${aws_lambda_function.dyndns-server.function_name}"
  retention_in_days = 3
}

# See also the following AWS managed policy: AWSLambdaBasicExecutionRole
data "aws_iam_policy_document" "lambda_logging" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_policy" "route53" {
  name        = "lambda-dyndns-server-dns"
  description = "IAM policy for logging from a lambda"
  policy      = data.aws_iam_policy_document.lambda-route53.json
}

data "aws_iam_policy_document" "lambda-route53" {
  statement {
    effect = "Allow"
    actions = [
      "route53:ChangeResourceRecordSets"
    ]
    resources = [
      "arn:aws:route53:::hostedzone/${var.hosted_zone}"
    ]
  }
}

resource "aws_iam_role_policy_attachment" "lambda_route53" {
  role       = aws_iam_role.dyndns-server.name
  policy_arn = aws_iam_policy.route53.arn
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "lambda_logging_dyndns_server"
  path        = "/"
  description = "IAM policy for logging from a lambda"
  policy      = data.aws_iam_policy_document.lambda_logging.json
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.dyndns-server.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}
