variable "project" {
  type        = string
  description = "Name prefix for all lab resources."
}

variable "environment" {
  type        = string
  description = "Logical environment name (local only; no real AWS accounts)."
}

variable "aws_region" {
  type        = string
  description = "Region the AWS provider reports; LocalStack ignores billing regions."
  default     = "us-east-1"
}

variable "backend_url" {
  type        = string
  description = "Where the Go API is reachable from other services."
  default     = "http://go-backend-service:8080"
}

variable "validator_url" {
  type        = string
  description = "Where the validator is reachable from the backend. In-cluster service or a GPU host."
  default     = "http://kedi-filter-service:8000"
}

variable "storage_endpoint" {
  type        = string
  description = "S3-compatible endpoint host:port the backend uses (in-cluster MinIO or a VPS)."
  default     = "minio-service:9000"
}

variable "storage_bucket" {
  type        = string
  description = "Bucket the backend reads/writes. Defaults to the in-cluster cats bucket."
  default     = "cats"
}

variable "storage_secure" {
  type        = bool
  description = "HTTPS when talking to object storage."
  default     = false
}

variable "ssh_public_key" {
  type        = string
  description = "OpenSSH public key for EC2 key pair."
  default     = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCs7D2J0rwtJz8ZOjnwyDwp/NVx26Zn4v7SW5BkYznB6pkYrbTnJRo4BsYSuB7yrbZs+N4R7wCOgjbDNUrIjGCEsgZaYzT9EaVG4tU3y1/SG1OoBMlQ3w793XW+eIg/z0oQCiLNvLjd1fy9+dSzvluA5VE83hy+ECNmMbZ06/YSvxNOj/A8p48WRsg/3qwus2c7VddcSHLgrlO0oXDINsf67i1Lhb6661FpRDW/f5X9mKXiDnQVumDS1fah1hSP6C/s+ksz0GvtT2Vg/m8Y9+djMPUv7yHoEj9coFlPxOr56U3N39LaAw+9Gn6T9KA8A+ZW0s6po5Yr4mq+Zmq5+GB/ hidircan@localstack"
}
