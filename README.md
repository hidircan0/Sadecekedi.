Sadece Kedi (Only Cat) - AI-Powered Validation Platform 🐈🛡️
A high-performance, secure, and AI-driven web platform designed exclusively for uploading and validating cat photographs. Built with a strict Zero-Trust DevSecOps architecture, it features a lightweight Go backend that delegates heavy image processing to an isolated Python AI microservice.

🚀 Architecture & Tech Stack
This project is divided into an API/Traffic gateway and an isolated AI Validation Engine, running inside Docker containers and protected by Cloudflare Tunnels.

Frontend: HTMX & pure HTML/CSS (Zero JS bloat).

Core Backend / Traffic Controller: Go (net/http). Handles heavy concurrent traffic efficiently with minimal RAM footprint (Goroutines).

AI Validation Microservice: Python (FastAPI).

Machine Learning & OCR Models:

YOLO11n: Object detection (Strictly validates the presence of a 'cat').

NudeNet: NSFW/Inappropriate content filtering.

Tesseract OCR: Text extraction to block banned words and phone numbers.

Infrastructure & Security: Docker Compose, Cloudflare Tunnels (IP masking & Reverse Proxy), Azure NSG.

Observability: Prometheus & Grafana for real-time memory, CPU, and traffic monitoring.

🛡️ Security Features (Defense in Depth)
Network Layer: True IP is hidden behind Cloudflare Tunnels. Azure NSG drops all external port scans (Filtered).

Application Layer (Magic Bytes Validation): Built-in defense against Application Layer DoS attacks. File extensions are ignored; uploaded files are validated via deep byte inspection (PIL.UnidentifiedImageError) to block disguised malicious payloads (e.g., .txt disguised as .jpg).

Content Layer: Multi-stage AI pipeline ensures no NSFW content, hidden advertisements, or irrelevant images pass the gateway.

📂 Project Structure
Plaintext
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
├── docker-compose.yml        # Multi-container orchestration
├── Dockerfile                # Go backend containerization
└── go.mod & go.sum           # Go dependencies
🛠️ How to Run
Since the application relies on an isolated AI microservice and database, it is recommended to run the entire stack via Docker.

1. Clone the repository:

Bash
git clone https://github.com/yourusername/sadecekedi.git
cd sadecekedi
2. Start the DevSecOps stack (Detached mode):

Bash
docker-compose up -d --build
The Go server will be available at http://localhost:8080, and the Python AI service will run internally.

Local Development (Go Backend Only):

Bash
go run ./cmd/web
📊 Observability
Memory management is strictly monitored. The Python microservice is designed to load heavy ML models into memory and release temporary tensors (breathing/garbage collection) efficiently to prevent memory leaks during high traffic spikes.
