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
