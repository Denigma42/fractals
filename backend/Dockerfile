# ========== 1. Этап сборки ==========
FROM golang:1.26.4 AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    g++ make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /build/internal/fractals
RUN g++ -c -fPIC -O3 -std=c++17 -pthread fractals.cpp -o fractals.o
RUN ar rcs libfractals.a fractals.o

WORKDIR /build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o fractal-server cmd/fractals/main.go

# ========== 2. Финальный образ (Bookworm) ==========
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /build/fractal-server .
EXPOSE 8080
CMD ["./fractal-server"]