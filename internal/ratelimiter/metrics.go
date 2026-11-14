package ratelimiter

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type SystemTelemetry struct {
	Uptime          string     `json:"uptime"`
	CPU             CPUMetrics `json:"cpu"`
	GoProcessMemory GoMemory   `json:"go_process_memory"`
	HostSystemRAM   HostRAM    `json:"host_system_ram"`
}

type CPUMetrics struct {
	LogicalCores     int     `json:"logical_cores"`
	ActiveGoroutines int     `json:"active_goroutines"`
	LoadAvg1Min      float64 `json:"load_avg_1m"`
	LoadAvg5Min      float64 `json:"load_avg_5m"`
	EstimatedUsage   string  `json:"estimated_cpu_usage_percent"`
}

type GoMemory struct {
	AllocMB      float64 `json:"heap_alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_reserved_mb"`
	NumGC        uint32  `json:"gc_cycles"`
}

type HostRAM struct {
	TotalMB     uint64  `json:"total_mb"`
	UsedMB      uint64  `json:"used_mb"`
	AvailableMB uint64  `json:"available_mb"`
	UsedPercent float64 `json:"used_percent"`
}

var startTime = time.Now()

func CollectTelemetry() SystemTelemetry {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	cores := runtime.NumCPU()
	load1, load5 := readLinuxLoadAvg()
	cpuPercent := (load1 / float64(cores)) * 100
	if cpuPercent > 100 {
		cpuPercent = 100
	}

	return SystemTelemetry{
		Uptime: time.Since(startTime).Round(time.Second).String(),
		CPU: CPUMetrics{
			LogicalCores:     cores,
			ActiveGoroutines: runtime.NumGoroutine(),
			LoadAvg1Min:      load1,
			LoadAvg5Min:      load5,
			EstimatedUsage:   fmt.Sprintf("%.2f%%", cpuPercent),
		},
		GoProcessMemory: GoMemory{
			AllocMB:      roundFloat(float64(m.Alloc) / 1024 / 1024),
			TotalAllocMB: roundFloat(float64(m.TotalAlloc) / 1024 / 1024),
			SysMB:        roundFloat(float64(m.Sys) / 1024 / 1024),
			NumGC:        m.NumGC,
		},
		HostSystemRAM: readLinuxRAM(),
	}
}

func readLinuxLoadAvg() (float64, float64) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0, 0
	}
	l1, _ := strconv.ParseFloat(fields[0], 64)
	l5, _ := strconv.ParseFloat(fields[1], 64)
	return l1, l5
}

func readLinuxRAM() HostRAM {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return HostRAM{}
	}

	var totalKB, availKB uint64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				totalKB, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				availKB, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}

	if totalKB == 0 {
		return HostRAM{}
	}

	usedKB := totalKB - availKB
	pct := (float64(usedKB) / float64(totalKB)) * 100

	return HostRAM{
		TotalMB:     totalKB / 1024,
		UsedMB:      usedKB / 1024,
		AvailableMB: availKB / 1024,
		UsedPercent: roundFloat(pct),
	}
}

func roundFloat(val float64) float64 {
	return float64(int(val*100)) / 100
}
