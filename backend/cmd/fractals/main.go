package main

import (
    "log"
    "net/http"

    "go-fractals/internal/repository"
    "go-fractals/internal/service"
    "go-fractals/internal/transport"
)

func main() {
    // Инициализация кэша
    cache, err := repository.NewLRUCache(2000)
    if err != nil {
        log.Fatal(err)
    }

    // Сервис
    svc := service.NewTileService(cache)
    defer svc.Stop()

    // HTTP обработчик
    handler := transport.NewTileHandler(svc)

    http.Handle("/tile", handler)

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}