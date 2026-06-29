package main

import (
    "log"
    "net/http"

	"github.com/rs/cors"
    "go-fractals/internal/repository"
    "go-fractals/internal/service"
    "go-fractals/internal/transport"
)

func main() {
	mux := http.NewServeMux()

    // Инициализация кэша
    cache, err := repository.NewLRUCache(2000)
    if err != nil {
        log.Fatal(err)
    }

    // Сервис
    svc := service.NewTileService(cache)
    defer svc.Stop()

    // HTTP обработчик
    tileHandler := transport.NewTileHandler(svc)

    mux.Handle("/tile", tileHandler)

	handler := cors.Default().Handler(mux)
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}