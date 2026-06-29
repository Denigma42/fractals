#pragma once
#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

typedef enum {
    FRACTAL_MANDELBROT = 0,
    FRACTAL_JULIA = 1,
    FRACTAL_BURNING_SHIP = 2
} FractalType;

typedef struct {
    int width, height;
    double xMin, xMax;
    double yMin, yMax;
    int maxIter;
    FractalType fractalType;    // переименовано с type
    double juliaReal;
    double juliaImag;
    uint16_t* output;
    volatile int* stopFlag;
} TileParams;

void ComputeTile(TileParams* params);

#ifdef __cplusplus
}
#endif