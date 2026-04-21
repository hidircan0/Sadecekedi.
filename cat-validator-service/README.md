# Cat Validator Service

Temporary FastAPI microservice for image validation before YOLO integration.

## Run

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8000
```

## Endpoints

- `GET /health` -> basic health check.
- `POST /validate` -> accepts multipart field `image` and returns mock:
  `{"is_cat": true, "confidence": 0.95}`.
