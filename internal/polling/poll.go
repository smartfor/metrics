package polling

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/smartfor/metrics/internal/core"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

type MetricStore map[string]MetricsModel

func PollMainMetrics() MetricStore {
	store := make(MetricStore)

	var ms = runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	store["Alloc"] = MetricsModel{Type: core.Gauge, Key: "Alloc", Value: strconv.FormatUint(ms.Alloc, 10)}
	store["BuckHashSys"] = MetricsModel{Type: core.Gauge, Key: "BuckHashSys", Value: strconv.FormatUint(ms.BuckHashSys, 10)}
	store["Frees"] = MetricsModel{Type: core.Gauge, Key: "Frees", Value: strconv.FormatUint(ms.Frees, 10)}
	store["GCCPUFraction"] = MetricsModel{Type: core.Gauge, Key: "GCCPUFraction", Value: strconv.FormatFloat(ms.GCCPUFraction, 'f', -1, 64)}
	store["GCSys"] = MetricsModel{Type: core.Gauge, Key: "GCSys", Value: strconv.FormatUint(ms.GCSys, 10)}
	store["HeapAlloc"] = MetricsModel{Type: core.Gauge, Key: "HeapAlloc", Value: strconv.FormatUint(ms.HeapAlloc, 10)}
	store["HeapIdle"] = MetricsModel{Type: core.Gauge, Key: "HeapIdle", Value: strconv.FormatUint(ms.HeapIdle, 10)}
	store["HeapInuse"] = MetricsModel{Type: core.Gauge, Key: "HeapInuse", Value: strconv.FormatUint(ms.HeapInuse, 10)}
	store["HeapReleased"] = MetricsModel{Type: core.Gauge, Key: "HeapReleased", Value: strconv.FormatUint(ms.HeapReleased, 10)}
	store["HeapObjects"] = MetricsModel{Type: core.Gauge, Key: "HeapObjects", Value: strconv.FormatUint(ms.HeapObjects, 10)}
	store["HeapSys"] = MetricsModel{Type: core.Gauge, Key: "HeapSys", Value: strconv.FormatUint(ms.HeapSys, 10)}
	store["LastGC"] = MetricsModel{Type: core.Gauge, Key: "LastGC", Value: strconv.FormatUint(ms.LastGC, 10)}
	store["Lookups"] = MetricsModel{Type: core.Gauge, Key: "Lookups", Value: strconv.FormatUint(ms.Lookups, 10)}
	store["MCacheInuse"] = MetricsModel{Type: core.Gauge, Key: "MCacheInuse", Value: strconv.FormatUint(ms.MCacheInuse, 10)}
	store["MCacheSys"] = MetricsModel{Type: core.Gauge, Key: "MCacheSys", Value: strconv.FormatUint(ms.MCacheSys, 10)}
	store["MSpanInuse"] = MetricsModel{Type: core.Gauge, Key: "MSpanInuse", Value: strconv.FormatUint(ms.MSpanInuse, 10)}
	store["MSpanSys"] = MetricsModel{Type: core.Gauge, Key: "MSpanSys", Value: strconv.FormatUint(ms.MSpanSys, 10)}
	store["Mallocs"] = MetricsModel{Type: core.Gauge, Key: "Mallocs", Value: strconv.FormatUint(ms.Mallocs, 10)}
	store["NextGC"] = MetricsModel{Type: core.Gauge, Key: "NextGC", Value: strconv.FormatUint(ms.NextGC, 10)}
	store["NumForcedGC"] = MetricsModel{Type: core.Gauge, Key: "NumForcedGC", Value: strconv.FormatUint(uint64(ms.NumForcedGC), 10)}
	store["NumGC"] = MetricsModel{Type: core.Gauge, Key: "NumGC", Value: strconv.FormatUint(uint64(ms.NumGC), 10)}
	store["OtherSys"] = MetricsModel{Type: core.Gauge, Key: "OtherSys", Value: strconv.FormatUint(ms.OtherSys, 10)}
	store["PauseTotalNs"] = MetricsModel{Type: core.Gauge, Key: "PauseTotalNs", Value: strconv.FormatUint(ms.PauseTotalNs, 10)}
	store["StackInuse"] = MetricsModel{Type: core.Gauge, Key: "StackInuse", Value: strconv.FormatUint(ms.StackInuse, 10)}
	store["StackSys"] = MetricsModel{Type: core.Gauge, Key: "StackSys", Value: strconv.FormatUint(ms.StackSys, 10)}
	store["Sys"] = MetricsModel{Type: core.Gauge, Key: "Sys", Value: strconv.FormatUint(ms.Sys, 10)}
	store["TotalAlloc"] = MetricsModel{Type: core.Gauge, Key: "TotalAlloc", Value: strconv.FormatUint(ms.TotalAlloc, 10)}
	store["RandomValue"] = MetricsModel{Type: core.Gauge, Key: "RandomValue", Value: strconv.FormatFloat(rand.Float64(), 'f', -1, 64)}

	return store
}

func PollAdvancedMetrics() (MetricStore, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// Get CPU stats
	cpus, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}

	store := make(MetricStore)
	store["TotalMemory"] = MetricsModel{Type: core.Gauge, Key: "TotalMemory", Value: strconv.FormatUint(v.Total, 10)}
	store["FreeMemory"] = MetricsModel{Type: core.Gauge, Key: "FreeMemory", Value: strconv.FormatUint(v.Free, 10)}
	for i, c := range cpus {
		store[fmt.Sprintf("CPUutilization%d", i+1)] = MetricsModel{Type: core.Gauge, Key: fmt.Sprintf("CPUUtilization%d", i+1), Value: strconv.FormatFloat(c, 'f', -1, 64)}
	}

	return store, nil
}
