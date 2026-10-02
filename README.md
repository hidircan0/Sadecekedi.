# Sadecekedi

Sadecekedi is a cat image upload and validation platform. A Go web application handles requests and persistence, while an isolated FastAPI service validates uploaded images with object detection, content filtering, and OCR.

The project also includes a local infrastructure lab built with LocalStack, Terraform, Ansible, and k3s.

## Architecture

- **Web application:** Go, `net/http`, HTML templates, and HTMX
- **Image validation:** Python, FastAPI, YOLO, NudeNet, and Tesseract OCR
- **Database:** PostgreSQL
- **Object storage:** S3-compatible storage
- **Container orchestration:** k3s and Kubernetes manifests
- **Ingress:** Traefik
- **External access:** Cloudflare Tunnel
- **Observability:** Prometheus, Grafana, and Node Exporter
- **Autoscaling:** Kubernetes HPA and KEDA
- **Infrastructure lab:** LocalStack, Terraform, and Ansible
- **CI/CD:** GitHub Actions and GitHub Container Registry

## Request Flow

```text
Client
  |
  v
Cloudflare Tunnel / Traefik
  |
  v
Go backend
  |-- PostgreSQL
  |-- S3-compatible object storage
  `-- FastAPI validator
        |-- YOLO
        |-- NudeNet
        `-- Tesseract OCR
```

## Project Structure

```text
.
├── .github/workflows/        GitHub Actions workflows
├── cat-validator-service/    FastAPI image validation service
├── cmd/web/                  Go application entry point
├── internal/                 Domain, application, HTTP, and storage packages
├── k8s/                      Kubernetes manifests
├── localstack-lab/
│   ├── ansible/              Build and deployment orchestration
│   └── ansible/terraform/    LocalStack infrastructure definitions
├── scripts/                  Secret rendering and k3s image import scripts
├── web/                      HTML templates and static assets
├── Dockerfile                Go backend image
└── skaffold.yaml             Local Kubernetes development configuration
```

## Requirements

- Linux
- Docker
- k3s
- Terraform
- Python 3
- AWS CLI
- `curl`
- `sudo` access for k3s operations

KEDA must be installed once because the deployment includes a `ScaledObject`:

```bash
helm repo add kedacore https://kedacore.github.io/charts
helm repo update
helm install keda kedacore/keda --namespace keda --create-namespace
```

## Configuration

Create the application environment file:

```bash
cp .env.example .env
```

Fill every required value in `.env`, then generate the Kubernetes Secret manifest:

```bash
./scripts/render-k8s-secrets.sh
```

Create the LocalStack environment file:

```bash
cp localstack-lab/.env.example localstack-lab/.env
```

Replace `SSH_PUBLIC_KEY` with your own OpenSSH public key when needed. Do not commit either `.env` file or the generated Kubernetes Secret.

## Start the Complete Local Environment

```bash
cd localstack-lab
make up
```

This command:

1. starts LocalStack;
2. provisions the simulated AWS resources with Terraform;
3. builds the backend and validator images with Ansible;
4. imports the images into k3s;
5. applies the Kubernetes manifests;
6. starts local port forwarding.

Local endpoints:

- Application: `http://localhost:8080`
- Grafana: `http://localhost:3000`
- Traefik dashboard: `http://localhost:9080/dashboard/`
- LocalStack API: `http://localhost:4566`

Stop and remove the local application environment:

```bash
cd localstack-lab
make down
```

## Development

Run Go checks:

```bash
go vet ./...
go test ./... -race -count=1
```

Run validator tests:

```bash
cd cat-validator-service
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt pytest httpx
pytest -v
```

Run the CPU Kubernetes development profile:

```bash
skaffold dev -n sadecekedi -p cpu
```

## CI/CD

The CI workflow validates the Go and Python code and verifies both container builds on GitHub-hosted runners.

After CI succeeds on `main`, the CD workflow publishes versioned backend and validator images to GitHub Container Registry. Remote deployment is provided as a disabled template until a target server and its SSH secrets are configured.

## Security Notes

- Secrets are generated from ignored environment files.
- Uploaded files are validated by content rather than filename alone.
- The validator runs as a separate service from the web application.
- Cloudflare Tunnel can expose selected services without opening inbound host ports.
- The Traefik dashboard should only be exposed on trusted development networks.
