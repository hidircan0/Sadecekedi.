package handlers

import (
	"context"
	"errors"
	"html/template"
    "io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
    "strings"
    "os"
	"cat-uploader-go/internal/app"
	"cat-uploader-go/internal/domain"
)

const maxUploadSize = 8 << 20

// 1. SADECEKEDİ Servis Arayüzü
type PhotoService interface {
	Upload(ctx context.Context, source multipart.File, originalName string) (domain.Photo, error)
	List(ctx context.Context) ([]domain.Photo, error)
	GetRandom(ctx context.Context) (domain.Photo, error)
	GetNext(ctx context.Context, id uint) (domain.Photo, error)
	GetPrevious(ctx context.Context, id uint) (domain.Photo, error)
	Delete(ctx context.Context, id uint) error
    GetPhotoStream(ctx context.Context, filename string) (io.ReadCloser, string, error)
}

// 2. Gözden kaçan meşhur Struct'lar
type HomeHandler struct {
	service PhotoService
	tmpl    *template.Template
}

type HomeViewData struct {
	Photos    []domain.Photo
	AdminMode bool
	IsZulaPage bool
	TargetKediID string
	Error     template.HTML
	Success   template.HTML
}

func NewHomeHandler(service PhotoService) (*HomeHandler, error) {
	templateFuncs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	tmpl, err := template.New("layout").Funcs(templateFuncs).ParseFiles(
		"web/templates/layout.html",
		"web/templates/index.html",
		"web/templates/partials/photo_list.html",
		"web/templates/partials/single_cat.html",
	)
	if err != nil {
		return nil, err
	}
	return &HomeHandler{service: service, tmpl: tmpl}, nil
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
    // URL'de "kedi=" yazıyorsa o numarayı yakala
    kediID := r.URL.Query().Get("kedi") 
    
    h.renderPage(w, HomeViewData{
        TargetKediID: kediID,
    }, http.StatusOK)
}

// --- SADECEKEDİ: RASTGELE, ÖNCEKİ, SONRAKİ ---
func (h *HomeHandler) Random(w http.ResponseWriter, r *http.Request) {
    setHTMLContentType(w)
    photo, err := h.service.GetRandom(r.Context())
    if err != nil {
        w.Write([]byte("<p>[!] Veritabanında hiç kedi yok.</p>"))
        return
    }
    
    // Kedinin kaçıncı sırada olduğunu buluyoruz
    rank, total := h.getCatStats(r.Context(), photo.ID)
    
    // Veriyi paketleyip HTML'e yolluyoruz
    h.tmpl.ExecuteTemplate(w, "single_cat", map[string]interface{}{
        "Photo": photo,
        "Rank":  rank,
        "Total": total,
    })
}

func (h *HomeHandler) Next(w http.ResponseWriter, r *http.Request) {
    setHTMLContentType(w)
    id, _ := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)

    photo, err := h.service.GetNext(r.Context(), uint(id))
    if err != nil {
        h.Random(w, r)
        return
    }
    
    rank, total := h.getCatStats(r.Context(), photo.ID)
    h.tmpl.ExecuteTemplate(w, "single_cat", map[string]interface{}{
        "Photo": photo, "Rank": rank, "Total": total,
    })
}

func (h *HomeHandler) Previous(w http.ResponseWriter, r *http.Request) {
    setHTMLContentType(w)
    id, _ := strconv.ParseUint(r.URL.Query().Get("id"), 10, 32)

    photo, err := h.service.GetPrevious(r.Context(), uint(id))
    if err != nil {
        h.Random(w, r)
        return
    }
    
    rank, total := h.getCatStats(r.Context(), photo.ID)
    h.tmpl.ExecuteTemplate(w, "single_cat", map[string]interface{}{
        "Photo": photo, "Rank": rank, "Total": total,
    })
}

// BU YENİ: Kedinin aktif kedi listesindeki kaçıncı sırada olduğunu bulur
func (h *HomeHandler) getCatStats(ctx context.Context, currentID uint) (int, int) {
    photos, err := h.service.List(ctx)
    if err != nil || len(photos) == 0 {
        return 1, 1
    }
    
    rank := 1
    for i, p := range photos {
        if p.ID == currentID {
            rank = i + 1
            break
        }
    }
    return rank, len(photos)
}

// --- GÜMRÜK BÖLÜMÜ (Artık Sıkıştırma Yapmıyor, Orijinali Yolluyor) ---
func (h *HomeHandler) Upload(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseMultipartForm(maxUploadSize); err != nil {
        http.Error(w, "invalid multipart form", http.StatusBadRequest)
        return
    }

    file, fileHeader, err := r.FormFile("cat_photo")
    if err != nil {
        http.Error(w, "missing file input", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // ORİJİNAL, yüksek kaliteli fotoğrafı Servis'e (YOLO'ya) yolluyoruz
    if _, err := h.service.Upload(r.Context(), file, filepath.Base(fileHeader.Filename)); err != nil {
        if errors.Is(err, app.ErrInvalidFileType) || errors.Is(err, app.ErrEmptyFile) || errors.Is(err, app.ErrFileNameCollision) || errors.Is(err, app.ErrNotCat) {
            h.renderPartialFeedback(w, err.Error(), "", http.StatusOK)
            return
        }
        h.renderPartialFeedback(w, "Sistem hatası: "+err.Error(), "", http.StatusInternalServerError)
        return
    }

    h.renderPartialFeedback(w, "", "Yapay zeka onayladı, sıkıştırıldı ve arşive eklendi.", http.StatusOK)
}
// Gümrük bildirimlerini HTMX'in istediği saf HTML'e çevirdik
func (h *HomeHandler) renderPartialFeedback(w http.ResponseWriter, errorText string, successText string, status int) {
	setHTMLContentType(w)
	w.WriteHeader(status)
	if errorText != "" {
		w.Write([]byte(`<span class="text-error">[!] İHLAL: ` + errorText + `</span>`))
	} else if successText != "" {
		w.Write([]byte(`<span class="text-success">[+] ONAY: ` + successText + `</span>`))
	}
}
func (h *HomeHandler) renderPage(w http.ResponseWriter, data HomeViewData, status int) {
    setHTMLContentType(w)
    w.WriteHeader(status)
    if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
        // İŞTE BURASI: Artık hatanın ne olduğunu tarayıcıya basacak!
        http.Error(w, "failed to render page: "+err.Error(), http.StatusInternalServerError)
    }
}

/// --- SADECEKEDİ: ZULA (ARŞİV) ODAKLI BÖLÜM ---
func (h *HomeHandler) Zula(w http.ResponseWriter, r *http.Request) {
    setHTMLContentType(w)
    isAdmin := false

    // GÜVENLİK DUVARI: Biri root olmak isterse kimlik sor!
    if r.URL.Query().Get("mode") == "root" {
        user, pass, ok := r.BasicAuth()

        expectedUser := os.Getenv("ADMIN_USER")
        expectedPass := os.Getenv("ADMIN_PASS")
        
        // DİKKAT: KULLANICI ADI VE ŞİFREYİ BURAYA YAZ (Örn: hidir / kedi123)
        if !ok || user !=  expectedUser|| pass != expectedPass {
            w.Header().Set("WWW-Authenticate", `Basic realm="Gizli Operasyon Paneli"`)
            http.Error(w, "Sisteme sızma girişimi engellendi. Yetkiniz yok.", http.StatusUnauthorized)
            return
        }
        isAdmin = true // Şifre doğruysa yetkiyi ver
    }

    photos, err := h.service.List(r.Context())
    if err != nil {
        http.Error(w, "Arşive ulaşılamadı", http.StatusInternalServerError)
        return
    }

    h.tmpl.ExecuteTemplate(w, "layout", HomeViewData{
        Photos:     photos,
        AdminMode:  isAdmin,
        IsZulaPage: true,
    })
}
func (h *HomeHandler) Delete(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // SİLME İŞLEMİ İÇİN KESİN KONTROL (Arka kapıdan istek atılmasını engeller)
    user, pass, ok := r.BasicAuth()
    expectedUser := os.Getenv("ADMIN_USER")
    expectedPass := os.Getenv("ADMIN_PASS")
    if !ok || user != expectedUser|| pass != expectedPass {
        http.Error(w, "Yetkisiz silme girişimi!", http.StatusForbidden)
        return
    }

    idStr := filepath.Base(r.URL.Path)
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    if err := h.service.Delete(r.Context(), uint(id)); err != nil {
        http.Error(w, "Silme başarısız", http.StatusInternalServerError)
        return
    }

    setHTMLContentType(w)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(""))
}
func setHTMLContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}
// Zula'dan fırlatılan özel kediyi bulup Yeryüzüne çeken fonksiyon
func (h *HomeHandler) LoadCat(w http.ResponseWriter, r *http.Request) {
    setHTMLContentType(w)
    idStr := r.URL.Query().Get("id")
    id, _ := strconv.ParseUint(idStr, 10, 32)

    // Arşivi tarayıp kediyi ve sırasını buluyoruz
    photos, err := h.service.List(r.Context())
    if err != nil || len(photos) == 0 {
        h.Random(w, r)
        return
    }

    var targetPhoto domain.Photo
    rank := 1
    found := false

    for i, p := range photos {
        if p.ID == uint(id) {
            targetPhoto = p
            rank = i + 1
            found = true
            break
        }
    }

    // Eğer kedi silinmişse veya yoksa rastgele ver
    if !found {
        h.Random(w, r)
        return
    }

    // O kusursuz single_cat tasarımını (butonlarla) basıyoruz
    h.tmpl.ExecuteTemplate(w, "single_cat", map[string]interface{}{
        "Photo": targetPhoto,
        "Rank":  rank,
        "Total": len(photos),
    })
}
// --- KURYE (REVERSE PROXY) FONKSİYONU ---
func (h *HomeHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
    // 1. URL'den dosya adını al (/images/kedi.jpg -> kedi.jpg)
    filename := strings.TrimPrefix(r.URL.Path, "/images/")

    // 2. Servisten fotoğrafın veri akışını (stream) iste
    stream, contentType, err := h.service.GetPhotoStream(r.Context(), filename)
    if err != nil {
        http.Error(w, "Fotoğraf bulunamadı", http.StatusNotFound)
        return
    }
    defer stream.Close() // İş bitince akışı kapat

    // 3. Veriyi şelale gibi telefona akıt
    w.Header().Set("Content-Type", contentType)
    io.Copy(w, stream)
}
http.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
    // Google'ın sevdiği standart sitemap şablonu
    sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
   <url>
      <loc>https://sadecekedi.com.tr/</loc>
      <lastmod>2026-04-23</lastmod>
      <changefreq>daily</changefreq>
      <priority>1.0</priority>
   </url>
</urlset>`

    // İçeriğin XML olduğunu tarayıcıya ve Google'a söylüyoruz
    w.Header().Set("Content-Type", "application/xml")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(sitemap))
})