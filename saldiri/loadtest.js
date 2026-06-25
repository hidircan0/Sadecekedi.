import http from 'k6/http';
import { check, sleep } from 'k6';

// Test için kullanacağımız standart kedi fotoğrafını RAM'e alıyoruz.
// Script ile aynı klasörde 'test_kedi.jpg' adında bir dosya olmak zorunda!
const testImage = open('./kopek.jpeg', 'b');

export const options = {
  stages: [
    { duration: '30s', target: 50 },  // 30 saniye içinde 5 eşzamanlı kullanıcıya çık
    { duration: '1m', target: 20 },  // 1 dakika boyunca 10 kullanıcı aralıksız fotoğraf yüklesin
    { duration: '30s', target: 0 },  // Son 30 saniyede yükü sıfırla (Soğuma)
  ],
};

export default function () {
  // Azure sunucu IP'ni veya alan adını buraya yaz
  const url = 'http://sadecekedi.com.tr/upload';

  const data = {
    // FastAPI kodumuzda (image: UploadFile = File(...)) parametre adı "image" olduğu için bunu kullanıyoruz.
	cat_photo: http.file(testImage, 'kopek.jpeg', 'image/jpeg'),
  };

  const res = http.post(url, data);

  // Gelen yanıtların doğruluğunu kontrol ediyoruz
  check(res, {
    'Durum kodu 200 mü?': (r) => r.status === 200,
    'gümrük reddetti mi': (r) => r.body.includes('[!] İHLAL: Bu bir kedi değil!'),
  });

  // Her fotoğraf yüklendikten sonra sunucuya 1 saniye nefes payı veriyoruz
  sleep(1);
}

