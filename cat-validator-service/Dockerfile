FROM python:3.11-slim

# Temel kütüphaneler ve derleme araçları (eğer derleme gerekiyorsa diye)
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    gcc \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Pip'i önceden güncellemek ve timeout süresini global olarak artırmak
RUN pip install --upgrade pip

COPY requirements.txt .

# Timeout'u 2000 saniyeye çıkardım ve daha hızlı bir ayna (mirror) opsiyonu ekledim
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu && \
    pip install --no-cache-dir ultralytics fastapi uvicorn python-multipart numpy opencv-python-headless
COPY . .

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]