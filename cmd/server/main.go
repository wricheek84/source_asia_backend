package main

import (
	"log"
	"net/http"

	"github.com/wricheek84/source_asia_backend/internal/handler"
	"github.com/wricheek84/source_asia_backend/internal/middleware"
	"github.com/wricheek84/source_asia_backend/internal/store"
)

func main() {
	
	rlStore := store.NewRateLimitStore()
	pStore := store.NewProductStore()

	
	rlHandler := handler.NewRateLimitHandler(rlStore)
	pHandler := handler.NewProductHandler(pStore)

	
	mux := http.NewServeMux()

	mux.HandleFunc("/request", rlHandler.HandleRequest)
	mux.HandleFunc("/stats", rlHandler.HandleStats)

	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			pHandler.CreateProduct(w, r)
		} else {
			pHandler.ListProducts(w, r)
		}
	})
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		pHandler.HandleProductDetailOrMedia(w, r)
	})

	wrappedServer := middleware.RateLimitMiddleware(rlStore)(mux)

	
	port := ":8080"
	log.Printf("Server is starting up cleanly on port %s...", port)
	if err := http.ListenAndServe(port, wrappedServer); err != nil {
		log.Fatalf("Server failed to boot: %v", err)
	}
}