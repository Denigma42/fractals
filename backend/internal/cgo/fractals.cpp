#include "fractals.h"
#include <thread>
#include <vector>
#include <complex>
#include <atomic>
#include <functional>
#include <queue>
#include <mutex>
#include <condition_variable>
#include <chrono>
#include <cmath>

class ThreadPool {
public:
    ThreadPool(size_t numThreads) : stop(false) {
        for (size_t i = 0; i < numThreads; ++i) {
            workers.emplace_back([this] { workerLoop(); });
        }
    }
    ~ThreadPool() {
        {
            std::unique_lock<std::mutex> lock(queueMutex);
            stop = true;
        }
        condition.notify_all();
        for (std::thread &worker : workers) {
            worker.join();
        }
    }
    void enqueue(std::function<void()> task) {
        {
            std::unique_lock<std::mutex> lock(queueMutex);
            tasks.push(std::move(task));
        }
        condition.notify_one();
    }
private:
    void workerLoop() {
        while (true) {
            std::function<void()> task;
            {
                std::unique_lock<std::mutex> lock(queueMutex);
                condition.wait(lock, [this] { return stop || !tasks.empty(); });
                if (stop && tasks.empty()) return;
                task = std::move(tasks.front());
                tasks.pop();
            }
            task();
        }
    }
    std::vector<std::thread> workers;
    std::queue<std::function<void()>> tasks;
    std::mutex queueMutex;
    std::condition_variable condition;
    bool stop;
};

static ThreadPool& getPool() {
    static ThreadPool pool(std::thread::hardware_concurrency());
    return pool;
}

class FractalCalculator {
public:
    virtual int iterate(double real, double imag) const = 0;
    virtual ~FractalCalculator() {}
};

class Mandelbrot : public FractalCalculator {
public:
    Mandelbrot(int maxIter) : maxIter_(maxIter) {}
    int iterate(double real, double imag) const override {
        std::complex<double> c(real, imag);
        std::complex<double> z(0, 0);
        int iter = 0;
        while (iter < maxIter_ && std::norm(z) < 4.0) {
            z = z * z + c;
            ++iter;
        }
        return iter;
    }
private:
    int maxIter_;
};

class Julia : public FractalCalculator {
public:
    Julia(int maxIter, double cReal, double cImag) : maxIter_(maxIter), c_(cReal, cImag) {}
    int iterate(double real, double imag) const override {
        std::complex<double> z(real, imag);
        int iter = 0;
        while (iter < maxIter_ && std::norm(z) < 4.0) {
            z = z * z + c_;
            ++iter;
        }
        return iter;
    }
private:
    int maxIter_;
    std::complex<double> c_;
};

class BurningShip : public FractalCalculator {
public:
    BurningShip(int maxIter) : maxIter_(maxIter) {}
    int iterate(double real, double imag) const override {
        std::complex<double> c(real, imag);
        std::complex<double> z(0, 0);
        int iter = 0;
        while (iter < maxIter_ && std::norm(z) < 4.0) {
            double re = std::abs(z.real());
            double im = std::abs(z.imag());
            z = std::complex<double>(re*re - im*im + c.real(), 2*re*im + c.imag());
            ++iter;
        }
        return iter;
    }
private:
    int maxIter_;
};

static FractalCalculator* createCalculator(FractalType type, int maxIter, double jr, double ji) {
    switch (type) {
        case FRACTAL_MANDELBROT:   return new Mandelbrot(maxIter);
        case FRACTAL_JULIA:        return new Julia(maxIter, jr, ji);
        case FRACTAL_BURNING_SHIP: return new BurningShip(maxIter);
        default: return nullptr;
    }
}

extern "C" void ComputeTile(TileParams* params) {
    if (!params || !params->output) return;

    int width = params->width;
    int height = params->height;
    double xMin = params->xMin;
    double xMax = params->xMax;
    double yMin = params->yMin;
    double yMax = params->yMax;
    int maxIter = params->maxIter;
    FractalType type = params->fractalType;
    double jr = params->juliaReal;
    double ji = params->juliaImag;
    volatile int* stopFlag = params->stopFlag;
    uint16_t* output = params->output;

    FractalCalculator* calc = createCalculator(type, maxIter, jr, ji);
    if (!calc) return;

    int numThreads = std::thread::hardware_concurrency();
    if (numThreads == 0) numThreads = 4;
    int rowsPerThread = height / numThreads;
    std::atomic<int> completedRows(0);

    for (int t = 0; t < numThreads; ++t) {
        int yStart = t * rowsPerThread;
        int yEnd = (t == numThreads - 1) ? height : yStart + rowsPerThread;
        // Исправленная лямбда: захват по значению, кроме completedRows по ссылке
        getPool().enqueue([=, &completedRows]() {
            for (int y = yStart; y < yEnd; ++y) {
                if (stopFlag && *stopFlag) break;
                for (int x = 0; x < width; ++x) {
                    if ((x & 15) == 0 && stopFlag && *stopFlag) goto exit_loop;
                    double real = xMin + (xMax - xMin) * x / width;
                    double imag = yMin + (yMax - yMin) * y / height;
                    int iter = calc->iterate(real, imag);
                    output[y * width + x] = static_cast<uint16_t>(iter);
                }
                exit_loop:;
            }
            completedRows.fetch_add(1);
        });
    }

    while (completedRows.load() < height) {
        if (stopFlag && *stopFlag) break;
        std::this_thread::sleep_for(std::chrono::milliseconds(5));
    }

    delete calc;
}