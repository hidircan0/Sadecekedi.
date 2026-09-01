# Sadece Kedi (Only Cat) - AI-Powered Validation Platform 🐈🛡️

A high-performance, secure, and AI-driven web platform designed exclusively for uploading and validating cat photographs. Built with a strict **Zero-Trust** DevSecOps architecture, it features a lightweight Go backend that delegates heavy image processing to an isolated Python AI microservice.

##  Architecture & Tech Stack

This project is divided into an API/Traffic gateway and an isolated AI Validation Engine, running on Kubernetes and protected by Cloudflare Tunnels.

* **Frontend:** HTMX & pure HTML/CSS (Zero JS bloat).
* **Core Backend / Traffic Controller:** Go (`net/http`). Handles heavy concurrent traffic efficiently with minimal RAM footprint (Goroutines).
* **AI Validation Microservice:** Python (FastAPI).
* **Machine Learning & OCR Models:**
  * **YOLO11n:** Object detection (Strictly validates the presence of a 'cat').
  * **NudeNet:** NSFW/Inappropriate content filtering.
  * **Tesseract OCR:** Text extraction to block banned words and phone numbers.
* **Infrastructure & Security:** Kubernetes, Skaffold, Cloudflare Tunnels (IP masking & Reverse Proxy).
* **Observability:** Prometheus & Grafana for real-time memory, CPU, and traffic monitoring.

##  Security Features (Defense in Depth)

1. **Network Layer:** True IP is hidden behind Cloudflare Tunnels. Azure NSG drops all external port scans (Filtered).
2. **Application Layer (Magic Bytes Validation):** Built-in defense against `Application Layer DoS` attacks. File extensions are ignored; uploaded files are validated via deep byte inspection (`PIL.UnidentifiedImageError`) to block disguised malicious payloads (e.g., `.txt` disguised as `.jpg`).
3. **Content Layer:** Multi-stage AI pipeline ensures no NSFW content, hidden advertisements, or irrelevant images pass the gateway.

## 📂 Project Structure
```text
.
├── .github/                  # CI/CD Pipelines (GitHub Actions)
├── cat-validator-service/    # Python/FastAPI Microservice (YOLO, NudeNet, OCR)
├── cmd/web/                  # Go application entrypoint & dependency wiring
├── internal/
│   ├── app/                  # Use-case & Service layer
│   ├── domain/               # Domain entities
│   ├── http/                 # Go Handlers & Router
│   └── storage/local/        # Local filesystem repository
├── web/
│   ├── static/               # CSS and static assets
│   └── templates/            # HTML templates and HTMX partials
├── k8s/                      # Kubernetes manifests
├── skaffold.yaml             # Build & deploy orchestration
├── Dockerfile                # Go backend containerization
└── go.mod & go.sum           # Go dependencies
```

##  How to Run

**1. Clone the repository:**
```bash
git clone https://github.com/hidircanaslan/sadecekedi.git
cd sadecekedi
```

**2. Secrets dosyasını hazırla (bir kez):**
```bash
# k8s/01-config-secrets.yaml oluştur, değerleri doldur
```

**3. Tüm stack:**
```bash
skaffold dev -n sadecekedi --cache-artifacts=false
```

Namespace, secrets, build, k3s import ve deploy tek komutta. Port forward: 8080 (backend), 8000 (validator), 9001 (minio console), 3000 (grafana).

**Local Development (Go Backend Only):**
```bash
go run ./cmd/web
```

##  Observability

Memory management is strictly monitored. The Python microservice is designed to load heavy ML models into memory and release temporary tensors (breathing/garbage collection) efficiently to prevent memory leaks during high traffic spikes.
