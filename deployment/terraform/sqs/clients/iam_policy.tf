resource "aws_iam_policy" "dyndns_client_policy" {
  name        = "dyndns-client-${var.environment}"
  description = "Allows pushing messages to the SQS queue"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:SendMessage"
      ],
      "Resource": "${var.sqs_arn}"
    }
  ]
}
EOF
}
