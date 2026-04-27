from fastapi import FastAPI, UploadFile, File
from ultralytics import YOLO
import io
from PIL import Image

app = FastAPI()
model = YOLO("yolo11n.pt") 
# Mevcut importların altına ekle
@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "cat-validator-victus"}

# Mevcut /validate kodu burada kalmaya devam edecek...
@app.post("/validate")
async def validate_cat(image: UploadFile = File(...)):
    contents = await image.read()
    img = Image.open(io.BytesIO(contents))
    results = model.predict(img, conf=0.2)
    
    is_cat = False
   for r in results:
        for box in r.boxes:
            label = model.names[int(box.cls)]
            confidence = float(box.conf) # Güven skoru
            print(f"Tespit edilen: {label} - Güven Skoru: {confidence}") # Loglara basar
            
            if label == 'cat' and confidence > 0.15: # Buradan da kontrol edebilirsin
                is_cat = True
                break
                
    return {"is_cat": is_cat}
