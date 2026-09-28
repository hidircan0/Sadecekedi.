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

output "vpc_id" {
  value = aws_vpc.lab_vpc.id
}

output "subnet_id" {
  value = aws_subnet.lab_subnet.id
}

output "security_group_id" {
  value = aws_security_group.lab_sg.id
}

output "ec2_instance_id" {
  value = aws_instance.web_server.id
}

output "ec2_public_ip" {
  value = aws_instance.web_server.public_ip
}

output "ec2_private_ip" {
  value = aws_instance.web_server.private_ip
}

output "s3_object_key" {
  value = aws_s3_object.sample_file.key
}
