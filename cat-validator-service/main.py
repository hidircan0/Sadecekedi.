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
    results = model(img)
    
    is_cat = False
    for r in results:
        for c in r.boxes.cls:
            label = model.names[int(c)]
            if label == 'cat':
                is_cat = True
                break
                
    return {"is_cat": is_cat}
