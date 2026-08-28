FROM python:3.11-slim-bookworm

# 1. Debian mirror'larını HTTPS'e çevir (bağlantı resetlenmelerini önler)
RUN sed -i 's|http://deb.debian.org|https://deb.debian.org|g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
    sed -i 's|http://deb.debian.org|https://deb.debian.org|g' /etc/apt/sources.list

# 2. Sistem paketleri + curl
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libgl1 \
    libglib2.0-0 \
    gcc \
    python3-dev \
    tesseract-ocr \
    tesseract-ocr-tur \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 3. Pip güncellemesi + PyTorch CPU (Cachelenir)
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir torch torchvision --index-url https://download.pytorch.org/whl/cpu

# 4. requirements.txt dosyasını al ve kalan paketleri kur (Cachelenir)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 5. Modeli indir (Cachelenir)
RUN curl -L -o yolo11n.pt https://github.com/ultralytics/assets/releases/download/v8.3.0/yolo11n.pt

# 6. Uygulama kodlarını kopyala
COPY . .

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "2"]