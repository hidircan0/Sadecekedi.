from fastapi import FastAPI, UploadFile, File
from ultralytics import YOLO
import io
import os
from PIL import Image
import pytesseract
import re
from nudenet import NudeClassifier # Hazır NSFW Kütüphanesi

app = FastAPI()

model = YOLO("yolo11n.pt") 
nsfw_classifier = NudeClassifier() # Hazır modeli belleğe yükler

BANNED_WORDS = ["satılık", "uyuşturucu", "hap", "numaram", "telegram", "fiyat", "eskort", "dm"]

@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "cat-validator-victus-v2-nsfw"}

@app.post("/validate")
async def validate_cat(image: UploadFile = File(...)):
    contents = await image.read()
    
    # NudeNet için dosyayı geçici olarak kaydedelim
    temp_filename = f"temp_{image.filename}"
    with open(temp_filename, "wb") as f:
        f.write(contents)
    
    try:
        # --- 0. KATMAN: NSFW (UYGUNSUZLUK) KONTROLÜ ---
        # NudeNet fotoğrafı tarar. Sonuç şöyle döner: {'temp.jpg': {'safe': 0.1, 'unsafe': 0.9}}
        nsfw_result = nsfw_classifier.classify(temp_filename)
        
        # Eğer dosya analizi başarılıysa unsafe (uygunsuz) skoruna bakalım
        if temp_filename in nsfw_result:
            unsafe_score = nsfw_result[temp_filename].get('unsafe', 0)
            if unsafe_score > 0.4: # %40'tan fazla uygunsuzluk sezerse acımaz!
                print(f"🚨 GÜVENLİK İHLALİ: NSFW Tespit Edildi! Skor: {unsafe_score}")
                return {"is_cat": False, "reason": "Uygunsuz içerik tespit edildi!"}

        # --- 1. KATMAN: OCR (METİN VE NUMARA AVI) ---
        img = Image.open(io.BytesIO(contents))
        extracted_text = pytesseract.image_to_string(img).lower()
        
        phone_pattern = re.compile(r'\+?\d{10,13}')
        if phone_pattern.search(extracted_text):
            return {"is_cat": False, "reason": "Fotoğrafta telefon numarası tespit edildi!"}
        
        for word in BANNED_WORDS:
            if word in extracted_text:
                return {"is_cat": False, "reason": f"Yasaklı metin tespit edildi!"}

        # --- 2. KATMAN: YOLO (KEDİ TESPİTİ) ---
        results = model.predict(img, conf=0.15) 
        
        is_cat = False
        for r in results:
            for box in r.boxes:
                label = model.names[int(box.cls)]
                if label == 'cat':
                    is_cat = True
                    break

        # --- 3. KATMAN: KARAR ---
        if is_cat:
            return {"is_cat": True, "reason": "Gümrük onayladı, saf kedi!"}
            
        return {"is_cat": False, "reason": "Kedi bulunamadı veya görüntü kalitesiz."}
        
    finally:
        # İşlem bitince geçici dosyayı sunucudan sil ki yer kaplamasın
        if os.path.exists(temp_filename):
            os.remove(temp_filename)