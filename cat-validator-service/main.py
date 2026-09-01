from fastapi import FastAPI, UploadFile, File
from ultralytics import YOLO
import io
import os
import uuid
from PIL import Image, ImageOps, UnidentifiedImageError
import pytesseract
import re
from nudenet import NudeDetector
import torch

app = FastAPI()

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
print(f"[*] GÜMRÜK BAŞLATILDI -> Aktif Donanım: {DEVICE.upper()}")

model = YOLO("yolo11n.pt")
model.to(DEVICE)

nsfw_detector = NudeDetector()
BANNED_WORDS = ["satılık", "uyuşturucu", "hap", "numaram", "telegram", "fiyat", "eskort", "dm", "alp bora songül"]

@app.get("/health")
def health_check():
    return {"status": "ok", "device": DEVICE}

# DİKKAT: 'async def' yerine 'def' yapıldı (Threadpool kalkanı)
@app.post("/validate")
def validate_cat(image: UploadFile = File(...)):
    contents = image.file.read()
    
    unique_id = uuid.uuid4().hex
    temp_filename = f"/tmp/temp_{unique_id}.jpg"
    
    try:
        # 0. GÖRSELİ BELLEKTE DOĞRULA VE HAZIRLA
        img = Image.open(io.BytesIO(contents))
        img = ImageOps.exif_transpose(img)
        img.load()
        if img.mode != 'RGB':
            img = img.convert('RGB')
        img.save(temp_filename, format="JPEG")

        # --- 1. KATMAN: JET HIZINDA GPU YOLO KONTROLÜ (15ms) ---
        # Kedi değilse CPU'yu hiç yormadan ANINDA reddet!
        results = model.predict(temp_filename, conf=0.45, device=DEVICE, verbose=False)
        is_cat = any(model.names[int(box.cls)] == 'cat' for r in results for box in r.boxes)

        if not is_cat:
            return {"is_cat": False, "reason": "Kedi bulunamadı veya görüntü kalitesiz."}

        # --- 2. KATMAN: NSFW KONTROLÜ (Sadece kedi geçenlere uygulanır) ---
        nsfw_result = nsfw_detector.detect(temp_filename)
        for detection in nsfw_result:
            if detection.get('score', 0) > 0.4:
                return {"is_cat": False, "reason": "Uygunsuz içerik tespit edildi!"}

        # --- 3. KATMAN: OCR KONTROLÜ (Sadece kedi geçenlere uygulanır) ---
        extracted_text = pytesseract.image_to_string(temp_filename).lower()
        if re.search(r'\+?\d{10,13}', extracted_text):
            return {"is_cat": False, "reason": "Fotoğrafta telefon numarası tespit edildi!"}
        
        for word in BANNED_WORDS:
            if word in extracted_text:
                return {"is_cat": False, "reason": "Yasaklı metin tespit edildi!"}
        return {"is_cat": True, "reason": "Gümrük onayladı, saf kedi!"}

    except UnidentifiedImageError:
        return {"is_cat": False, "reason": "Geçersiz dosya formatı!"}
    except Exception as e:
        return {"is_cat": False, "reason": f"Hata: {str(e)}"}
    finally:
        if os.path.exists(temp_filename):
            os.remove(temp_filename)