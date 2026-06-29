package model

type TileRequest struct {
    FractalType string  `json:"type"`    // "mandelbrot", "julia", "burning_ship"
    Zoom        int     `json:"z"`
    X           int     `json:"x"`
    Y           int     `json:"y"`
    // для Жюлиа:
    JuliaReal   float64 `json:"jr,omitempty"`
    JuliaImag   float64 `json:"ji,omitempty"`
}