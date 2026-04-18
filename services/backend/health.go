package main

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/disk"
)

const HEART_BEAT = 10 // seconds

type Sample struct {
	CPU        float64
	NetSent    uint64
	NetRecv    uint64
	DiskRead   uint64
	DiskWrite  uint64
}

var (
	samples = make([]Sample, HEART_BEAT)
	index   = 0
	count   = 0
	mu      sync.Mutex

	// previous counters (for rate calculation)
	prevNetSent   uint64
	prevNetRecv   uint64
	prevDiskRead  uint64
	prevDiskWrite uint64
)

func collectMetrics() {
	for {
		cpuPercent, _ := cpu.Percent(0, false)
		netIO, _ := net.IOCounters(false)
		diskIO, _ := disk.IOCounters()

		var netSent, netRecv uint64
		if len(netIO) > 0 {
			netSent = netIO[0].BytesSent
			netRecv = netIO[0].BytesRecv
		}

		var diskRead, diskWrite uint64
		for _, d := range diskIO {
			diskRead += d.ReadBytes
			diskWrite += d.WriteBytes
		}

		// compute per-second delta (rate)
		deltaNetSent := netSent - prevNetSent
		deltaNetRecv := netRecv - prevNetRecv
		deltaDiskRead := diskRead - prevDiskRead
		deltaDiskWrite := diskWrite - prevDiskWrite

		prevNetSent = netSent
		prevNetRecv = netRecv
		prevDiskRead = diskRead
		prevDiskWrite = diskWrite

		mu.Lock()
		samples[index] = Sample{
			CPU:       cpuPercent[0],
			NetSent:   deltaNetSent,
			NetRecv:   deltaNetRecv,
			DiskRead:  deltaDiskRead,
			DiskWrite: deltaDiskWrite,
		}

		index = (index + 1) % HEART_BEAT
		if count < HEART_BEAT {
			count++
		}
		mu.Unlock()

		time.Sleep(1 * time.Second)
	}
}


func getAverages() (cpuAvg float64, netSent uint64, netRecv uint64, diskRead uint64, diskWrite uint64) {
	mu.Lock()
	defer mu.Unlock()

	if count == 0 {
		return
	}

	for i := 0; i < count; i++ {
		cpuAvg += samples[i].CPU
		netSent += samples[i].NetSent
		netRecv += samples[i].NetRecv
		diskRead += samples[i].DiskRead
		diskWrite += samples[i].DiskWrite
	}

	cpuAvg /= float64(count)

	return
}

