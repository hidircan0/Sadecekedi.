# Runtime addresses are not AWS resources. Terraform only records them and
# writes env/ConfigMap files so Skaffold and host processes share one contract.

resource "local_file" "backend_env" {
  filename = "${path.module}/generated/backend.env"
  content = templatefile("${path.module}/templates/backend.env.tftpl", {
    validator_url    = var.validator_url
    storage_endpoint = var.storage_endpoint
    storage_bucket   = var.storage_bucket
    storage_secure   = var.storage_secure
  })
}

resource "local_file" "validator_env" {
  filename = "${path.module}/generated/validator.env"
  content = templatefile("${path.module}/templates/validator.env.tftpl", {
    backend_url = var.backend_url
  })
}

resource "local_file" "k8s_endpoints" {
  filename = "${path.module}/generated/k8s-endpoints.yaml"
  content = templatefile("${path.module}/templates/k8s-endpoints.yaml.tftpl", {
    validator_url    = var.validator_url
    storage_endpoint = var.storage_endpoint
    storage_bucket   = var.storage_bucket
    storage_secure   = var.storage_secure
  })
}
