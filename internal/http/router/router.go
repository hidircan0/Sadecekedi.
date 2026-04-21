package router

import (
	"net/http"

	"cat-uploader-go/internal/app"
	"cat-uploader-go/internal/http/handlers"
)

type HandlerSet struct {
	Home *handlers.HomeHandler
}

func NewHandlerSet(service *app.PhotoService) *HandlerSet {
	home, err := handlers.NewHomeHandler(service)
	if err != nil {
		panic(err)
	}
	return &HandlerSet{Home: home}
}

func New(h *HandlerSet) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /{$}", h.Home.Home)
    mux.HandleFunc("POST /upload", h.Home.Upload)
    mux.HandleFunc("GET /zula", h.Home.Zula)
    
    // RASTGELE VE HEDEFLİ KEDİ GETİRME (EKSİK OLAN YOLU AÇTIK)
    mux.HandleFunc("GET /cat/random", h.Home.Random)
    mux.HandleFunc("GET /cat/load", h.Home.LoadCat) // <--- İŞTE SİHRİ YAPACAK SATIR BU
    
    // YÖN KONTROLLERİ VE SİLME
    mux.HandleFunc("GET /cat/next", h.Home.Next)
    mux.HandleFunc("GET /cat/prev", h.Home.Previous)
    
    // Silme işlemini de en stabil haliyle Go 1.22'ye sabitledik
    mux.HandleFunc("DELETE /cat/delete/{id}", h.Home.Delete) 
    
    mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
    // Kurye Rotası (Bütün resim isteklerini buraya yönlendir)
    mux.HandleFunc("GET /images/", h.Home.ServeImage)
    return mux
}
