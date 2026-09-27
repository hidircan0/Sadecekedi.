module "demo_bucket" {
  source      = "./modules/s3_bucket"
  bucket_name = "${local.name_prefix}-bucket"
  tags        = local.tags
}

resource "aws_dynamodb_table" "items" {
  name         = "${local.name_prefix}-items"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}
