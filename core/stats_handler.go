package core

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Stats struct {
	Time int64 `json:"time"`
	// runtime
	GoVersion    string `json:"go_version"`
	GoOs         string `json:"go_os"`
	GoArch       string `json:"go_arch"`
	CpuNum       int    `json:"cpu_num"`
	GoroutineNum int    `json:"goroutine_num"`
	Gomaxprocs   int    `json:"gomaxprocs"`
	CgoCallNum   int64  `json:"cgo_call_num"`
	// memory
	MemoryAlloc      uint64 `json:"memory_alloc"`
	MemoryTotalAlloc uint64 `json:"memory_total_alloc"`
	MemorySys        uint64 `json:"memory_sys"`
	MemoryLookups    uint64 `json:"memory_lookups"`
	MemoryMallocs    uint64 `json:"memory_mallocs"`
	MemoryFrees      uint64 `json:"memory_frees"`
	// stack
	StackInUse uint64 `json:"memory_stack"`
	// heap
	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapIdle     uint64 `json:"heap_idle"`
	HeapInuse    uint64 `json:"heap_inuse"`
	HeapReleased uint64 `json:"heap_released"`
	HeapObjects  uint64 `json:"heap_objects"`
	// garbage collection
	GcNext           uint64    `json:"gc_next"`
	GcLast           uint64    `json:"gc_last"`
	GcNum            uint32    `json:"gc_num"`
	GcPerSecond      float64   `json:"gc_per_second"`
	GcPausePerSecond float64   `json:"gc_pause_per_second"`
	GcPause          []float64 `json:"gc_pause"`
}

var lastSampleTime = time.Now()
var lastPauseNs uint64 = 0
var lastNumGc uint32 = 0
var nsInMs = float64(time.Millisecond)

var statsMux sync.Mutex

func getStats() *Stats {
	statsMux.Lock()
	defer statsMux.Unlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	now := time.Now()

	var gcPausePerSecond float64

	if lastPauseNs > 0 {
		pauseSinceLastSample := mem.PauseTotalNs - lastPauseNs
		gcPausePerSecond = float64(pauseSinceLastSample) / nsInMs
	}

	lastPauseNs = mem.PauseTotalNs

	countGc := int(mem.NumGC - lastNumGc)

	var gcPerSecond float64

	if lastNumGc > 0 {
		diff := float64(countGc)
		diffTime := now.Sub(lastSampleTime).Seconds()
		gcPerSecond = diff / diffTime
	}

	if countGc > 256 {
		// lagging GC pause times
		countGc = 256
	}

	gcPause := make([]float64, countGc)

	for i := 0; i < countGc; i++ {
		idx := int((mem.NumGC-uint32(i))+255) % 256
		pause := float64(mem.PauseNs[idx])
		gcPause[i] = pause / nsInMs
	}

	lastNumGc = mem.NumGC
	lastSampleTime = time.Now()

	return &Stats{
		Time:         now.UnixNano(),
		GoVersion:    runtime.Version(),
		GoOs:         runtime.GOOS,
		GoArch:       runtime.GOARCH,
		CpuNum:       runtime.NumCPU(),
		GoroutineNum: runtime.NumGoroutine(),
		Gomaxprocs:   runtime.GOMAXPROCS(0),
		CgoCallNum:   runtime.NumCgoCall(),
		// memory
		MemoryAlloc:      mem.Alloc,
		MemoryTotalAlloc: mem.TotalAlloc,
		MemorySys:        mem.Sys,
		MemoryLookups:    mem.Lookups,
		MemoryMallocs:    mem.Mallocs,
		MemoryFrees:      mem.Frees,
		// stack
		StackInUse: mem.StackInuse,
		// heap
		HeapAlloc:    mem.HeapAlloc,
		HeapSys:      mem.HeapSys,
		HeapIdle:     mem.HeapIdle,
		HeapInuse:    mem.HeapInuse,
		HeapReleased: mem.HeapReleased,
		HeapObjects:  mem.HeapObjects,
		// garbage collection
		GcNext:           mem.NextGC,
		GcLast:           mem.LastGC,
		GcNum:            mem.NumGC,
		GcPerSecond:      gcPerSecond,
		GcPausePerSecond: gcPausePerSecond,
		GcPause:          gcPause,
	}
}

func StatsHandler(logger *zap.SugaredLogger) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		stats := connectionPool.Stat()
		fullStats := map[string]interface{} {
			"acquire_count": stats.AcquireCount(),
			"acquired_conns": stats.AcquiredConns(),
			"max_conns": stats.MaxConns(),
			"acquire_duration": stats.AcquireDuration(),
			"idle_conns": stats.IdleConns(),
			"total_conns": stats.TotalConns(),
			"empty_acquire_count": stats.EmptyAcquireCount(),
			"stats": getStats(),
		}

		json.NewEncoder(writer).Encode(fullStats)
		logger.Infow("@", "stats", fullStats)
	}
}
