resource "aws_iam_policy" "dyndns_server" {
  name        = "dyndns-server-${var.environment}"
  description = "Allows receiving and deleting messages from the SQS queue as well as updating hosting zone records"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage"
      ],
      "Resource": "${var.sqs_arn}"
    },

    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets"
      ],
      "Resource": "arn:aws:route53:::hostedzone/${var.hosted_zone}"
    }
  ]
}
EOF
}
