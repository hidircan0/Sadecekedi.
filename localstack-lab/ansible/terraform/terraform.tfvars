project     = "localstack-lab"
environment = "local"
aws_region  = "us-east-1"

# In-cluster defaults. For a GPU box: validator_url = "http://192.168.x.x:8000"
backend_url      = "http://go-backend-service:8080"
validator_url    = "http://kedi-filter-service:8000"
storage_endpoint = "minio-service:9000"
storage_bucket   = "cats"
storage_secure   = false
