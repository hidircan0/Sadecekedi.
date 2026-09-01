import http from 'k6/http';
import { check, sleep } from 'k6';

// Test için kullanılacak gerçek kedi görseli
const catImage = open('./kedi.jpeg', 'b');

export const options = {
  stages: [
    { duration: '20s', target: 20 },  // 1. Isınma (20 VU)
    { duration: '40s', target: 50 },  // 2. Orta Yük (50 VU)
    { duration: '45s', target: 100 }, // 3. Zirve Stres (100 VU - GPU + DB + S3 Sınırı)
    { duration: '20s', target: 30 },  // 4. Kademeli Düşüş
    { duration: '15s', target: 0 },   // 5. Soğuma
  ],
  thresholds: {
    // Tüm pipeline (YOLO + NSFW + OCR + MinIO + DB) 100 VU altında P95 < 4s kalmalı
    http_req_duration: ['p(95)<4000'],
    // 500 hatası ve timeout toleransı <%2 olmalı
    http_req_failed: ['rate<0.02'],
    // İş kuralları doğrulaması %95+ olmalı
    checks: ['rate>0.95'],
  },
};

export default function () {
  const url = 'http://127.0.0.1:8080/upload';

  // MinIO üzerinde dosya adı çakışmasını önleyen benzersiz isimlendirme
  const uniqueFileName = `cat_vu${__VU}_it${__ITER}_${Date.now()}.jpeg`;

  const payload = {
    // Go backend'in beklediği multipart form alanı ('cat_photo' veya 'image')
    cat_photo: http.file(catImage, uniqueFileName, 'image/jpeg'),
  };

  const params = {
    headers: {
      'User-Agent': 'k6-cat-stress-test/2.0',
    },
    timeout: '15s',
  };

  const res = http.post(url, payload, params);

  check(res, {
    '1. Sunucu Başarılı (200 OK)': (r) => r.status === 200,
    '2. HTML/JSON Onay Geldi': (r) => {
      if (!r.body) return false;
      return (
        r.body.includes('ONAY') ||
        r.body.includes('saf kedi') ||
        r.body.includes('success') ||
        r.body.includes('yüklendi') ||
        !r.body.includes('text-error')
      );
    },
    '3. 500 Sunucu Hatası Yok': (r) => r.status !== 500,
  });

  // VU döngü aralığı (0.2s ile hızlı ve dengeli akış)
  sleep(0.2);
}