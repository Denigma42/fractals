package transport

import (
	"context"
	"encoding/binary"
	"go-fractals/internal/model"
	"go-fractals/internal/service"
	"net/http"
	"strconv"
	"time"
)

type TileHandler struct {
    service service.TileService
}

func NewTileHandler(s service.TileService) *TileHandler {
    return &TileHandler{service: s}
}

func (h *TileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Парсинг параметров
    fractalType := r.URL.Query().Get("type")
    if fractalType == "" {
        fractalType = "mandelbrot"
    }
    zoomStr := r.URL.Query().Get("z")
    xStr := r.URL.Query().Get("x")
    yStr := r.URL.Query().Get("y")
    if zoomStr == "" || xStr == "" || yStr == "" {
        http.Error(w, "missing parameters", http.StatusBadRequest)
        return
    }
    zoom, _ := strconv.Atoi(zoomStr)
    x, _ := strconv.Atoi(xStr)
    y, _ := strconv.Atoi(yStr)

    // Доп. параметры для Жюлиа
    jrStr := r.URL.Query().Get("jr")
    jiStr := r.URL.Query().Get("ji")
    var jr, ji float64
    if jrStr != "" {
        jr, _ = strconv.ParseFloat(jrStr, 64)
    }
    if jiStr != "" {
        ji, _ = strconv.ParseFloat(jiStr, 64)
    }

    req := model.TileRequest{
        FractalType: fractalType,
        Zoom:        zoom,
        X:           x,
        Y:           y,
        JuliaReal:   jr,
        JuliaImag:   ji,
    }

    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    data, err := h.service.GetTile(ctx, req)
    if err != nil {
        if err == context.Canceled {
            http.Error(w, "cancelled", http.StatusGone)
        } else if err == context.DeadlineExceeded {
            http.Error(w, "timeout", http.StatusRequestTimeout)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/octet-stream")
    // Конвертируем []uint16 в []byte (little-endian)
    buf := make([]byte, len(data)*2)
    for i, v := range data {
        binary.LittleEndian.PutUint16(buf[i*2:], v)
    }
    w.Write(buf)
}