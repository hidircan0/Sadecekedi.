# syntax=docker/dockerfile:1
FROM pytorch/pytorch:2.2.0-cuda12.1-cudnn8-runtime

ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility
ENV DEBIAN_FRONTEND=noninteractive

# NVIDIA CUDA apt repo'sunu kapat, Ubuntu'yu HTTPS + universe yap.
RUN set -eux; \
    rm -f /etc/apt/apt.conf.d/docker-clean; \
    echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache; \
    rm -f /etc/apt/sources.list.d/cuda*.list \
          /etc/apt/sources.list.d/nvidia*.list; \
    if [ -f /etc/apt/sources.list ]; then \
      sed -i \
        -e 's|http://archive.ubuntu.com/ubuntu|https://archive.ubuntu.com/ubuntu|g' \
        -e 's|http://security.ubuntu.com/ubuntu|https://security.ubuntu.com/ubuntu|g' \
        -e 's/ main restricted$/ main restricted universe/' \
        -e 's/ main$/ main universe/' \
        /etc/apt/sources.list; \
    fi

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    set -eux; \
    apt-get -o Acquire::Retries=5 update; \
    apt-get -o Acquire::Retries=5 install -y --no-install-recommends \
      ca-certificates \
      libglib2.0-0 \
      libgl1 \
      gcc \
      tesseract-ocr \
      curl; \
    tessdir="$(ls -d /usr/share/tesseract-ocr/*/tessdata 2>/dev/null | head -1)"; \
    if [ -z "$tessdir" ]; then tessdir=/usr/share/tesseract-ocr/4.00/tessdata; mkdir -p "$tessdir"; fi; \
    if [ ! -f "$tessdir/tur.traineddata" ]; then \
      curl -fsSL -o "$tessdir/tur.traineddata" \
        https://github.com/tesseract-ocr/tessdata/raw/main/tur.traineddata; \
    fi

WORKDIR /app

COPY requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --upgrade pip && \
    pip install -r requirements.txt onnxruntime-gpu && \
    python -c "import torch; print('torch', torch.__version__, 'cuda_build', torch.version.cuda)"

RUN --mount=type=cache,target=/root/.cache/yolo \
    if [ -f /root/.cache/yolo/yolo11n.pt ]; then \
      cp /root/.cache/yolo/yolo11n.pt /app/yolo11n.pt; \
    else \
      curl -fsSL -o /app/yolo11n.pt https://github.com/ultralytics/assets/releases/download/v8.3.0/yolo11n.pt; \
      mkdir -p /root/.cache/yolo && cp /app/yolo11n.pt /root/.cache/yolo/yolo11n.pt; \
    fi

COPY . .

EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000", "--workers", "1"]