output "bucket_name" {
  value = module.demo_bucket.bucket
}

output "bucket_arn" {
  value = module.demo_bucket.arn
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.items.name
}

output "dynamodb_table_arn" {
  value = aws_dynamodb_table.items.arn
}

output "queue_name" {
  value = aws_sqs_queue.demo.name
}

output "queue_arn" {
  value = aws_sqs_queue.demo.arn
}

output "topic_name" {
  value = aws_sns_topic.demo.name
}

output "topic_arn" {
  value = aws_sns_topic.demo.arn
}

output "worker_role_name" {
  value = aws_iam_role.worker.name
}

output "worker_role_arn" {
  value = aws_iam_role.worker.arn
}

output "backend_url" {
  value = var.backend_url
}

output "validator_url" {
  value = var.validator_url
}

output "storage_endpoint" {
  value = var.storage_endpoint
}

output "storage_bucket" {
  value = var.storage_bucket
}
