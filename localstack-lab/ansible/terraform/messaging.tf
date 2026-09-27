resource "aws_sqs_queue" "demo" {
  name = "${local.name_prefix}-queue"
  tags = local.tags
}

resource "aws_sns_topic" "demo" {
  name = "${local.name_prefix}-events"
  tags = local.tags
}

# Queue policy is required so SNS is allowed to push; subscription alone is not enough.
data "aws_iam_policy_document" "sqs_from_sns" {
  statement {
    sid     = "AllowSnsSend"
    actions = ["sqs:SendMessage"]
    resources = [
      aws_sqs_queue.demo.arn,
    ]

    principals {
      type        = "Service"
      identifiers = ["sns.amazonaws.com"]
    }

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"
      values   = [aws_sns_topic.demo.arn]
    }
  }
}

resource "aws_sqs_queue_policy" "demo" {
  queue_url = aws_sqs_queue.demo.id
  policy    = data.aws_iam_policy_document.sqs_from_sns.json
}

resource "aws_sns_topic_subscription" "queue" {
  topic_arn = aws_sns_topic.demo.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.demo.arn
}
