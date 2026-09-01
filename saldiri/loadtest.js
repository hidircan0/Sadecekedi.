import http from 'k6/http';
import { check, sleep } from 'k6';

const TARGET_URL = 'http://127.0.0.1:8080/upload';
const dogImage = open('./kopek.jpeg', 'b');

export const options = {
  stages: [
    { duration: '15s', target: 25 },
    { duration: '30s', target: 50 },
    { duration: '30s', target: 100 },
    { duration: '15s', target: 50 },
    { duration: '10s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1500'], // P95 < 1.5s
  },
};

export default function () {
  const payload = {
    image: http.file(dogImage, 'kopek.jpeg', 'image/jpeg'),
  };

  const params = {
    // 400 ve 422'yi beklenen/başarılı savunma yanıtı say:
    responseType: 'text',
    expectedStatuses: [200, 400, 422],
  };

  const res = http.post(TARGET_URL, payload, params);

  check(res, {
    'Gümrük Köpeği Reddetti (400/422)': (r) => r.status === 400 || r.status === 422,
    'Sunucu Hatası Yok (Not 500)': (r) => r.status !== 500,
  });

  sleep(0.05);
}