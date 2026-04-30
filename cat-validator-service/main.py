from fastapi import FastAPI, UploadFile, File
from ultralytics import YOLO
import io
import os
from PIL import Image, UnidentifiedImageError # YENİ: Sahte dosya yakalayıcı zırh eklendi
import pytesseract
import re
from nudenet import NudeClassifier 

app = FastAPI()

model = YOLO("yolo11m.pt") 
nsfw_classifier = NudeClassifier() 

BANNED_WORDS = ["satılık", "uyuşturucu", "hap", "numaram", "telegram", "fiyat", "eskort", "dm", "alp bora songül"]

@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "cat-validator-victus-v2-nsfw"}

@app.post("/validate")
async def validate_cat(image: UploadFile = File(...)):
    contents = await image.read()
    temp_filename = f"temp_{image.filename}"
    
    try:
        # --- 0. KATMAN: SAHTE DOSYA KORUMASI (MAGIC BYTES ARAMASI) ---
        # Pillow uzantıya değil, dosyanın DNA'sına bakar. Eğer bu bir txt ise,
        # anında UnidentifiedImageError fırlatır ve kod aşağıya inmeden except bloğuna uçar.
        img = Image.open(io.BytesIO(contents))
        img.load() # Görüntüyü tam anlamıyla belleğe alıp kırık/bozuk olmadığını doğrular
        
        # Eğer buraya kadar geldiyse dosya GERÇEKTEN bir fotoğraftır. Şimdi diske yazabiliriz.
        with open(temp_filename, "wb") as f:
            f.write(contents)
            
        # --- 1. KATMAN: NSFW (UYGUNSUZLUK) KONTROLÜ ---
        nsfw_result = nsfw_classifier.classify(temp_filename)
        if temp_filename in nsfw_result:
            unsafe_score = nsfw_result[temp_filename].get('unsafe', 0)
            if unsafe_score > 0.4: 
                print(f"🚨 GÜVENLİK İHLALİ: NSFW Tespit Edildi! Skor: {unsafe_score}")
                return {"is_cat": False, "reason": "Uygunsuz içerik tespit edildi!"}

        # --- 2. KATMAN: OCR (METİN VE NUMARA AVI) ---
        extracted_text = pytesseract.image_to_string(img).lower()
        
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
        # Arkadaşının denediği o "uzantısı değiştirilmiş .txt" saldırısı tam olarak buraya düşer
        print(f"🚨 SİBER SAVUNMA: Sahte dosya uzantısı engellendi! ({image.filename})")
        return {"is_cat": False, "reason": "Sistemi kandıramazsın, bu geçerli bir fotoğraf değil!"}
        
    except Exception as e:
        # Sistemin ne olursa olsun çökmesini (HTTP 500) engelleyen ana kalkan
        print(f"🚨 SİSTEM HATASI: Beklenmeyen bir durum oluştu -> {e}")
        return {"is_cat": False, "reason": "Görüntü analiz edilirken sunucuda bir hata oluştu."}
        
    finally:
        # Kod başarılı da olsa, saldırı da gelse bu blok ÇALIŞMAK ZORUNDADIR. 
        # Diskte çöp dosya kalmasını engeller.
        if os.path.exists(temp_filename):
            os.remove(temp_filename)