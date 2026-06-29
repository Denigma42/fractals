package service

import (
    "context"
    "fmt"
    "runtime"
    "sync"
    cgo "go-fractals/internal/fractals"
    "go-fractals/internal/model"
    "go-fractals/internal/repository"
)

const (
    TileSize = 256
    MaxIter  = 1000
    XMin     = -2.5
    XMax     = 1.0
    YMin     = -1.2
    YMax     = 1.2
)

type TileService interface {
    GetTile(ctx context.Context, req model.TileRequest) ([]uint16, error)
    Stop()
}

type tileService struct {
    cache      repository.TileCache
    taskQueue  chan *tileTask
    workerPool sync.WaitGroup
    stopOnce   sync.Once
    stopChan   chan struct{}
}

type tileTask struct {
    req    model.TileRequest
    ctx    context.Context
    result chan tileResult
}

type tileResult struct {
    data []uint16
    err  error
}

func NewTileService(cache repository.TileCache) *tileService {
    s := &tileService{
        cache:     cache,
        taskQueue: make(chan *tileTask, 10000),
        stopChan:  make(chan struct{}),
    }
    numWorkers := runtime.NumCPU()
    for i := 0; i < numWorkers; i++ {
        s.workerPool.Add(1)
        go s.worker()
    }
    return s
}

func (s *tileService) Stop() {
    s.stopOnce.Do(func() {
        close(s.stopChan)
        close(s.taskQueue)
    })
    s.workerPool.Wait()
}

func (s *tileService) GetTile(ctx context.Context, req model.TileRequest) ([]uint16, error) {
    cacheKey := s.cacheKey(req)
    if data, ok := s.cache.Get(cacheKey); ok {
        return data, nil
    }

    task := &tileTask{
        req:    req,
        ctx:    ctx,
        result: make(chan tileResult, 1),
    }

    select {
    case s.taskQueue <- task:
    default:
        return nil, fmt.Errorf("server busy")
    }

    select {
    case res := <-task.result:
        return res.data, res.err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (s *tileService) worker() {
    defer s.workerPool.Done()
    for {
        select {
        case task, ok := <-s.taskQueue:
            if !ok {
                return
            }
            s.processTask(task)
        case <-s.stopChan:
            return
        }
    }
}

func (s *tileService) processTask(task *tileTask) {
    select {
    case <-task.ctx.Done():
        task.result <- tileResult{err: task.ctx.Err()}
        return
    default:
    }

    tiles := 1 << task.req.Zoom
    tileW := (XMax - XMin) / float64(tiles)
    tileH := (YMax - YMin) / float64(tiles)
    xMin := XMin + float64(task.req.X)*tileW
    xMax := XMin + float64(task.req.X+1)*tileW
    yMin := YMin + float64(task.req.Y)*tileH
    yMax := YMin + float64(task.req.Y+1)*tileH

    var fractalType int
    switch task.req.FractalType {
    case "mandelbrot":
        fractalType = cgo.FractalMandelbrot
    case "julia":
        fractalType = cgo.FractalJulia
    case "burning_ship":
        fractalType = cgo.FractalBurningShip
    default:
        task.result <- tileResult{err: fmt.Errorf("unknown fractal type")}
        return
    }

    // Вызов C++ (без stopFlag)
    data, err := cgo.ComputeTile(
        TileSize, TileSize,
        xMin, xMax, yMin, yMax,
        MaxIter,
        fractalType,
        task.req.JuliaReal, task.req.JuliaImag,
    )
    if err != nil {
        task.result <- tileResult{err: err}
        return
    }

    // Проверяем контекст после вычислений (если отменён – не кэшируем)
    select {
    case <-task.ctx.Done():
        task.result <- tileResult{err: task.ctx.Err()}
        return
    default:
    }

    cacheKey := s.cacheKey(task.req)
    s.cache.Add(cacheKey, data)
    task.result <- tileResult{data: data}
}

func (s *tileService) cacheKey(req model.TileRequest) string {
    return fmt.Sprintf("%s/%d/%d/%d", req.FractalType, req.Zoom, req.X, req.Y)
}