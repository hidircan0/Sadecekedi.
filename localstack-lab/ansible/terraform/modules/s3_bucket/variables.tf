variable "bucket_name" {
  type        = string
  description = "Bucket name. Must look globally unique even on LocalStack."
}

variable "tags" {
  type        = map(string)
  description = "Tags copied onto the bucket."
  default     = {}
}
