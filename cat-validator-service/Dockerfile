FROM python:3.11-slim

# --- SİBER GÜVENLİK GÜNCELLEMESİ ---
# tesseract-ocr (İngilizce/Rakam için) ve tesseract-ocr-tur (Türkçe karakterler için) eklendi
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    gcc \
    python3-dev \
    tesseract-ocr \
    tesseract-ocr-tur \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN pip install --upgrade pip

# requirements.txt kopyalanıyor (Ama ana kurulumu altta manuel yapıyoruz)
COPY requirements.txt .

# --- KÜTÜPHANE GÜNCELLEMESİ ---
# En sona 'pytesseract' ve 'nudenet' eklendi
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu && \
    pip install --no-cache-dir ultralytics fastapi uvicorn python-multipart numpy opencv-python-headless pytesseract nudenet

COPY . .

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]