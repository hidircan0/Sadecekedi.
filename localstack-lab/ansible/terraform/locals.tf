locals {
  # Single prefix so names stay consistent across S3, SQS, SNS, DynamoDB, IAM.
  name_prefix = "${var.project}-${var.environment}"

  tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}
