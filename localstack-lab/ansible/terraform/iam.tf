# Fictional worker identity. No Lambda is created; the goal is to read an IAM graph.
data "aws_iam_policy_document" "assume_lambda" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "worker" {
  name               = "${local.name_prefix}-worker"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda.json
  tags               = local.tags
}

data "aws_iam_policy_document" "worker" {
  statement {
    sid = "BucketObjects"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = [
      "${module.demo_bucket.arn}/*",
    ]
  }

  statement {
    sid = "QueueConsume"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
    ]
    resources = [
      aws_sqs_queue.demo.arn,
    ]
  }
}

resource "aws_iam_role_policy" "worker" {
  name   = "${local.name_prefix}-worker"
  role   = aws_iam_role.worker.id
  policy = data.aws_iam_policy_document.worker.json
}
