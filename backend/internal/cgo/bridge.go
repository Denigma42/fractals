package cgo

/*
#cgo CFLAGS: -I./internal/cgo
#cgo LDFLAGS: -L. -lfractals -pthread -lstdc++
#include <stdlib.h>
#include "fractals.h"
*/
import "C"
import (
	"context"
	"unsafe"
)

const (
	FractalMandelbrot  = C.FRACTAL_MANDELBROT
	FractalJulia       = C.FRACTAL_JULIA
	FractalBurningShip = C.FRACTAL_BURNING_SHIP
)

func ComputeTile(width, height int, xMin, xMax, yMin, yMax float64, maxIter int,
	fractalType int, juliaReal, juliaImag float64) ([]uint16, error) {

	bufSize := width * height * int(unsafe.Sizeof(uint16(0)))
	cBuf := C.malloc(C.size_t(bufSize))
	if cBuf == nil {
		return nil, context.Canceled
	}
	defer C.free(cBuf)

	cStopFlag := (*C.int)(C.malloc(C.sizeof_int))
	if cStopFlag == nil {
		return nil, context.Canceled
	}
	defer C.free(unsafe.Pointer(cStopFlag))
	*cStopFlag = 0

	params := C.TileParams{
		width:       C.int(width),
		height:      C.int(height),
		xMin:        C.double(xMin),
		xMax:        C.double(xMax),
		yMin:        C.double(yMin),
		yMax:        C.double(yMax),
		maxIter:     C.int(maxIter),
		fractalType: C.FractalType(fractalType),
		juliaReal:   C.double(juliaReal),
		juliaImag:   C.double(juliaImag),
		output:      (*C.uint16_t)(cBuf),
		stopFlag:    cStopFlag,
	}

	C.ComputeTile(&params)

	if *cStopFlag != 0 {
		return nil, context.Canceled
	}

	data := make([]uint16, width*height)
	cBufSlice := unsafe.Slice((*uint16)(cBuf), width*height)
	copy(data, cBufSlice)

	return data, nil
}