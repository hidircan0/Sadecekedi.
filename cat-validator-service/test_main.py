import sys
from unittest.mock import MagicMock

# CI ortamında ağır ML modellerini import etmeden endpoint testleri çalıştır.
sys.modules.setdefault("ultralytics", MagicMock())
sys.modules.setdefault("nudenet", MagicMock())
sys.modules.setdefault("pytesseract", MagicMock())

from fastapi.testclient import TestClient  # noqa: E402
from main import BANNED_WORDS, app  # noqa: E402


def test_health_endpoint():
    client = TestClient(app)
    response = client.get("/health")

    assert response.status_code == 200
    payload = response.json()
    assert payload["status"] == "healthy"
    assert "cat-validator" in payload["service"]


def test_banned_words_list_not_empty():
    assert len(BANNED_WORDS) > 0
    assert all(isinstance(word, str) and word for word in BANNED_WORDS)
