from fastapi import FastAPI, UploadFile, File
from ultralytics import YOLO
import io
import os
import uuid # YENİ: Çarpışma (Race condition) önleyici eklendi
from PIL import Image, UnidentifiedImageError
import pytesseract
import re
from nudenet import NudeDetector

app = FastAPI()

# MLOps Standartları: Daha keskin model
model = YOLO("yolo11m.pt") 
nsfw_detector = NudeDetector()

BANNED_WORDS = ["satılık", "uyuşturucu", "hap", "numaram", "telegram", "fiyat", "eskort", "dm", "alp bora songül"]

@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "cat-validator-victus-v2-nsfw"}

@app.post("/validate")
async def validate_cat(image: UploadFile = File(...)):
    contents = await image.read()
    
    # --- 0. KATMAN: AĞ SAVUNMASI (RACE CONDITION) ---
    # Her dosyaya benzersiz isim vererek dosyaların birbirini ezmesini engelliyoruz.
    # Uzantıyı ne olursa olsun .jpg olarak kilitliyoruz.
    unique_id = uuid.uuid4().hex
    temp_filename = f"temp_{unique_id}.jpg"
    
    try:
        # --- 0.5. KATMAN: SAHTE DOSYA KORUMASI VE VERİ STANDARTİZASYONU ---
        img = Image.open(io.BytesIO(contents))
        img.load() 
        
        # Format MPO, WEBP, PNG fark etmez, yapay zekanın midesi bulanmasın diye saf RGB yapıyoruz.
        if img.mode != 'RGB':
            img = img.convert('RGB')
            
        # Orijinal dosyayı (with open...) ile DEĞİL, temizlenmiş halini güvenle diske yazıyoruz.
        img.save(temp_filename, format="JPEG")
            
        # --- 1. KATMAN: NSFW (UYGUNSUZLUK) KONTROLÜ ---
        nsfw_result = nsfw_detector.detect(temp_filename)
        
        for detection in nsfw_result:
            if detection.get('score', 0) > 0.4: 
                print(f"🚨 GÜVENLİK İHLALİ: NSFW ({detection.get('class')}) Tespit Edildi! Skor: {detection.get('score')}")
                return {"is_cat": False, "reason": "Uygunsuz içerik tespit edildi!"}

        # --- 2. KATMAN: OCR (METİN VE NUMARA AVI) ---
        # Tesseract'ın çökmemesi için bellekteki img objesini değil, diske yazdığımız standart JPG'yi okutuyoruz!
        extracted_text = pytesseract.image_to_string(temp_filename).lower()
        
        phone_pattern = re.compile(r'\+?\d{10,13}')
        if phone_pattern.search(extracted_text):
            return {"is_cat": False, "reason": "Fotoğrafta telefon numarası tespit edildi!"}
        
        for word in BANNED_WORDS:
            if word in extracted_text:
                return {"is_cat": False, "reason": f"Yasaklı metin tespit edildi!"}

        # --- 3. KATMAN: YOLO (KEDİ TESPİTİ) ---
        results = model.predict(temp_filename, conf=0.45) 
        
        is_cat = False
        for r in results:
            for box in r.boxes:
                label = model.names[int(box.cls)]
                if label == 'cat':
                    is_cat = True
                    break

        # --- 4. KATMAN: KESİN KARAR ---
        if is_cat:
            return {"is_cat": True, "reason": "Gümrük onayladı, saf kedi!"}
            
        return {"is_cat": False, "reason": "Kedi bulunamadı veya görüntü kalitesiz."}
        
    except UnidentifiedImageError:
        print(f"🚨 SİBER SAVUNMA: Sahte dosya uzantısı engellendi! ({image.filename})")
        return {"is_cat": False, "reason": "Sistemi kandıramazsın, bu geçerli bir fotoğraf değil!"}
        
    except Exception as e:
        print(f"🚨 SİSTEM HATASI: Beklenmeyen bir durum oluştu -> {e}")
        return {"is_cat": False, "reason": "Görüntü analiz edilirken sunucuda bir hata oluştu."}
        
    finally:
        if os.path.exists(temp_filename):
            os.remove(temp_filename)